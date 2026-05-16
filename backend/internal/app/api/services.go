package api

import (
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres"
	infrapq "github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/service/admin"
	"github.com/antclaw/antclaw/internal/service/ai"
	"github.com/antclaw/antclaw/internal/service/alerts"
	"github.com/antclaw/antclaw/internal/service/audit"
	"github.com/antclaw/antclaw/internal/service/backtest"
	"github.com/antclaw/antclaw/internal/service/calibration"
	"github.com/antclaw/antclaw/internal/service/calendar"
	"github.com/antclaw/antclaw/internal/service/cot"
	"github.com/antclaw/antclaw/internal/service/datasource"
	feedpkg "github.com/antclaw/antclaw/internal/service/feed"
	"github.com/antclaw/antclaw/internal/service/macro"
	mtsvc "github.com/antclaw/antclaw/internal/service/mt"
	"github.com/antclaw/antclaw/internal/notify"
	"github.com/antclaw/antclaw/internal/service/presence"
	"github.com/antclaw/antclaw/internal/service/price"
	reportsvc "github.com/antclaw/antclaw/internal/service/report"
	searchpkg "github.com/antclaw/antclaw/internal/service/search"
	"github.com/antclaw/antclaw/internal/service/sentiment"
	signalssvc "github.com/antclaw/antclaw/internal/service/signals"
	strategysvc "github.com/antclaw/antclaw/internal/service/strategy"
	systemaisvc "github.com/antclaw/antclaw/internal/service/systemai"
	"github.com/antclaw/antclaw/internal/service/ta"
	traderpkg "github.com/antclaw/antclaw/internal/service/trader"
	trendpkg "github.com/antclaw/antclaw/internal/service/trend"
	"github.com/antclaw/antclaw/internal/service/user"
	"github.com/antclaw/antclaw/internal/service/vol"
)

// Services holds all business services constructed with real dependencies.
type Services struct {
	Price      *price.Service
	Vol        *vol.Service
	Macro      *macro.Service
	TA         *ta.Service
	Sentiment  *sentiment.Service
	SystemAI   *systemaisvc.Service
	AI         *ai.Service
	Alerts     *alerts.Service
	User       *user.Service
	Admin      *admin.Service
	Calendar   *calendar.Service
	COT        *cot.Service
	Backtest   *backtest.Service
	Strategy   *strategysvc.Service
	Signals    *signalssvc.Service
	Report     *reportsvc.Service
	Notify     *notify.Service
	DataSource *datasource.Service
	Feed       *feedpkg.Service
	Trader     *traderpkg.Service
	Search     *searchpkg.Service
	Trend      *trendpkg.Service
	MT         *mtsvc.Service
	CalibStore *calibration.Store
	Audit      *audit.AuditService
	Presence   *presence.Tracker
}

// InitServices constructs all business services from infrastructure.
func InitServices(inf *Infra) *Services {
	pgPool := inf.PGPool
	queries := inf.Queries

	priceSvc := price.NewServiceWithPool(pgPool)
	volSvc := vol.NewService().WithFirecrawl(inf.Firecrawl)
	macroSvc := macro.NewServiceWithFRED(inf.FredClient)
	taSvc := ta.NewServiceWithPool(pgPool)
	sentimentSvc := sentiment.NewServiceWithPool(pgPool)
	systemAISvc := systemaisvc.NewService(pgPool, inf.SecretBox)
	aiSvc := ai.NewService(systemAISvc, pgPool)
	alertsSvc := alerts.NewService(pgPool)
	userSvc := user.NewService()
	auditSvc := audit.NewAuditService(queries, inf.RDB)
	adminSvc := admin.NewService(queries, auditSvc, inf.RDB.Raw())
	calendarSvc := calendar.NewServiceWithFetcher(inf.MQL5Fetcher)
	cotSvc := cot.NewService()
	backtestSvc := backtest.NewService(pgPool)
	strategySvc := strategysvc.NewService(pgPool, strategysvc.NewBaselineRunner(pgPool))

	priceProv := postgres.NewPricePgProvider(pgPool)
	cotProv := postgres.NewCOTPgProvider(pgPool)
	regimeProv := postgres.NewRegimePgProvider(pgPool)
	factorProv := postgres.NewFactorPgProvider(pgPool)
	volProv := postgres.NewVolPgProvider(pgPool)
	flowProv := postgres.NewFlowPgProvider(pgPool)
	signalRepo := postgres.NewSignalRepoPg(pgPool)
	signalsSvc := signalssvc.NewService(signalssvc.Deps{
		Price:  priceProv, COT: cotProv, Regime: regimeProv,
		Factor: factorProv, Vol: volProv, Flow: flowProv, Signal: signalRepo,
	})
	reportSvc := reportsvc.NewService(signalsSvc, backtestSvc)
	notifySvc := notify.NewService(queries, inf.RDB.Raw())
	presenceTracker := presence.NewTracker()

	feedRepo := infrapq.NewFeedRepository(pgPool)
	feedSvc := feedpkg.NewService(feedRepo)
	traderRepo := infrapq.NewTraderRepository(pgPool)
	traderSvc := traderpkg.NewService(traderRepo)
	searchRepo := infrapq.NewSearchRepository(pgPool)
	searchSvc := searchpkg.NewService(searchRepo)
	trendRepo := infrapq.NewTrendRepository(pgPool)
	trendSvc := trendpkg.NewService(trendRepo)

	mtRepo := infrapq.NewMTAccountRepository(pgPool)
	mtConnMgr := mtsvc.NewConnectionManager(inf.MT4GWURL, inf.MT5GWURL)
	mtSvc := mtsvc.NewService(mtRepo, mtConnMgr)
	_ = mtConnMgr // retained for lifecycle

	dataSourceSvc := datasource.NewService(pgPool, inf.SecretBox, inf.RDB.Raw())
	calibStore := calibration.NewStore(pgPool)

	return &Services{
		Price: priceSvc, Vol: volSvc, Macro: macroSvc, TA: taSvc,
		Sentiment: sentimentSvc, SystemAI: systemAISvc, AI: aiSvc,
		Alerts: alertsSvc, User: userSvc, Admin: adminSvc,
		Calendar: calendarSvc, COT: cotSvc, Backtest: backtestSvc,
		Strategy: strategySvc, Signals: signalsSvc, Report: reportSvc,
		Notify: notifySvc, DataSource: dataSourceSvc,
		Feed: feedSvc, Trader: traderSvc, Search: searchSvc, Trend: trendSvc,
		MT: mtSvc, CalibStore: calibStore, Audit: auditSvc, Presence: presenceTracker,
	}
}
