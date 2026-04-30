package runner

import (
	"runtime"
	"time"

	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
)

func buildReportConsumerPoolConfig(config model.DynamicWorkerPoolConfigJSON) util.DynamicWorkerPoolConfig {
	return util.DynamicWorkerPoolConfig{
		Name:         "sinker-report-consumer",
		MinWorkers:   pickPoolMin(config.MinWorkers, runtime.GOMAXPROCS(0)),
		MaxWorkers:   pickPoolMax(config.MaxWorkers, minInt(64, runtime.GOMAXPROCS(0)*4)),
		QueueSize:    pickPoolValue(config.QueueSize, 4096),
		TuneInterval: time.Duration(pickPoolValue(config.TuneInterval, 2)) * time.Second,
		DrainTimeout: time.Duration(pickPoolValue(config.DrainTimeout, 30)) * time.Second,
	}
}

func buildReportPersistPoolConfig(config model.DynamicWorkerPoolConfigJSON) util.DynamicWorkerPoolConfig {
	defaultMin := maxInt(1, runtime.GOMAXPROCS(0)/2)
	defaultMax := minInt(16, runtime.GOMAXPROCS(0)*2)

	return util.DynamicWorkerPoolConfig{
		Name:         "sinker-report-persist",
		MinWorkers:   pickPoolMin(config.MinWorkers, defaultMin),
		MaxWorkers:   pickPoolMax(config.MaxWorkers, defaultMax),
		QueueSize:    pickPoolValue(config.QueueSize, 1024),
		TuneInterval: time.Duration(pickPoolValue(config.TuneInterval, 2)) * time.Second,
		DrainTimeout: time.Duration(pickPoolValue(config.DrainTimeout, 30)) * time.Second,
	}
}

func pickPoolMin(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func pickPoolMax(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func pickPoolValue(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
