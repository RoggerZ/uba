package runner

import (
	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

func logTraceStage(decoded DecodedMessage, stage string, phase string, extraFields ...zap.Field) {
	if !util.ShouldTraceSinkerOffset(decoded.Input.Offset) {
		return
	}

	session := util.CurrentSinkerDiagnosticSession()
	fields := append(decodedLogFields(decoded),
		zap.Int32("generation_id", decoded.Input.GenerationID),
		zap.Int64("trace_target_offset", session.TraceOffset),
		zap.Time("trace_expires_at", session.ExpiresAt),
		zap.String("trace_source", session.Source),
		zap.String("trace_stage", stage),
		zap.String("trace_phase", phase),
	)
	fields = append(fields, extraFields...)

	logs.Logger.Info("target offset trace", fields...)
}
