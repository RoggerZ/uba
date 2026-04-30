package runner

import (
	"os"
	"sync"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

const reportConsumerDirectExecEnv = "SINKER_REPORT_DIRECT_EXEC"

var reportConsumerExecMode reportConsumerExecModeConfig

type reportConsumerExecModeConfig struct {
	once   sync.Once
	direct bool
}

func (c *reportConsumerExecModeConfig) load() {
	c.once.Do(func() {
		rawValue := os.Getenv(reportConsumerDirectExecEnv)
		if rawValue != "" {
			c.direct = util.IsTruthyEnvValue(rawValue)
			return
		}

		c.direct = model.GlobConfig.Sinker.ReportConsumerDirectExec
	})
}

func isReportConsumerDirectExec() bool {
	reportConsumerExecMode.load()
	return reportConsumerExecMode.direct
}
