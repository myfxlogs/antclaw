package api

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antclaw/antclaw/internal/adapter/storage/postgres"
	"github.com/antclaw/antclaw/internal/adapter/storage/postgres/db"
	"github.com/antclaw/antclaw/internal/auth"
	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	"github.com/antclaw/antclaw/internal/infra/apiclient/firecrawl"
	"github.com/antclaw/antclaw/internal/infra/apiclient/fred"
	"github.com/antclaw/antclaw/internal/infra/apiclient/mql5"
	infrapq "github.com/antclaw/antclaw/internal/infra/postgres"
	"github.com/antclaw/antclaw/internal/infra/redis"
	"github.com/antclaw/antclaw/internal/service/datasource"
)

// Infra holds all infrastructure components needed by the API server.
type Infra struct {
	PGPool      *pgxpool.Pool
	Queries     *db.Queries
	RDB         *redis.Client
	SecretBox   *cryptopkg.SecretBox
	RSAMgr      *cryptopkg.RSAManager
	Resolver    *datasource.CredentialResolver
	FredClient  *fred.Client
	MQL5Fetcher *mql5.Fetcher
	Firecrawl   *firecrawl.Client
	MT4GWURL    string
	MT5GWURL    string
}

// InitInfra initializes all infrastructure from config.
func InitInfra(cfg Config) (*Infra, error) {
	if err := auth.LoadKeys(); err != nil {
		return nil, err
	}

	pgPool, err := infrapq.NewPoolFromEnv()
	if err != nil {
		return nil, err
	}

	if err := postgres.RunMigrations(context.Background(), pgPool); err != nil {
		pgPool.Close()
		return nil, err
	}
	if err := postgres.EnsureAdminSchema(context.Background(), pgPool); err != nil {
		pgPool.Close()
		return nil, err
	}

	rdb := redis.NewClientFromEnv()
	if err := rdb.Ping(context.Background()); err != nil {
		pgPool.Close()
		return nil, err
	}

	masterKey, err := cryptopkg.LoadOrCreateMasterKey()
	if err != nil {
		pgPool.Close()
		return nil, err
	}
	secretBox, err := cryptopkg.NewSecretBox(masterKey)
	if err != nil {
		pgPool.Close()
		return nil, err
	}

	rsaMgr, err := cryptopkg.LoadOrCreateRSA(cfg.RSAKeyPath)
	if err != nil {
		pgPool.Close()
		return nil, err
	}

	envFallback := datasource.BuildEnvFallback()
	resolver := datasource.NewCredentialResolver(pgPool, secretBox, envFallback, nil)
	if err := resolver.ReloadAll(context.Background()); err != nil {
		log.Printf("warn: warm-up credentials failed: %v", err)
	}

	fredKey := resolver.GetSecret("fred")
	fredClient := fred.NewClient(fredKey)
	mql5Fetcher := mql5.NewFetcher()

	fcKey := resolver.GetSecret("firecrawl")
	fcSrc := apiclient.NewSource("firecrawl", apiclient.Options{Timeout: 60 * time.Second})
	fcClient := firecrawl.NewClientWithKey(fcSrc, fcKey)

	return &Infra{
		PGPool:      pgPool,
		Queries:     db.New(pgPool),
		RDB:         rdb,
		SecretBox:   secretBox,
		RSAMgr:      rsaMgr,
		Resolver:    resolver,
		FredClient:  fredClient,
		MQL5Fetcher: mql5Fetcher,
		Firecrawl:   fcClient,
		MT4GWURL:    cfg.MT4GWURL,
		MT5GWURL:    cfg.MT5GWURL,
	}, nil
}

// Close shuts down infrastructure resources.
func (inf *Infra) Close() {
	inf.PGPool.Close()
}
