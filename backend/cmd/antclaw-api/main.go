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

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	antclawv1connect "github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/adapter/rpc"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/auth"
	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	infrapq "github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/service/admin"
	"github.com/antclaw/antclaw/internal/service/ai"
	"github.com/antclaw/antclaw/internal/service/alerts"
	"github.com/antclaw/antclaw/internal/service/audit"
	"github.com/antclaw/antclaw/internal/service/backtest"
	"github.com/antclaw/antclaw/internal/service/calendar"
	"github.com/antclaw/antclaw/internal/service/cot"
	"github.com/antclaw/antclaw/internal/service/datasource"
	"github.com/antclaw/antclaw/internal/service/macro"
	"github.com/antclaw/antclaw/internal/service/price"
	reportsvc "github.com/antclaw/antclaw/internal/service/report"
	"github.com/antclaw/antclaw/internal/service/sentiment"
	signalssvc "github.com/antclaw/antclaw/internal/service/signals"
	"github.com/antclaw/antclaw/internal/service/regime"
	strategysvc "github.com/antclaw/antclaw/internal/service/strategy"
	systemaisvc "github.com/antclaw/antclaw/internal/service/systemai"
	"github.com/antclaw/antclaw/internal/service/ta"
	"github.com/antclaw/antclaw/internal/service/user"
	"github.com/antclaw/antclaw/internal/service/vol"
)

func main() {
boot := time.Now()
	if err := auth.LoadKeys(); err != nil {
		log.Fatalf("failed to load JWT keys: %v", err)
	}

	// Database connection
	pgPool, err := infrapq.NewPoolFromEnv()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pgPool.Close()
	// 启动自检：确保审计等 Admin 相关表存在（幂等）
	if err := postgres.EnsureAdminSchema(context.Background(), pgPool); err != nil {
		log.Fatalf("failed to ensure admin schema: %v", err)
	}
	queries := db.New(pgPool)

	// Redis connection for SSE events and nonce dedup
	redisClient := redis.NewClientFromEnv()
	if err := redisClient.Ping(context.Background()); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	// ===== 加密能力初始化 =====
	// SecretBox：用于数据源敏感字段的存储级加密（Argon2id+AES-GCM）
	secretBox, err := cryptopkg.NewSecretBox(os.Getenv("ANTCLAW_SECRET_MASTER_KEY"))
	if err != nil {
		log.Fatalf("failed to init SecretBox: %v", err)
	}
	rsaKeyPath := os.Getenv("ANTCLAW_RSA_KEY_PATH")
	if rsaKeyPath == "" {
		rsaKeyPath = "/data/rsa_private.pem"
	}
	rsaMgr, err := cryptopkg.LoadOrCreateRSA(rsaKeyPath)
	if err != nil {
		log.Fatalf("failed to load/create RSA key: %v", err)
	}
	envFallback := datasource.BuildEnvFallback()
	resolver := datasource.NewCredentialResolver(pgPool, secretBox, envFallback, nil)
	if err := resolver.ReloadAll(context.Background()); err != nil {
		log.Printf("warn: warm-up credentials failed: %v", err)
	}
	fredKey := resolver.GetSecret("fred")
	fredClient := apiclient.NewFredClient(fredKey)
	mql5Fetcher := apiclient.NewMQL5Fetcher()

	// Initialize storage layer
	userStore := postgres.NewUserStore(queries)
	auditSvc := audit.NewAuditService(queries, redisClient)

	// Initialize all services with real data sources
	priceSvc := price.NewService()
	volSvc := vol.NewService()
	macroSvc := macro.NewServiceWithFRED(fredClient)
	taSvc := ta.NewService()
	sentimentSvc := sentiment.NewService()
	systemAISvc := systemaisvc.NewService(pgPool, secretBox)
	aiSvc := ai.NewService(systemAISvc, pgPool)
	alertsSvc := alerts.NewService(pgPool)
	userSvc := user.NewService()
	adminSvc := admin.NewService(queries, auditSvc, redisClient.Raw())
	calendarSvc := calendar.NewServiceWithFetcher(mql5Fetcher)
	cotSvc := cot.NewService()
	backtestSvc := backtest.NewService(pgPool)
	strategySvc := strategysvc.NewService(pgPool, strategysvc.NewMockRunner())

	priceProv := postgres.NewPricePgProvider(pgPool)
	cotProv := postgres.NewCOTPgProvider(pgPool)
	regimeProv := postgres.NewRegimePgProvider(pgPool)
	factorProv := postgres.NewFactorPgProvider(pgPool)
	volProv := postgres.NewVolPgProvider(pgPool)
	flowProv := postgres.NewFlowPgProvider(pgPool)
	signalRepo := postgres.NewSignalRepoPg(pgPool)
	signalsService := signalssvc.NewService(signalssvc.Deps{
		Price:  priceProv,
		COT:    cotProv,
		Regime: regimeProv,
		Factor: factorProv,
		Vol:    volProv,
		Flow:   flowProv,
		Signal: signalRepo,
	})
	reportSvc := reportsvc.NewService(signalsService, backtestSvc)

	// Initialize all handlers with dependency injection
	priceHandler := rpc.NewPriceHandler(priceSvc)
	volHandler := rpc.NewVolHandler(volSvc)
	signalsHandler := rpc.NewSignalsHandler(signalsService)
	macroHandler := rpc.NewMacroHandler(macroSvc)
	taHandler := rpc.NewTAHandler(taSvc)
	sentimentHandler := rpc.NewSentimentHandler(sentimentSvc)
	aiHandler := rpc.NewAIHandler(aiSvc)
	alertsHandler := rpc.NewAlertsHandler(alertsSvc)
	userHandler := rpc.NewUserHandler(userSvc)
	adminHandler := rpc.NewAdminHandler(adminSvc)
	calendarHandler := rpc.NewCalendarHandler(calendarSvc)
	cotHandler := rpc.NewCOTHandler(cotSvc)
	backtestHandler := rpc.NewBacktestHandler(backtestSvc)
	reportConnectHandler := rpc.NewReportHandler(reportSvc)
	strategyHandler := rpc.NewStrategyHandler(strategySvc)
	systemAIHandler := rpc.NewSystemAIConnectHandler(systemAISvc)
	dataSourceSvc := datasource.NewService(pgPool, secretBox, redisClient.Raw())
	dataSourceHandler := rpc.NewDataSourceConnectHandler(dataSourceSvc)
	mt4Handler := rpc.NewMT4Handler()
	mt5Handler := rpc.NewMT5Handler()

	// AuthHandler with real PostgreSQL store
	authHandler := rpc.NewAuthHandler(userStore, nil, auditSvc)

	// Create HTTP mux and register Connect RPC handlers
	mux := http.NewServeMux()

	// Register all service handlers
	mux.Handle(antclawv1connect.NewAuthServiceHandler(authHandler))
	cryptoHandler := rpc.NewCryptoConnectHandler(rsaMgr, redisClient)
	mux.Handle(antclawv1connect.NewCryptoServiceHandler(cryptoHandler))
	mux.Handle(antclawv1connect.NewPriceServiceHandler(priceHandler))
	mux.Handle(antclawv1connect.NewVolServiceHandler(volHandler))
	mux.Handle(antclawv1connect.NewSignalsServiceHandler(signalsHandler))
	mux.Handle(antclawv1connect.NewMacroServiceHandler(macroHandler))
	mux.Handle(antclawv1connect.NewTAServiceHandler(taHandler))
	mux.Handle(antclawv1connect.NewSentimentServiceHandler(sentimentHandler))
	mux.Handle(antclawv1connect.NewAIServiceHandler(aiHandler))
	mux.Handle(antclawv1connect.NewAlertServiceHandler(alertsHandler))
	mux.Handle(antclawv1connect.NewUserServiceHandler(userHandler))
	mux.Handle(antclawv1connect.NewAdminServiceHandler(adminHandler))
	mux.Handle(antclawv1connect.NewCalendarServiceHandler(calendarHandler))
	mux.Handle(antclawv1connect.NewCOTServiceHandler(cotHandler))
	mux.Handle(antclawv1connect.NewBacktestServiceHandler(backtestHandler))
	mux.Handle(antclawv1connect.NewReportServiceHandler(reportConnectHandler))
	mux.Handle(antclawv1connect.NewStrategyServiceHandler(strategyHandler))
	mux.Handle(antclawv1connect.NewSystemAIServiceHandler(systemAIHandler))
	mux.Handle(antclawv1connect.NewDataSourceServiceHandler(dataSourceHandler))
	adminDataHandler := rpc.NewAdminDataConnectHandler(pgPool)
	mux.Handle(antclawv1connect.NewAdminDataServiceHandler(adminDataHandler))
	mux.Handle(antclawv1connect.NewMT4ServiceHandler(mt4Handler))
	mux.Handle(antclawv1connect.NewMT5ServiceHandler(mt5Handler))
	// System service (Connect health/info)
	systemHandler := rpc.NewSystemHandler(pgPool, redisClient, boot)
	mux.Handle(antclawv1connect.NewSystemServiceHandler(systemHandler))

	// M2~M4 新增 Service（骨架，逐步落地真实数据）
	mux.Handle(antclawv1connect.NewOptionsServiceHandler(rpc.NewOptionsHandler()))
	mux.Handle(antclawv1connect.NewOnchainServiceHandler(rpc.NewOnchainHandler(pgPool)))
	mux.Handle(antclawv1connect.NewDeFiServiceHandler(rpc.NewDeFiHandler()))
	mux.Handle(antclawv1connect.NewSECServiceHandler(rpc.NewSECHandler()))
	mux.Handle(antclawv1connect.NewFedWatchServiceHandler(rpc.NewFedWatchHandler()))
	mux.Handle(antclawv1connect.NewMacroExtrasServiceHandler(rpc.NewMacroExtrasHandler()))
	mux.Handle(antclawv1connect.NewTreasuryServiceHandler(rpc.NewTreasuryHandler()))
	mux.Handle(antclawv1connect.NewSentimentExtrasServiceHandler(rpc.NewSentimentExtrasHandler()))
	mux.Handle(antclawv1connect.NewRegimeServiceHandler(rpc.NewRegimeHandler(regime.NewService(pgPool))))

	// CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	// Health check endpoint
	
	// SSE endpoints for real-time jobs & audit logs
	mux.HandleFunc("/sse/jobs", jobsEventsHandler(redisClient.Raw()))
	mux.HandleFunc("/sse/audit", auditEventsHandler(redisClient.Raw()))

	// Create HTTP server with h2c (HTTP/2 without TLS)
	server := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(corsHandler(mux), &http2.Server{}),
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
