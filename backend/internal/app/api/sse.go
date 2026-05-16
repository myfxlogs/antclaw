package api

import "net/http"

// registerSSE registers all SSE endpoints on the mux.
// The actual handler implementations live in cmd/antclaw-api/sse_handlers.go.
func registerSSE(mux *http.ServeMux,
	jobsHandler, auditHandler, macroHandler, optionsHandler, signalsHandler, notificationsHandler http.HandlerFunc,
) {
	mux.HandleFunc("/sse/jobs", jobsHandler)
	mux.HandleFunc("/sse/audit", auditHandler)
	mux.HandleFunc("/sse/macro_alerts", macroHandler)
	mux.HandleFunc("/sse/options_alerts", optionsHandler)
	mux.HandleFunc("/sse/signals_alerts", signalsHandler)
	mux.HandleFunc("/sse/notifications", notificationsHandler)
}
