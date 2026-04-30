package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestParseDiagnosticCLICommandExplicitMode(t *testing.T) {
	command, err := parseDiagnosticCLICommand([]string{
		"enable",
		"--duration", "3m",
		"--trace-offset", "123",
		"--report-handler-threshold", "5s",
		"--admin-addr", "http://10.0.0.1:8094",
		"--admin-token", "secret",
	})
	if err != nil {
		t.Fatalf("parseDiagnosticCLICommand returned error: %v", err)
	}
	if command.Action != "enable" || command.Duration != "3m" {
		t.Fatalf("unexpected command: %+v", command)
	}
	if command.TraceOffset == nil || *command.TraceOffset != 123 {
		t.Fatalf("unexpected trace offset: %+v", command.TraceOffset)
	}
	if command.ReportHandlerStageTimingThreshold != "5s" {
		t.Fatalf("unexpected report handler threshold: %+v", command)
	}
	if command.AdminAddr != "http://10.0.0.1:8094" || command.AdminToken != "secret" {
		t.Fatalf("unexpected admin target: %+v", command)
	}
	if command.Output != "text" {
		t.Fatalf("default output = %q, want text", command.Output)
	}
}

func TestParseDiagnosticCLICommandJSONMode(t *testing.T) {
	command, err := parseDiagnosticCLICommand([]string{
		"-json", `{"action":"status","adminAddr":"http://10.0.0.1:8094","adminToken":"secret"}`,
	})
	if err != nil {
		t.Fatalf("parseDiagnosticCLICommand returned error: %v", err)
	}
	if command.Action != "status" || command.AdminAddr != "http://10.0.0.1:8094" || command.AdminToken != "secret" {
		t.Fatalf("unexpected command: %+v", command)
	}
	if command.Output != "text" {
		t.Fatalf("json mode default output = %q, want text", command.Output)
	}
}

func TestParseDiagnosticCLICommandJSONModeWithReportHandlerThreshold(t *testing.T) {
	command, err := parseDiagnosticCLICommand([]string{
		"-json", `{"action":"enable","reportHandlerStageTimingThreshold":"1500ms"}`,
	})
	if err != nil {
		t.Fatalf("parseDiagnosticCLICommand returned error: %v", err)
	}
	if command.ReportHandlerStageTimingThreshold != "1500ms" {
		t.Fatalf("unexpected report handler threshold: %+v", command)
	}
}

func TestParseDiagnosticCLICommandRejectsMixedJSONAndFlags(t *testing.T) {
	_, err := parseDiagnosticCLICommand([]string{
		"enable",
		"-json", `{"action":"enable"}`,
	})
	if err == nil {
		t.Fatal("expected mixed json and explicit flags to fail")
	}
}

func TestDiagnosticHelpTextContainsKeyUsage(t *testing.T) {
	helpText := diagnosticHelpText()
	if !strings.Contains(helpText, "sinker diagnostic enable") {
		t.Fatalf("help text should contain diagnostic enable usage, got: %s", helpText)
	}
	if !strings.Contains(helpText, "sinker diagnostic -json") {
		t.Fatalf("help text should contain diagnostic json usage, got: %s", helpText)
	}
	if !strings.Contains(helpText, "--output") {
		t.Fatalf("help text should contain output usage, got: %s", helpText)
	}
	if !strings.Contains(helpText, "--report-handler-threshold") {
		t.Fatalf("help text should contain report handler threshold usage, got: %s", helpText)
	}
}

func TestPrintMainUsageContainsDiagnosticSection(t *testing.T) {
	var buffer bytes.Buffer
	originalOutput := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(&buffer)
	defer flag.CommandLine.SetOutput(originalOutput)

	printMainUsage()

	usageText := buffer.String()
	if !strings.Contains(usageText, "sinker diagnostic --help") {
		t.Fatalf("main usage should contain diagnostic summary entry, got: %s", usageText)
	}
	if !strings.Contains(usageText, "--report-handler-threshold") {
		t.Fatalf("main usage should contain report handler threshold usage, got: %s", usageText)
	}
	if !strings.Contains(usageText, "服务模式参数") {
		t.Fatalf("main usage should contain service flag section, got: %s", usageText)
	}
}

func TestFormatDiagnosticCLIResponseForStatus(t *testing.T) {
	expiresAt := "2026-04-15T12:03:00+08:00"
	output := formatDiagnosticCLIResponse("status", "text", []byte(`{"enabled":true,"traceOffsetEnabled":true,"traceOffset":123,"reportHandlerStageTimingThreshold":"2s","defaultReportHandlerStageTimingThreshold":"2s","expiresAt":"`+expiresAt+`","remainingSeconds":180,"source":"admin_http"}`))

	if !strings.Contains(output, "诊断状态: 已开启") {
		t.Fatalf("formatted output should contain status summary, got: %s", output)
	}
	if !strings.Contains(output, "Trace Offset: 123") {
		t.Fatalf("formatted output should contain trace offset summary, got: %s", output)
	}
	if !strings.Contains(output, "Report Handler 阈值: 2s") {
		t.Fatalf("formatted output should contain report handler threshold summary, got: %s", output)
	}
	if !strings.Contains(output, "\"remainingSeconds\": 180") {
		t.Fatalf("formatted output should contain pretty json, got: %s", output)
	}
}

func TestFormatDiagnosticCLIResponseForEnable(t *testing.T) {
	output := formatDiagnosticCLIResponse("enable", "text", []byte(`{"enabled":true,"traceOffsetEnabled":false,"traceOffset":0,"reportHandlerStageTimingThreshold":"5s","defaultReportHandlerStageTimingThreshold":"2s","remainingSeconds":180,"source":"admin_http"}`))

	if !strings.Contains(output, "启用结果: 已开启") {
		t.Fatalf("formatted output should contain enable summary, got: %s", output)
	}
	if !strings.Contains(output, "Report Handler 阈值: 5s") {
		t.Fatalf("formatted output should contain report handler threshold summary, got: %s", output)
	}
	if !strings.Contains(output, "\"enabled\": true") {
		t.Fatalf("formatted output should contain pretty json, got: %s", output)
	}
}

func TestFormatDiagnosticCLIResponseForDisable(t *testing.T) {
	output := formatDiagnosticCLIResponse("disable", "text", []byte(`{"enabled":false,"traceOffsetEnabled":false,"traceOffset":0,"reportHandlerStageTimingThreshold":"2s","defaultReportHandlerStageTimingThreshold":"2s","remainingSeconds":0,"source":"admin_http"}`))

	if !strings.Contains(output, "关闭结果: 已关闭") {
		t.Fatalf("formatted output should contain disable summary, got: %s", output)
	}
	if !strings.Contains(output, "Report Handler 阈值: 2s") {
		t.Fatalf("formatted output should contain report handler threshold summary, got: %s", output)
	}
	if !strings.Contains(output, "\"enabled\": false") {
		t.Fatalf("formatted output should contain pretty json, got: %s", output)
	}
}

func TestFormatDiagnosticCLIResponseForJSONOutput(t *testing.T) {
	raw := `{"enabled":true,"traceOffsetEnabled":false,"traceOffset":0,"reportHandlerStageTimingThreshold":"2s","defaultReportHandlerStageTimingThreshold":"2s","remainingSeconds":180,"source":"admin_http"}`
	output := formatDiagnosticCLIResponse("status", "json", []byte(raw))
	if output != raw {
		t.Fatalf("json output = %q, want raw json %q", output, raw)
	}
}

func TestParseDiagnosticCLICommandRejectsReportHandlerThresholdForDisable(t *testing.T) {
	_, err := parseDiagnosticCLICommand([]string{
		"disable",
		"--report-handler-threshold", "5s",
	})
	if err == nil {
		t.Fatal("expected disable to reject report-handler-threshold")
	}
}

func TestParseDiagnosticCLICommandRejectsReportHandlerThresholdForStatusJSON(t *testing.T) {
	_, err := parseDiagnosticCLICommand([]string{
		"-json", `{"action":"status","reportHandlerStageTimingThreshold":"5s"}`,
	})
	if err == nil {
		t.Fatal("expected status json mode to reject reportHandlerStageTimingThreshold")
	}
}
