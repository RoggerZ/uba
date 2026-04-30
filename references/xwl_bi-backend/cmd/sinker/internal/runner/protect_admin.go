package runner

import (
	"encoding/json"
	"net/http"
)

func (s *sinkerAdminServer) registerProtectRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/admin/protect/status", s.handleProtectStatus)
	mux.HandleFunc("/admin/protect/enable", s.handleProtectEnable)
	mux.HandleFunc("/admin/protect/disable", s.handleProtectDisable)
	mux.HandleFunc("/admin/protect/set", s.handleProtectSet)
}

func (s *sinkerAdminServer) handleProtectStatus(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.protectControl == nil {
		writeAdminError(writer, http.StatusServiceUnavailable, "protect controller is unavailable")
		return
	}
	writeAdminJSON(writer, http.StatusOK, s.protectControl.Status())
}

func (s *sinkerAdminServer) handleProtectEnable(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.protectControl == nil {
		writeAdminError(writer, http.StatusServiceUnavailable, "protect controller is unavailable")
		return
	}
	s.protectControl.Enable()
	writeAdminJSON(writer, http.StatusOK, s.protectControl.Status())
}

func (s *sinkerAdminServer) handleProtectDisable(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.protectControl == nil {
		writeAdminError(writer, http.StatusServiceUnavailable, "protect controller is unavailable")
		return
	}
	s.protectControl.Disable()
	writeAdminJSON(writer, http.StatusOK, s.protectControl.Status())
}

func (s *sinkerAdminServer) handleProtectSet(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writeAdminError(writer, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !s.authorize(request) {
		writeAdminError(writer, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.protectControl == nil {
		writeAdminError(writer, http.StatusServiceUnavailable, "protect controller is unavailable")
		return
	}

	var payload protectSetRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeAdminError(writer, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := s.protectControl.Set(payload); err != nil {
		writeAdminError(writer, http.StatusBadRequest, err.Error())
		return
	}
	writeAdminJSON(writer, http.StatusOK, s.protectControl.Status())
}
