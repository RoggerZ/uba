package main

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestPrintMainUsageContainsProtectAndDiagnosticSections(t *testing.T) {
	var buffer bytes.Buffer

	originalOutput := flag.CommandLine.Output()
	flag.CommandLine.SetOutput(&buffer)
	defer flag.CommandLine.SetOutput(originalOutput)

	printMainUsage()

	usageText := buffer.String()
	for _, want := range []string{
		"sinker diagnostic --help",
		"sinker protect --help",
		"sinker diagnostic enable",
		"sinker diagnostic -json",
		"sinker protect status",
		"sinker protect set",
	} {
		if !strings.Contains(usageText, want) {
			t.Fatalf("main usage should contain %q, got: %s", want, usageText)
		}
	}

	for _, want := range []string{
		"sinker diagnostic enable --duration 3m",
		`sinker diagnostic -json "{\"action\":\"enable\"`,
		"sinker protect status --admin-addr",
		"sinker protect set --observe-only=true",
	} {
		if count := strings.Count(usageText, want); count != 1 {
			t.Fatalf("main usage should contain %q exactly once, got %d in: %s", want, count, usageText)
		}
	}
}
