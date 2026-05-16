package api

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	redisv9 "github.com/redis/go-redis/v9"
	"github.com/antclaw/antclaw/internal/service/presence"
)

// App is the composed API server with all infrastructure, services, and HTTP handler.
type App struct {
	inf     *Infra
	svc     *Services
	mux     *http.ServeMux
	handler http.Handler
}

// New creates a new App with all components initialized.
func New(cfg Config) (*App, error) {
	boot := time.Now()

	inf, err := InitInfra(cfg)
	if err != nil {
		return nil, err
	}

	svc := InitServices(inf)
	mux := http.NewServeMux()
	registerHandlers(mux, inf, svc, boot)

	return &App{
		inf: inf,
		svc: svc,
		mux: mux,
	}, nil
}

// RegisterSSE wires SSE handlers into the app's mux. Must be called before Handler().
func (a *App) RegisterSSE(h SSEHandlers) {
	registerSSE(a.mux,
		h.Jobs, h.Audit,
		h.MacroAlerts, h.OptionsAlerts,
		h.SignalsAlerts, h.Notifications,
	)
}

// corsMiddleware adds CORS headers, echoing the request Origin (required for credentialed requests).
func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Connect-Protocol-Version")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// Handler returns the composed HTTP handler ready for http.Server.
func (a *App) Handler() http.Handler {
	if a.handler == nil {
		a.handler = h2c.NewHandler(corsMiddleware(a.mux), &http2.Server{})
	}
	return a.handler
}

// RDB returns the underlying Redis client for SSE handler wiring.
func (a *App) RDB() *redisv9.Client { return a.inf.RDB.Raw() }

// Presence returns the presence tracker for SSE handler wiring.
func (a *App) Presence() *presence.Tracker { return a.svc.Presence }

// Close shuts down infrastructure resources gracefully.
func (a *App) Close() {
	if a.inf != nil {
		a.inf.Close()
	}
	log.Printf("API server shut down")
}

// SSEHandlers holds the external SSE handler functions (defined in cmd/antclaw-api).
type SSEHandlers struct {
	Jobs, Audit, MacroAlerts, OptionsAlerts, SignalsAlerts, Notifications http.HandlerFunc
}
