package http

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/liftel/epic-fain/internal/domain/port"
)

// Router sets up all HTTP API routes.
type Router struct {
	telemetrySvc port.TelemetryService
	controlSvc   port.DeviceControlService
	alertSvc     port.AlertService
	installSvc   port.InstallationService
	auth         *APIKeyAuth
}

func NewRouter(
	telemetrySvc port.TelemetryService,
	controlSvc port.DeviceControlService,
	alertSvc port.AlertService,
	installSvc port.InstallationService,
	auth *APIKeyAuth,
) *Router {
	return &Router{
		telemetrySvc: telemetrySvc,
		controlSvc:   controlSvc,
		alertSvc:     alertSvc,
		installSvc:   installSvc,
		auth:         auth,
	}
}

// Handler returns the configured HTTP handler with all routes.
func (rt *Router) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health check (no auth)
	mux.HandleFunc("GET /health", rt.healthCheck)

	// Protected routes
	protected := http.NewServeMux()

	// Telemetry endpoints (Hito 2.5 / 3.3)
	protected.HandleFunc("GET /api/v1/telemetry/{installationID}", rt.getTelemetry)
	protected.HandleFunc("GET /api/v1/telemetry/{installationID}/unconsumed", rt.getUnconsumed)
	protected.HandleFunc("POST /api/v1/telemetry/{id}/ack", rt.ackConsumed)
	protected.HandleFunc("DELETE /api/v1/telemetry/consumed", rt.deleteConsumed)

	// Device control endpoints (Hito 3.4)
	protected.HandleFunc("POST /api/v1/control/{installationID}/restart", rt.restartDevice)
	protected.HandleFunc("POST /api/v1/control/{installationID}/vvvf/enable", rt.enableVVVF)
	protected.HandleFunc("POST /api/v1/control/{installationID}/vvvf/disable", rt.disableVVVF)
	protected.HandleFunc("POST /api/v1/control/{installationID}/superfast", rt.enableSuperfast)

	// Alert endpoints (Hito 3.1)
	protected.HandleFunc("GET /api/v1/alerts/{installationID}", rt.getAlerts)
	protected.HandleFunc("POST /api/v1/alerts/{id}/ack", rt.ackAlert)

	// Installation endpoints (Hito 2.2)
	protected.HandleFunc("GET /api/v1/installations", rt.listInstallations)
	protected.HandleFunc("GET /api/v1/installations/{installationID}", rt.getInstallation)
	protected.HandleFunc("POST /api/v1/installations", rt.createInstallation)

	mux.Handle("/", rt.auth.Middleware(protected))

	return mux
}

// --- Health ---

func (rt *Router) healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "epic-fain"})
}

// --- Telemetry ---

func (rt *Router) getTelemetry(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	from, to := parseTimeRange(r)

	records, err := rt.telemetrySvc.GetTelemetry(r.Context(), instID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (rt *Router) getUnconsumed(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	records, err := rt.telemetrySvc.GetUnconsumed(r.Context(), instID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, records)
}

func (rt *Router) ackConsumed(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	actor := ActorFromContext(r.Context())

	if err := rt.telemetrySvc.AckConsumed(r.Context(), id, actor); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

func (rt *Router) deleteConsumed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actor := ActorFromContext(r.Context())
	if err := rt.telemetrySvc.DeleteConsumed(r.Context(), req.IDs, actor); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Device Control ---

func (rt *Router) restartDevice(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	actor := ActorFromContext(r.Context())

	if err := rt.controlSvc.ResetInverter(r.Context(), instID, actor); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "restart_sent"})
}

func (rt *Router) enableVVVF(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	actor := ActorFromContext(r.Context())

	if err := rt.controlSvc.EnableVVVF(r.Context(), instID, actor); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "vvvf_enabled"})
}

func (rt *Router) disableVVVF(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	actor := ActorFromContext(r.Context())

	if err := rt.controlSvc.DisableVVVF(r.Context(), instID, actor); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "vvvf_disabled"})
}

func (rt *Router) enableSuperfast(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	actor := ActorFromContext(r.Context())

	var req struct {
		DurationSeconds int `json:"duration_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DurationSeconds <= 0 {
		writeError(w, http.StatusBadRequest, "duration_seconds must be a positive integer")
		return
	}

	duration := time.Duration(req.DurationSeconds) * time.Second
	if err := rt.controlSvc.EnableSuperfast(r.Context(), instID, duration, actor); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "superfast_enabled"})
}

// --- Alerts ---

func (rt *Router) getAlerts(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	from, to := parseTimeRange(r)

	alerts, err := rt.alertSvc.GetAlerts(r.Context(), instID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, alerts)
}

func (rt *Router) ackAlert(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := rt.alertSvc.AcknowledgeAlert(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

// --- Installations ---

func (rt *Router) listInstallations(w http.ResponseWriter, r *http.Request) {
	installations, err := rt.installSvc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, installations)
}

func (rt *Router) getInstallation(w http.ResponseWriter, r *http.Request) {
	instID := r.PathValue("installationID")
	inst, err := rt.installSvc.Get(r.Context(), instID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, inst)
}

func (rt *Router) createInstallation(w http.ResponseWriter, r *http.Request) {
	// TODO: implement with full validation
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[HTTP] Error encoding JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseTimeRange(r *http.Request) (time.Time, time.Time) {
	from := time.Now().Add(-24 * time.Hour)
	to := time.Now()

	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = t
		}
	}
	return from, to
}
