package runner

import (
	"testing"

	"github.com/1340691923/xwl_bi/model"
)

func TestIsReportConsumerDirectExec(t *testing.T) {
	testCases := []struct {
		name     string
		rawValue string
		want     bool
	}{
		{name: "default disabled", rawValue: "", want: false},
		{name: "one", rawValue: "1", want: true},
		{name: "true", rawValue: "true", want: true},
		{name: "yes", rawValue: "yes", want: true},
		{name: "on", rawValue: "on", want: true},
		{name: "false", rawValue: "false", want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv(reportConsumerDirectExecEnv, testCase.rawValue)
			reportConsumerExecMode = reportConsumerExecModeConfig{}

			if got := isReportConsumerDirectExec(); got != testCase.want {
				t.Fatalf("isReportConsumerDirectExec() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestIsReportConsumerDirectExecFallsBackToConfig(t *testing.T) {
	originalConfig := model.GlobConfig
	defer func() {
		model.GlobConfig = originalConfig
	}()

	t.Setenv(reportConsumerDirectExecEnv, "")
	model.GlobConfig.Sinker.ReportConsumerDirectExec = true
	reportConsumerExecMode = reportConsumerExecModeConfig{}

	if !isReportConsumerDirectExec() {
		t.Fatal("expected direct exec to follow sinker config when env is empty")
	}
}
