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
	"strconv"
	"strings"
	"time"
)

type protectCLICommand struct {
	Action                    string `json:"action"`
	ObserveOnly               *bool  `json:"observeOnly,omitempty"`
	SampleInterval            string `json:"sampleInterval,omitempty"`
	NormalRateLogInterval     string `json:"normalRateLogInterval,omitempty"`
	DiagnosticRateLogInterval string `json:"diagnosticRateLogInterval,omitempty"`
	SoftTargetRatePerSecond   *int   `json:"softTargetRatePerSecond,omitempty"`
	AdminAddr                 string `json:"adminAddr,omitempty"`
	AdminToken                string `json:"adminToken,omitempty"`
	Output                    string `json:"output,omitempty"`
}

type protectCLIConsumerRateSnapshot struct {
	Group          string  `json:"group"`
	Topic          string  `json:"topic"`
	CurrentOffset  int64   `json:"currentOffset"`
	LogEndOffset   int64   `json:"logEndOffset"`
	Lag            int64   `json:"lag"`
	DeltaLag       int64   `json:"deltaLag"`
	SpeedPerSecond float64 `json:"speedPerSecond"`
	SampledAt      string  `json:"sampledAt"`
	SamplingMode   string  `json:"samplingMode"`
}

type protectCLIGateSnapshot struct {
	InFlightMessages  int64 `json:"inFlightMessages"`
	WaitingTasks      int64 `json:"waitingTasks"`
	CompletedMessages int64 `json:"completedMessages"`
}

type protectCLIPipelineSnapshot struct {
	CommitterCount        int                    `json:"committerCount"`
	PendingCount          int                    `json:"pendingCount"`
	DoneCount             int                    `json:"doneCount"`
	LargestPendingGap     int64                  `json:"largestPendingGap"`
	OldestPendingOffset   int64                  `json:"oldestPendingOffset"`
	NewestCompletedOffset int64                  `json:"newestCompletedOffset"`
	Gate                  protectCLIGateSnapshot `json:"gate"`
}

type protectCLIWorkerPoolStats struct {
	Queued          int     `json:"Queued"`
	QueueCapacity   int     `json:"QueueCapacity"`
	QueueUsageRatio float64 `json:"QueueUsageRatio"`
	Running         int     `json:"Running"`
	Capacity        int     `json:"Capacity"`
	Idle            int     `json:"Idle"`
	BusyRatio       float64 `json:"BusyRatio"`
	MinWorkers      int     `json:"MinWorkers"`
	MaxWorkers      int     `json:"MaxWorkers"`
	SubmittedTotal  int64   `json:"SubmittedTotal"`
	CompletedTotal  int64   `json:"CompletedTotal"`
	RejectedTotal   int64   `json:"RejectedTotal"`
	Closed          bool    `json:"Closed"`
}

type protectCLIPersistenceErrors struct {
	CountLastMinute int    `json:"countLastMinute"`
	LastClass       string `json:"lastClass"`
	LastError       string `json:"lastError"`
	LastOccurredAt  string `json:"lastOccurredAt"`
}

type protectCLIDBHealthState struct {
	Name                string `json:"Name"`
	DriverName          string `json:"DriverName"`
	Enabled             bool   `json:"Enabled"`
	Status              string `json:"Status"`
	ConsecutiveFailures int    `json:"ConsecutiveFailures"`
	LastError           string `json:"LastError"`
	LastErrorAt         string `json:"LastErrorAt"`
	LastRecoveredAt     string `json:"LastRecoveredAt"`
	LastUpdatedAt       string `json:"LastUpdatedAt"`
}

type protectCLIStatusResponse struct {
	Enabled                 bool                           `json:"enabled"`
	ObserveOnly             bool                           `json:"observeOnly"`
	State                   string                         `json:"state"`
	SoftHoldUntil           string                         `json:"softHoldUntil"`
	HardHoldUntil           string                         `json:"hardHoldUntil"`
	LastTransitionAt        string                         `json:"lastTransitionAt"`
	LastTransitionReason    string                         `json:"lastTransitionReason"`
	CurrentSoftSignals      []string                       `json:"currentSoftSignals"`
	CurrentHardSignals      []string                       `json:"currentHardSignals"`
	CurrentRecoveryBlockers []string                       `json:"currentRecoveryBlockers"`
	ReportRate              protectCLIConsumerRateSnapshot `json:"reportRate"`
	RealTimeRate            protectCLIConsumerRateSnapshot `json:"realTimeRate"`
	ReportPipeline          protectCLIPipelineSnapshot     `json:"reportPipeline"`
	ReportConsumerPool      protectCLIWorkerPoolStats      `json:"reportConsumerPool"`
	ReportPersistPool       protectCLIWorkerPoolStats      `json:"reportPersistPool"`
	PersistenceErrors       protectCLIPersistenceErrors    `json:"persistenceErrors"`
	MySQLHealth             protectCLIDBHealthState        `json:"mysqlHealth"`
	Goroutines              int                            `json:"goroutines"`
	HeapAllocBytes          uint64                         `json:"heapAllocBytes"`
}

func maybeRunProtectCLI(args []string) (bool, int) {
	if len(args) == 0 || args[0] != "protect" {
		return false, 0
	}
	if len(args) == 1 || isHelpArg(args[1]) {
		fmt.Print(protectHelpText())
		return true, 0
	}

	exitCode, err := runProtectCLI(args[1:])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return true, 1
	}
	return true, exitCode
}

func runProtectCLI(args []string) (int, error) {
	command, err := parseProtectCLICommand(args)
	if err != nil {
		return 1, err
	}

	responseBody, err := executeProtectCLICommand(command)
	if err != nil {
		return 1, err
	}
	fmt.Println(responseBody)
	return 0, nil
}

func parseProtectCLICommand(args []string) (protectCLICommand, error) {
	var command protectCLICommand

	flagSet := flag.NewFlagSet("protect", flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	var (
		adminAddr                 string
		adminToken                string
		outputMode                string
		observeOnlyRaw            bool
		sampleIntervalRaw         string
		normalRateLogIntervalRaw  string
		diagnosticRateIntervalRaw string
		softTargetRateRaw         int
	)

	flagSet.StringVar(&adminAddr, "admin-addr", defaultAdminAddr, "admin http address")
	flagSet.StringVar(&adminToken, "admin-token", "", "admin token")
	flagSet.StringVar(&outputMode, "output", "text", "protect output mode: text or json")
	flagSet.BoolVar(&observeOnlyRaw, "observe-only", false, "observe only")
	flagSet.StringVar(&sampleIntervalRaw, "sample-interval", "", "sample interval")
	flagSet.StringVar(&normalRateLogIntervalRaw, "normal-rate-log-interval", "", "normal rate log interval")
	flagSet.StringVar(&diagnosticRateIntervalRaw, "diagnostic-rate-log-interval", "", "diagnostic rate log interval")
	flagSet.IntVar(&softTargetRateRaw, "soft-target-rate", 0, "soft limited target rate per second")

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

	command.Action = action
	command.AdminAddr = adminAddr
	command.AdminToken = adminToken
	command.Output = strings.TrimSpace(strings.ToLower(outputMode))
	if command.Output == "" {
		command.Output = "text"
	}
	if command.Output != "text" && command.Output != "json" {
		return command, errors.New("output 必须是 text 或 json")
	}

	if command.Action != "enable" && command.Action != "disable" && command.Action != "status" && command.Action != "set" {
		return command, errors.New("protect action 必须是 enable、disable、status 或 set")
	}
	if command.Action != "set" && len(visited) > 0 {
		for key := range visited {
			if key == "admin-addr" || key == "admin-token" || key == "output" {
				continue
			}
			return command, errors.New("只有 set 动作允许指定保护参数")
		}
	}
	if visited["observe-only"] {
		command.ObserveOnly = &observeOnlyRaw
	}
	command.SampleInterval = strings.TrimSpace(sampleIntervalRaw)
	command.NormalRateLogInterval = strings.TrimSpace(normalRateLogIntervalRaw)
	command.DiagnosticRateLogInterval = strings.TrimSpace(diagnosticRateIntervalRaw)
	if visited["soft-target-rate"] {
		command.SoftTargetRatePerSecond = &softTargetRateRaw
	}
	return command, nil
}

func executeProtectCLICommand(command protectCLICommand) (string, error) {
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
		url = adminAddr + "/admin/protect/enable"
		body = bytes.NewReader([]byte("{}"))
	case "disable":
		method = http.MethodPost
		url = adminAddr + "/admin/protect/disable"
		body = bytes.NewReader([]byte("{}"))
	case "status":
		method = http.MethodGet
		url = adminAddr + "/admin/protect/status"
	case "set":
		method = http.MethodPost
		url = adminAddr + "/admin/protect/set"
		payloadBytes, err := json.Marshal(command)
		if err != nil {
			return "", err
		}
		body = bytes.NewReader(payloadBytes)
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
		return "", fmt.Errorf("protect request failed: %s %s", response.Status, strings.TrimSpace(string(responseBytes)))
	}
	if command.Output == "json" {
		return strings.TrimSpace(string(responseBytes)), nil
	}
	return formatProtectCLIOutput(command.Action, responseBytes), nil
}

func formatProtectCLIOutput(action string, responseBytes []byte) string {
	raw := strings.TrimSpace(string(responseBytes))
	if raw == "" {
		return raw
	}

	var payload map[string]any
	if err := json.Unmarshal(responseBytes, &payload); err != nil {
		return raw
	}

	reportRate := asMap(payload["reportRate"])
	realTimeRate := asMap(payload["realTimeRate"])
	reportPipeline := asMap(payload["reportPipeline"])
	reportGate := asMap(reportPipeline["gate"])
	reportConsumerPool := asMap(payload["reportConsumerPool"])
	reportPersistPool := asMap(payload["reportPersistPool"])
	persistenceErrors := asMap(payload["persistenceErrors"])
	mysqlHealth := asMap(payload["mysqlHealth"])

	var builder strings.Builder
	if action != "status" {
		builder.WriteString("保护配置已生效\n")
	}
	builder.WriteString("保护开关: ")
	if asBool(payload["enabled"]) {
		builder.WriteString("已开启\n")
	} else {
		builder.WriteString("已关闭\n")
	}
	builder.WriteString("执行模式: ")
	if asBool(payload["observeOnly"]) {
		builder.WriteString("仅观测\n")
	} else {
		builder.WriteString("观测+执行\n")
	}
	builder.WriteString("当前状态: ")
	builder.WriteString(asString(payload["state"]))
	builder.WriteString("\n")
	builder.WriteString("软驻留截止: ")
	builder.WriteString(formatProtectTime(asString(payload["softHoldUntil"])))
	builder.WriteString("\n")
	builder.WriteString("硬驻留截止: ")
	builder.WriteString(formatProtectTime(asString(payload["hardHoldUntil"])))
	builder.WriteString("\n")
	builder.WriteString("最近切换: ")
	builder.WriteString(formatProtectTime(asString(payload["lastTransitionAt"])))
	builder.WriteString(" reason=")
	builder.WriteString(emptyFallback(asString(payload["lastTransitionReason"]), "无"))
	builder.WriteString("\n")
	builder.WriteString("当前 soft signals: ")
	builder.WriteString(formatProtectStringList(payload["currentSoftSignals"]))
	builder.WriteString("\n")
	builder.WriteString("当前 hard signals: ")
	builder.WriteString(formatProtectStringList(payload["currentHardSignals"]))
	builder.WriteString("\n")
	builder.WriteString("当前恢复阻塞: ")
	builder.WriteString(formatProtectStringList(payload["currentRecoveryBlockers"]))
	builder.WriteString("\n")
	builder.WriteString("进程资源: goroutines=")
	builder.WriteString(strconv.FormatInt(asInt64(payload["goroutines"]), 10))
	builder.WriteString(" heap_alloc=")
	builder.WriteString(formatBytes(uint64(asInt64(payload["heapAllocBytes"]))))
	builder.WriteString("\n")
	builder.WriteString(formatProtectRateBlockFromMap("ReportData2CK", reportRate))
	builder.WriteString(formatProtectRateBlockFromMap("RealTime", realTimeRate))
	builder.WriteString("消费链路: committer_count=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["committerCount"]), 10))
	builder.WriteString(" pending_count=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["pendingCount"]), 10))
	builder.WriteString(" done_count=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["doneCount"]), 10))
	builder.WriteString(" largest_pending_gap=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["largestPendingGap"]), 10))
	builder.WriteString(" oldest_pending_offset=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["oldestPendingOffset"]), 10))
	builder.WriteString(" newest_completed_offset=")
	builder.WriteString(strconv.FormatInt(asInt64(reportPipeline["newestCompletedOffset"]), 10))
	builder.WriteString("\n")
	builder.WriteString("Gate: in_flight_messages=")
	builder.WriteString(strconv.FormatInt(asInt64(reportGate["inFlightMessages"]), 10))
	builder.WriteString(" waiting_tasks=")
	builder.WriteString(strconv.FormatInt(asInt64(reportGate["waitingTasks"]), 10))
	builder.WriteString(" completed_messages=")
	builder.WriteString(strconv.FormatInt(asInt64(reportGate["completedMessages"]), 10))
	builder.WriteString("\n")
	builder.WriteString(formatProtectPoolBlockFromMap("消费池", reportConsumerPool))
	builder.WriteString(formatProtectPoolBlockFromMap("持久化池", reportPersistPool))
	builder.WriteString("持久化错误(近1m): ")
	builder.WriteString(strconv.FormatInt(asInt64(persistenceErrors["countLastMinute"]), 10))
	builder.WriteString(" last_class=")
	builder.WriteString(emptyFallback(asString(persistenceErrors["lastClass"]), "无"))
	builder.WriteString(" last_time=")
	builder.WriteString(formatProtectTime(asString(persistenceErrors["lastOccurredAt"])))
	builder.WriteString("\n")
	builder.WriteString("持久化最后错误: ")
	builder.WriteString(emptyFallback(asString(persistenceErrors["lastError"]), "无"))
	builder.WriteString("\n")
	builder.WriteString("MySQL健康: status=")
	builder.WriteString(asString(mysqlHealth["Status"]))
	builder.WriteString(" failures=")
	builder.WriteString(strconv.FormatInt(asInt64(mysqlHealth["ConsecutiveFailures"]), 10))
	builder.WriteString(" last_recovered_at=")
	builder.WriteString(formatProtectTime(asString(mysqlHealth["LastRecoveredAt"])))
	if strings.TrimSpace(asString(mysqlHealth["LastError"])) != "" {
		builder.WriteString("\n")
		builder.WriteString("MySQL最后错误: ")
		builder.WriteString(asString(mysqlHealth["LastError"]))
	}
	return strings.TrimSpace(builder.String())
}

func formatProtectRateBlockFromMap(title string, snapshot map[string]any) string {
	var builder strings.Builder
	builder.WriteString(title)
	builder.WriteString(": group=")
	builder.WriteString(asString(snapshot["group"]))
	builder.WriteString(" topic=")
	builder.WriteString(asString(snapshot["topic"]))
	builder.WriteString(" current_offset=")
	builder.WriteString(strconv.FormatInt(asInt64(snapshot["currentOffset"]), 10))
	builder.WriteString(" log_end_offset=")
	builder.WriteString(strconv.FormatInt(asInt64(snapshot["logEndOffset"]), 10))
	builder.WriteString(" lag=")
	builder.WriteString(strconv.FormatInt(asInt64(snapshot["lag"]), 10))
	builder.WriteString(" delta_lag=")
	builder.WriteString(strconv.FormatInt(asInt64(snapshot["deltaLag"]), 10))
	builder.WriteString(" speed_per_sec=")
	builder.WriteString(strconv.FormatFloat(asFloat64(snapshot["speedPerSecond"]), 'f', 2, 64))
	builder.WriteString(" sampled_at=")
	builder.WriteString(formatProtectTime(asString(snapshot["sampledAt"])))
	builder.WriteString(" mode=")
	builder.WriteString(asString(snapshot["samplingMode"]))
	builder.WriteString("\n")
	return builder.String()
}

func formatProtectPoolBlockFromMap(title string, stats map[string]any) string {
	var builder strings.Builder
	builder.WriteString(title)
	builder.WriteString(": running=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["Running"]), 10))
	builder.WriteString("/")
	builder.WriteString(strconv.FormatInt(asInt64(stats["Capacity"]), 10))
	builder.WriteString(" queued=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["Queued"]), 10))
	builder.WriteString("/")
	builder.WriteString(strconv.FormatInt(asInt64(stats["QueueCapacity"]), 10))
	builder.WriteString(" busy_ratio=")
	builder.WriteString(strconv.FormatFloat(asFloat64(stats["BusyRatio"]), 'f', 3, 64))
	builder.WriteString(" queue_usage_ratio=")
	builder.WriteString(strconv.FormatFloat(asFloat64(stats["QueueUsageRatio"]), 'f', 3, 64))
	builder.WriteString(" bounds=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["MinWorkers"]), 10))
	builder.WriteString("~")
	builder.WriteString(strconv.FormatInt(asInt64(stats["MaxWorkers"]), 10))
	builder.WriteString(" submitted=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["SubmittedTotal"]), 10))
	builder.WriteString(" completed=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["CompletedTotal"]), 10))
	builder.WriteString(" rejected=")
	builder.WriteString(strconv.FormatInt(asInt64(stats["RejectedTotal"]), 10))
	builder.WriteString(" closed=")
	builder.WriteString(strconv.FormatBool(asBool(stats["Closed"])))
	builder.WriteString("\n")
	return builder.String()
}

func formatProtectTime(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "0001-01-01T00:00:00Z" {
		return "未设置"
	}

	parsed, err := time.Parse(time.RFC3339Nano, trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.Format("2006-01-02 15:04:05 -07:00")
}

func formatBytes(size uint64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}

	value := float64(size)
	suffixes := []string{"KiB", "MiB", "GiB", "TiB"}
	for _, suffix := range suffixes {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.2f%s", value, suffix)
		}
	}
	return fmt.Sprintf("%.2fPiB", value/unit)
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func asMap(value any) map[string]any {
	if result, ok := value.(map[string]any); ok {
		return result
	}
	return map[string]any{}
}

func asString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", value)
	}
}

func asInt64(value any) int64 {
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int64:
		return typed
	case int32:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	default:
		return 0
	}
}

func asFloat64(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		parsed, _ := typed.Float64()
		return parsed
	default:
		return 0
	}
}

func asBool(value any) bool {
	if typed, ok := value.(bool); ok {
		return typed
	}
	return false
}

func formatProtectStringList(value any) string {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return "无"
	}

	values := make([]string, 0, len(items))
	for _, item := range items {
		values = append(values, asString(item))
	}
	return strings.Join(values, ",")
}

func protectHelpText() string {
	return strings.TrimSpace(`
protect 模式用法:

1. 显式参数模式
   sinker protect status --admin-addr http://127.0.0.1:8094 --admin-token secret --output json
   sinker protect enable --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
   sinker protect disable --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
   sinker protect set --observe-only=true --sample-interval 5s --normal-rate-log-interval 3m --diagnostic-rate-log-interval 5s --soft-target-rate 1000 --admin-addr http://127.0.0.1:8094 --admin-token secret --output text
`) + "\n"
}
