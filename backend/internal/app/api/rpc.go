package api

import (
	"net/http"
	"time"

	"connectrpc.com/connect"

	antclawv1connect "github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/adapter/rpc"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres"
	"github.com/antclaw/antclaw/internal/auth"
	"github.com/antclaw/antclaw/internal/service/alerts"
	"github.com/antclaw/antclaw/internal/service/regime"
)

// registerHandlers constructs all Connect-RPC handlers and registers them on mux.
func registerHandlers(mux *http.ServeMux, inf *Infra, svc *Services, boot time.Time) {
	pg := inf.PGPool

	// ── Core ──
	mux.Handle(antclawv1connect.NewAuthServiceHandler(rpc.NewAuthHandler(
		postgres.NewUserStore(inf.Queries), nil, svc.Audit, pg)))
	mux.Handle(antclawv1connect.NewCryptoServiceHandler(rpc.NewCryptoConnectHandler(inf.RSAMgr, inf.RDB)))
	mux.Handle(antclawv1connect.NewSystemServiceHandler(rpc.NewSystemHandler(pg, inf.RDB, boot, svc.Presence)))

	// ── Client services ──
	mux.Handle(antclawv1connect.NewPriceServiceHandler(rpc.NewPriceHandler(svc.Price)))
	mux.Handle(antclawv1connect.NewVolServiceHandler(rpc.NewVolHandler(svc.Vol)))
	mux.Handle(antclawv1connect.NewSignalsServiceHandler(rpc.NewSignalsHandler(svc.Signals).WithCalibration(svc.CalibStore)))
	mux.Handle(antclawv1connect.NewMacroServiceHandler(rpc.NewMacroHandler(svc.Macro)))
	mux.Handle(antclawv1connect.NewTAServiceHandler(rpc.NewTAHandler(svc.TA)))
	mux.Handle(antclawv1connect.NewSentimentServiceHandler(rpc.NewSentimentHandler(svc.Sentiment)))
	mux.Handle(antclawv1connect.NewAIServiceHandler(rpc.NewAIHandler(svc.AI)))
	mux.Handle(antclawv1connect.NewAlertServiceHandler(rpc.NewAlertsHandler(svc.Alerts).WithGate(alerts.NewGate(pg))))
	mux.Handle(antclawv1connect.NewUserServiceHandler(rpc.NewUserHandler(svc.User)))
	mux.Handle(antclawv1connect.NewCalendarServiceHandler(rpc.NewCalendarHandler(svc.Calendar)))
	mux.Handle(antclawv1connect.NewCOTServiceHandler(rpc.NewCOTHandler(svc.COT)))
	mux.Handle(antclawv1connect.NewBacktestServiceHandler(rpc.NewBacktestHandler(svc.Backtest)))
	mux.Handle(antclawv1connect.NewReportServiceHandler(rpc.NewReportHandler(svc.Report)))
	mux.Handle(antclawv1connect.NewStrategyServiceHandler(rpc.NewStrategyHandler(svc.Strategy)))
	mux.Handle(antclawv1connect.NewOptionsServiceHandler(rpc.NewOptionsHandler()))
	mux.Handle(antclawv1connect.NewOnchainServiceHandler(rpc.NewOnchainHandler(pg)))
	mux.Handle(antclawv1connect.NewDeFiServiceHandler(rpc.NewDeFiHandler()))
	mux.Handle(antclawv1connect.NewSECServiceHandler(rpc.NewSECHandler()))
	mux.Handle(antclawv1connect.NewFedWatchServiceHandler(rpc.NewFedWatchHandlerWithResolver(inf.Resolver)))
	mux.Handle(antclawv1connect.NewMacroExtrasServiceHandler(rpc.NewMacroExtrasHandler()))
	mux.Handle(antclawv1connect.NewTreasuryServiceHandler(rpc.NewTreasuryHandler()))
	mux.Handle(antclawv1connect.NewSentimentExtrasServiceHandler(rpc.NewSentimentExtrasHandlerWithResolver(inf.Resolver)))
	mux.Handle(antclawv1connect.NewRegimeServiceHandler(rpc.NewRegimeHandler(regime.NewService(pg))))
	mux.Handle(antclawv1connect.NewDeviceServiceHandler(rpc.NewDeviceHandler(pg),
		connect.WithInterceptors(auth.AuthInterceptor(true))))
	mux.Handle(antclawv1connect.NewMT4ServiceHandler(rpc.NewMT4Handler(svc.MT)))
	mux.Handle(antclawv1connect.NewMT5ServiceHandler(rpc.NewMT5Handler(svc.MT)))

	// ── Admin (auth + admin guard) ──
	adminInt := connect.WithInterceptors(auth.AuthInterceptor(true), auth.AdminInterceptor())
	mux.Handle(antclawv1connect.NewAdminServiceHandler(
		rpc.NewAdminHandler(svc.Admin, svc.Notify, svc.Presence, pg), adminInt))
	mux.Handle(antclawv1connect.NewSystemAIServiceHandler(rpc.NewSystemAIConnectHandler(svc.SystemAI), adminInt))
	mux.Handle(antclawv1connect.NewDataSourceServiceHandler(rpc.NewDataSourceConnectHandler(svc.DataSource), adminInt))
	mux.Handle(antclawv1connect.NewAdminDataServiceHandler(rpc.NewAdminDataConnectHandler(pg), adminInt))

	// ── Social services: mixed public/read and auth-required/write ──
	// See docs/安卓客户端技术文档包/12-服务端社交板块整改落地指南.md §S12-P0-06
	mux.Handle(antclawv1connect.NewFeedServiceHandler(rpc.NewFeedHandler(svc.Feed)))
	mux.Handle(antclawv1connect.NewTraderServiceHandler(rpc.NewTraderHandler(svc.Trader)))
	mux.Handle(antclawv1connect.NewChatServiceHandler(rpc.NewChatHandler(pg)))
	// CircleService disabled — circle publishing is not yet supported.
	// mux.Handle(antclawv1connect.NewCircleServiceHandler(rpc.NewCircleHandler(pg)))
	mux.Handle(antclawv1connect.NewMarketplaceServiceHandler(rpc.NewMarketplaceHandler(pg)))
	mux.Handle(antclawv1connect.NewSearchServiceHandler(rpc.NewSearchHandler(svc.Search)))
	mux.Handle(antclawv1connect.NewTrendServiceHandler(rpc.NewTrendHandler(svc.Trend)))
	mux.Handle(antclawv1connect.NewNotificationServiceHandler(
		rpc.NewNotificationHandler(svc.Notify, inf.Queries),
		connect.WithInterceptors(auth.AuthInterceptor(true))))

	// ── Streaming (protobuf binary, replaces SSE) ──
	mux.Handle(antclawv1connect.NewStreamServiceHandler(rpc.NewStreamHandler(inf.RDB.Raw())))
}
