package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
)

// OnchainHandler 链上指标服务，由 worker 采集落表 onchain_metrics 提供数据。
type OnchainHandler struct {
	db *pgxpool.Pool
}

func NewOnchainHandler(db *pgxpool.Pool) *OnchainHandler { return &OnchainHandler{db: db} }

func (h *OnchainHandler) GetMetrics(ctx context.Context, req *connect.Request[v1.OnchainServiceGetMetricsRequest]) (*connect.Response[v1.OnchainServiceGetMetricsResponse], error) {
	asset := req.Msg.Asset
	if asset == "" {
		asset = "BTC"
	}
	end := time.Now().UTC()
	if req.Msg.End != nil {
		end = req.Msg.End.AsTime()
	}
	start := end.AddDate(0, -3, 0)
	if req.Msg.Start != nil {
		start = req.Msg.Start.AsTime()
	}
	// onchain_metrics 为长表 (time, asset, metric, value)，做 PIVOT 到宽表
	rows, err := h.db.Query(ctx, `
		SELECT time::date AS d,
		       MAX(CASE WHEN metric IN ('active_addresses','active_addr') THEN value END) AS active_addr,
		       MAX(CASE WHEN metric IN ('tx_count','transactions')         THEN value END) AS tx_count,
		       MAX(CASE WHEN metric IN ('exchange_netflow','net_flow')     THEN value END) AS netflow,
		       MAX(CASE WHEN metric = 'mvrv' THEN value END) AS mvrv,
		       MAX(CASE WHEN metric = 'sopr' THEN value END) AS sopr
		  FROM onchain_metrics
		 WHERE asset = $1 AND time BETWEEN $2 AND $3
		 GROUP BY time::date
		 ORDER BY d ASC`, asset, start, end)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	defer rows.Close()
	out := &v1.OnchainServiceGetMetricsResponse{}
	for rows.Next() {
		var d time.Time
		var addr, tx, flow, mvrv, sopr *float64
		if err := rows.Scan(&d, &addr, &tx, &flow, &mvrv, &sopr); err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
		p := &v1.OnchainPoint{Time: timestamppb.New(d)}
		if addr != nil {
			p.ActiveAddresses = *addr
		}
		if tx != nil {
			p.TxCount = *tx
		}
		if flow != nil {
			p.ExchangeNetflow = *flow
		}
		if mvrv != nil {
			p.Mvrv = *mvrv
		}
		if sopr != nil {
			p.Sopr = *sopr
		}
		out.Points = append(out.Points, p)
	}
	return connect.NewResponse(out), nil
}

func (h *OnchainHandler) GetAnalysis(ctx context.Context, req *connect.Request[v1.OnchainServiceGetAnalysisRequest]) (*connect.Response[v1.OnchainServiceGetAnalysisResponse], error) {
	asset := req.Msg.Asset
	if asset == "" {
		asset = "BTC"
	}
	var score float64
	err := h.db.QueryRow(ctx, `
		SELECT COALESCE(AVG(value),0)
		  FROM onchain_metrics
		 WHERE asset = $1 AND metric = 'onchain_score' AND time >= NOW() - INTERVAL '30 days'`, asset).Scan(&score)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	regime := "neutral"
	switch {
	case score > 0.5:
		regime = "accumulation"
	case score < -0.5:
		regime = "distribution"
	}
	return connect.NewResponse(&v1.OnchainServiceGetAnalysisResponse{
		Regime: regime, Confidence: 0.5,
	}), nil
}

var _ antclawv1connect.OnchainServiceHandler = (*OnchainHandler)(nil)
