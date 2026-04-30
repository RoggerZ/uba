package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultAdminAddr = "http://127.0.0.1:8094"

// diagnosticCLICommand 描述一次 CLI 诊断请求的完整输入。
//
// 它同时承载两种入口：
// 1. `sinker diagnostic enable --duration ...`
// 2. `sinker diagnostic -json '{...}'`
//
// 这样命令行层只负责把不同输入形式统一折叠成一个结构，
// 后续真正发 HTTP 请求时就不需要分两套分支继续处理。
type diagnosticCLICommand struct {
	Action                            string `json:"action"`
	Duration                          string `json:"duration,omitempty"`
	TraceOffset                       *int64 `json:"traceOffset,omitempty"`
	ReportHandlerStageTimingThreshold string `json:"reportHandlerStageTimingThreshold,omitempty"`
	RateLogInterval                   string `json:"rateLogInterval,omitempty"`
	AdminAddr                         string `json:"adminAddr,omitempty"`
	AdminToken                        string `json:"adminToken,omitempty"`
	Output                            string `json:"output,omitempty"`
}

type diagnosticEnablePayload struct {
	DurationSeconds                   int    `json:"durationSeconds"`
	TraceOffset                       *int64 `json:"traceOffset,omitempty"`
	ReportHandlerStageTimingThreshold string `json:"reportHandlerStageTimingThreshold,omitempty"`
	RateLogInterval                   string `json:"rateLogInterval,omitempty"`
}

type diagnosticStatusCLIResponse struct {
	Enabled                                  bool       `json:"enabled"`
	TraceOffsetEnabled                       bool       `json:"traceOffsetEnabled"`
	TraceOffset                              int64      `json:"traceOffset"`
	ReportHandlerStageTimingThreshold        string     `json:"reportHandlerStageTimingThreshold"`
	DefaultReportHandlerStageTimingThreshold string     `json:"defaultReportHandlerStageTimingThreshold"`
	RateLogInterval                          string     `json:"rateLogInterval"`
	ExpiresAt                                *time.Time `json:"expiresAt,omitempty"`
	RemainingSeconds                         int        `json:"remainingSeconds"`
	Source                                   string     `json:"source"`
}

// maybeRunDiagnosticCLI 用来区分当前这次进程启动到底是：
// 1. 正常 sinker 服务模式
// 2. 诊断控制模式
//
// 只要命令前缀是 `sinker diagnostic ...`，
// 当前进程就只作为 HTTP client 发控制指令，随后立即退出。
func maybeRunDiagnosticCLI(args []string) (bool, int) {
	if len(args) == 0 || args[0] != "diagnostic" {
		return false, 0
	}
	if len(args) == 1 || isHelpArg(args[1]) {
		fmt.Print(diagnosticHelpText())
		return true, 0
	}

	exitCode, err := runDiagnosticCLI(args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return true, 1
	}
	return true, exitCode
}

func runDiagnosticCLI(args []string) (int, error) {
	command, err := parseDiagnosticCLICommand(args)
	if err != nil {
		return 1, err
	}

	responseBody, err := executeDiagnosticCLICommand(command)
	if err != nil {
		return 1, err
	}
	fmt.Println(responseBody)
	return 0, nil
}

// parseDiagnosticCLICommand 把显式 flags 模式和 `-json` 模式统一解析成一个命令结构。
//
// 这里刻意规定两种模式互斥，避免出现：
// 1. JSON 里写了一套动作
// 2. flags 又覆盖另一套动作
// 3. 最终到底以谁为准变得不清楚
func parseDiagnosticCLICommand(args []string) (diagnosticCLICommand, error) {
	var command diagnosticCLICommand

	flagSet := flag.NewFlagSet("diagnostic", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var (
		jsonPayload               string
		durationRaw               string
		traceOffsetRaw            int64
		reportHandlerThresholdRaw string
		rateLogIntervalRaw        string
		adminAddr                 string
		adminToken                string
		outputMode                string
	)

	flagSet.StringVar(&jsonPayload, "json", "", "diagnostic json payload")
	flagSet.StringVar(&durationRaw, "duration", "", "diagnostic duration, such as 3m")
	flagSet.Int64Var(&traceOffsetRaw, "trace-offset", 0, "trace offset")
	flagSet.StringVar(&reportHandlerThresholdRaw, "report-handler-threshold", "", "report handler slow timing threshold, such as 2s")
	flagSet.StringVar(&rateLogIntervalRaw, "rate-log-interval", "", "rate log interval, such as 5s")
	flagSet.StringVar(&adminAddr, "admin-addr", defaultAdminAddr, "admin http address")
	flagSet.StringVar(&adminToken, "admin-token", "", "admin token")
	flagSet.StringVar(&outputMode, "output", "text", "diagnostic output mode: text or json")

	action := ""
	restArgs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		action = strings.TrimSpace(strings.ToLower(args[0]))
		restArgs = args[1:]
	}

	if err := flagSet.Parse(restArgs); err != nil {
		return command, err
	}

	visited := map[string]bool{}
	flagSet.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if jsonPayload != "" {
		if action != "" || len(visited) > 1 {
			return command, errors.New("diagnostic -json 模式不能和显式 action/flags 混用")
		}
		if err := json.Unmarshal([]byte(jsonPayload), &command); err != nil {
			return command, fmt.Errorf("diagnostic json 解析失败: %w", err)
		}
	} else {
		command.Action = action
		command.Duration = durationRaw
		if visited["trace-offset"] {
			traceOffset := traceOffsetRaw
			command.TraceOffset = &traceOffset
		}
		command.ReportHandlerStageTimingThreshold = reportHandlerThresholdRaw
		command.RateLogInterval = rateLogIntervalRaw
		command.AdminAddr = adminAddr
		command.AdminToken = adminToken
		command.Output = outputMode
	}

	command.Action = strings.TrimSpace(strings.ToLower(command.Action))
	if command.Action != "enable" && command.Action != "disable" && command.Action != "status" {
		return command, errors.New("diagnostic action 必须是 enable、disable 或 status")
	}
	if strings.TrimSpace(command.AdminAddr) == "" {
		command.AdminAddr = defaultAdminAddr
	}
	if command.TraceOffset != nil && *command.TraceOffset < 0 {
		return command, errors.New("trace-offset 必须大于等于 0")
	}
	if command.Action != "enable" && command.Duration != "" {
		return command, errors.New("只有 enable 动作允许指定 duration")
	}
	if command.Action != "enable" && command.TraceOffset != nil {
		return command, errors.New("只有 enable 动作允许指定 trace-offset")
	}
	command.ReportHandlerStageTimingThreshold = strings.TrimSpace(command.ReportHandlerStageTimingThreshold)
	if command.Action != "enable" && command.ReportHandlerStageTimingThreshold != "" {
		return command, errors.New("只有 enable 动作允许指定 report-handler-threshold")
	}
	command.RateLogInterval = strings.TrimSpace(command.RateLogInterval)
	if command.Action != "enable" && command.RateLogInterval != "" {
		return command, errors.New("只有 enable 动作允许指定 rate-log-interval")
	}
	command.Output = strings.TrimSpace(strings.ToLower(command.Output))
	if command.Output == "" {
		command.Output = "text"
	}
	if command.Output != "text" && command.Output != "json" {
		return command, errors.New("output 必须是 text 或 json")
	}

	return command, nil
}

func executeDiagnosticCLICommand(command diagnosticCLICommand) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	adminAddr := strings.TrimRight(command.AdminAddr, "/")

	var (
		method string
		url    string
		body   io.Reader
	)

	switch command.Action {
	case "enable":
		method = http.MethodPost
		url = adminAddr + "/admin/diagnostic/enable"
		payload := diagnosticEnablePayload{
			TraceOffset:                       command.TraceOffset,
			ReportHandlerStageTimingThreshold: command.ReportHandlerStageTimingThreshold,
			RateLogInterval:                   command.RateLogInterval,
		}
		if command.Duration != "" {
			duration, err := time.ParseDuration(command.Duration)
			if err != nil {
				return "", fmt.Errorf("duration 解析失败: %w", err)
			}
			payload.DurationSeconds = int(duration.Seconds())
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		body = bytes.NewReader(payloadBytes)
	case "disable":
		method = http.MethodPost
		url = adminAddr + "/admin/diagnostic/disable"
		body = bytes.NewReader([]byte("{}"))
	case "status":
		method = http.MethodGet
		url = adminAddr + "/admin/diagnostic/status"
	default:
		return "", errors.New("unsupported diagnostic action")
	}

	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return "", err
	}
	if method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(command.AdminToken) != "" {
		request.Header.Set("X-Admin-Token", command.AdminToken)
	}

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("diagnostic request failed: %s %s", response.Status, strings.TrimSpace(string(responseBytes)))
	}
	return formatDiagnosticCLIResponse(command.Action, command.Output, responseBytes), nil
}

// formatDiagnosticCLIResponse 把 admin HTTP 的原始 JSON 响应整理成更适合终端阅读的文本。
//
// 当 output=text 时：
// 1. 先输出摘要
// 2. 再输出 pretty JSON
//
// 当 output=json 时：
// 1. 直接返回原始 JSON 文本
// 2. 不再附加摘要，方便脚本处理
func formatDiagnosticCLIResponse(action string, output string, responseBytes []byte) string {
	trimmed := strings.TrimSpace(string(responseBytes))
	if output == "json" {
		return trimmed
	}

	var payload diagnosticStatusCLIResponse
	if err := json.Unmarshal(responseBytes, &payload); err != nil {
		return trimmed
	}

	prettyJSONBytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return trimmed
	}

	statusLabel := "已关闭"
	if payload.Enabled {
		statusLabel = "已开启"
	}

	traceLabel := "未启用"
	if payload.TraceOffsetEnabled {
		traceLabel = fmt.Sprintf("%d", payload.TraceOffset)
	}

	reportHandlerThresholdLabel := strings.TrimSpace(payload.ReportHandlerStageTimingThreshold)
	if reportHandlerThresholdLabel == "" {
		reportHandlerThresholdLabel = strings.TrimSpace(payload.DefaultReportHandlerStageTimingThreshold)
	}
	if reportHandlerThresholdLabel == "" {
		reportHandlerThresholdLabel = "2s"
	}

	expiresAtLabel := "未设置"
	if payload.ExpiresAt != nil {
		expiresAtLabel = payload.ExpiresAt.Format(time.RFC3339)
	}

	actionLabel := strings.TrimSpace(strings.ToLower(action))
	actionSummary := "诊断状态"
	switch actionLabel {
	case "enable":
		actionSummary = "启用结果"
	case "disable":
		actionSummary = "关闭结果"
	}

	return strings.TrimSpace(fmt.Sprintf(
		`%s: %s
Trace Offset: %s
Report Handler 阈值: %s
剩余时长(秒): %d
过期时间: %s
来源: %s

JSON:
%s`,
		actionSummary,
		statusLabel,
		traceLabel,
		reportHandlerThresholdLabel,
		payload.RemainingSeconds,
		expiresAtLabel,
		payload.Source,
		string(prettyJSONBytes),
	))
}

func isHelpArg(arg string) bool {
	switch strings.TrimSpace(strings.ToLower(arg)) {
	case "-h", "--help", "help":
		return true
	default:
		return false
	}
}

func diagnosticHelpText() string {
	return strings.TrimSpace(`
diagnostic 模式用法:

1. 显式参数模式
   sinker diagnostic enable --duration 3m --trace-offset 123 --report-handler-threshold 5s --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
   sinker diagnostic disable --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
   sinker diagnostic status --admin-addr http://127.0.0.1:8094 --admin-token secret --output json

2. JSON 模式
   sinker diagnostic -json "{\"action\":\"enable\",\"duration\":\"3m\",\"traceOffset\":123,\"reportHandlerStageTimingThreshold\":\"5s\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"text\"}"
   sinker diagnostic -json "{\"action\":\"disable\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"text\"}"
   sinker diagnostic -json "{\"action\":\"status\",\"adminAddr\":\"http://127.0.0.1:8094\",\"adminToken\":\"secret\",\"output\":\"json\"}"

说明:
- diagnostic 命令不会启动 sinker 服务，只会向已运行的 sinker admin HTTP 发送控制请求后退出
- --duration 仅 enable 可用
- --trace-offset 仅 enable 可用
- --report-handler-threshold 仅 enable 可用，使用 Go duration 语法，例如 500ms、2s、5s
- --output 默认是 text，也可以指定为 json
- -json 与显式 action/flags 互斥
- 远程访问 admin HTTP 时需要通过 X-Admin-Token 传入 admin token
`) + "\n"
}
