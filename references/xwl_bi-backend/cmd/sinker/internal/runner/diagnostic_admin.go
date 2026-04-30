package runner

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/1340691923/xwl_bi/engine/logs"
	"github.com/1340691923/xwl_bi/model"
	"github.com/1340691923/xwl_bi/platform-basic-libs/util"
	"go.uber.org/zap"
)

type sinkerAdminServer struct {
	config         model.SinkerConfig
	listener       net.Listener
	server         *http.Server
	protectControl *reportConsumptionProtector
}

// diagnosticEnableRequest 是 `/admin/diagnostic/enable` 的请求体。
//
// 这里故意把持续时长交给服务端再做一次裁剪：
// 1. 空值回退到默认值
// 2. 超过上限时自动截断
//
// 这样即使 CLI 侧没限制好，服务端也不会把诊断日志无限放大。
type diagnosticEnableRequest struct {
	DurationSeconds                   int    `json:"durationSeconds"`
	TraceOffset                       *int64 `json:"traceOffset"`
	ReportHandlerStageTimingThreshold string `json:"reportHandlerStageTimingThreshold"`
	RateLogInterval                   string `json:"rateLogInterval"`
}

type diagnosticStatusResponse struct {
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

func newSinkerAdminServer(config model.SinkerConfig) (*sinkerAdminServer, error) {
	config = config.Normalize()
	if !isLoopbackHost(config.AdminHttpHost) && strings.TrimSpace(config.AdminToken) == "" {
		return nil, errors.New("sinker admin token is required when admin http host is not loopback")
	}

	mux := http.NewServeMux()
	adminServer := &sinkerAdminServer{
		config: config,
		server: &http.Server{
			Handler: mux,
		},
	}
	mux.HandleFunc("/admin/diagnostic/enable", adminServer.handleEnableDiagnostic)
	mux.HandleFunc("/admin/diagnostic/disable", adminServer.handleDisableDiagnostic)
	mux.HandleFunc("/admin/diagnostic/status", adminServer.handleDiagnosticStatus)
	adminServer.registerProtectRoutes(mux)

	address := net.JoinHostPort(config.AdminHttpHost, strconv.Itoa(int(config.AdminHttpPort)))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	adminServer.listener = listener
	return adminServer, nil
}

// Start 以独立 goroutine 方式启动 admin HTTP 服务。
//
// 这里和 pprof 刻意分开：
// 1. pprof 是性能观测入口
// 2. admin 是运行态控制入口
//
// 避免因为是否开启 pprof，顺手影响到诊断控制能力。
func (s *sinkerAdminServer) Start() {
	if s == nil || s.server == nil || s.listener == nil {
		return
	}

	go func() {
		logs.Logger.Info("sinker admin server started",
			zap.String("host", s.config.AdminHttpHost),
			zap.Uint16("port", s.config.AdminHttpPort),
		)
		if err := s.server.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logs.Logger.Error("sinker admin server stopped unexpectedly", zap.Error(err))
		}
	}()
}

func (s *sinkerAdminServer) Stop() error {
	if s == nil || s.server == nil {
		return nil
	}
	return s.server.Close()
}

func (s *sinkerAdminServer) SetProtectionController(controller *reportConsumptionProtector) {
	if s == nil {
		return
	}
	s.protectControl = controller
}

func (s *sinkerAdminServer) handleEnableDiagnostic(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	var payload diagnosticEnableRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeAdminError(writer, http.StatusBadRequest, "invalid request body")
		return
	}
	if payload.TraceOffset != nil && *payload.TraceOffset < 0 {
		writeAdminError(writer, http.StatusBadRequest, "traceOffset must be >= 0")
		return
	}

	var reportHandlerStageTimingThreshold *time.Duration
	if strings.TrimSpace(payload.ReportHandlerStageTimingThreshold) != "" {
		parsedThreshold, err := time.ParseDuration(strings.TrimSpace(payload.ReportHandlerStageTimingThreshold))
		if err != nil {
			writeAdminError(writer, http.StatusBadRequest, "reportHandlerStageTimingThreshold must be a valid Go duration")
			return
		}
		if parsedThreshold <= 0 {
			writeAdminError(writer, http.StatusBadRequest, "reportHandlerStageTimingThreshold must be > 0")
			return
		}
		reportHandlerStageTimingThreshold = &parsedThreshold
	}

	var rateLogInterval *time.Duration
	if strings.TrimSpace(payload.RateLogInterval) != "" {
		parsedInterval, err := time.ParseDuration(strings.TrimSpace(payload.RateLogInterval))
		if err != nil {
			writeAdminError(writer, http.StatusBadRequest, "rateLogInterval must be a valid Go duration")
			return
		}
		if parsedInterval <= 0 {
			writeAdminError(writer, http.StatusBadRequest, "rateLogInterval must be > 0")
			return
		}
		rateLogInterval = &parsedInterval
	}

	durationSeconds := payload.DurationSeconds
	if durationSeconds <= 0 {
		durationSeconds = s.config.DiagnosticDefaultTTLSeconds
	}
	if durationSeconds > s.config.DiagnosticMaxTTLSeconds {
		durationSeconds = s.config.DiagnosticMaxTTLSeconds
	}

	session := util.EnableSinkerDiagnosticSession(
		time.Duration(durationSeconds)*time.Second,
		payload.TraceOffset,
		reportHandlerStageTimingThreshold,
		rateLogInterval,
		"admin_http",
		time.Now(),
	)
	writeAdminJSON(writer, http.StatusOK, buildDiagnosticStatusResponse(session, time.Now()))
}

func (s *sinkerAdminServer) handleDisableDiagnostic(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	session := util.DisableSinkerDiagnosticSession("admin_http", time.Now())
	writeAdminJSON(writer, http.StatusOK, buildDiagnosticStatusResponse(session, time.Now()))
}

func (s *sinkerAdminServer) handleDiagnosticStatus(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}

	writeAdminJSON(writer, http.StatusOK, buildDiagnosticStatusResponse(util.CurrentSinkerDiagnosticSession(), time.Now()))
}

func (s *sinkerAdminServer) authorize(request *http.Request) bool {
	if isLoopbackRemoteAddr(request.RemoteAddr) {
		return true
	}
	if s.config.AdminToken == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(request.Header.Get("X-Admin-Token")), []byte(s.config.AdminToken)) == 1
}

func writeAdminJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

func writeAdminError(writer http.ResponseWriter, statusCode int, message string) {
	writeAdminJSON(writer, statusCode, map[string]string{"error": message})
}

func buildDiagnosticStatusResponse(session util.SinkerDiagnosticSession, now time.Time) diagnosticStatusResponse {
	response := diagnosticStatusResponse{
		Enabled:                                  session.Enabled,
		TraceOffsetEnabled:                       session.TraceOffsetEnabled,
		TraceOffset:                              session.TraceOffset,
		ReportHandlerStageTimingThreshold:        effectiveReportHandlerStageTimingSlowThreshold(session).String(),
		DefaultReportHandlerStageTimingThreshold: reportHandlerStageTimingSlowThresholdDefault.String(),
		RateLogInterval:                          session.RateLogInterval.String(),
		Source:                                   session.Source,
	}
	if !session.ExpiresAt.IsZero() {
		expiresAt := session.ExpiresAt
		response.ExpiresAt = &expiresAt
		if expiresAt.After(now) {
			response.RemainingSeconds = int(expiresAt.Sub(now).Seconds())
		}
	}
	return response
}

func isLoopbackHost(host string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(host))
	return parsedIP != nil && parsedIP.IsLoopback()
}

func isLoopbackRemoteAddr(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	parsedIP := net.ParseIP(strings.TrimSpace(host))
	return parsedIP != nil && parsedIP.IsLoopback()
}
