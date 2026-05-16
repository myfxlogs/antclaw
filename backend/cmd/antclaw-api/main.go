// Package main implements the AntClaw API server entry point.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antclaw/antclaw/internal/app/api"
)

func main() {
	cfg := api.DefaultConfig()

	app, err := api.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize API server: %v", err)
	}
	defer app.Close()

	// Wire SSE handlers using the app's Redis connection.
	rdb := app.RDB()
	sseH := api.SSEHandlers{
		Jobs:            jobsEventsHandler(rdb),
		Audit:           auditEventsHandler(rdb),
		MacroAlerts:     alertsEventsHandler(rdb, "stream:macro_alerts"),
		OptionsAlerts:   alertsEventsHandler(rdb, "stream:options_alerts"),
		SignalsAlerts:   alertsEventsHandler(rdb, "stream:signals_alerts"),
		Notifications:   userNotificationsSSE(rdb, app.Presence()),
	}
	app.RegisterSSE(sseH)

	server := &http.Server{
		Addr:    cfg.Addr,
		Handler: app.Handler(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}
