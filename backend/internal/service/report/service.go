package report

import (
	"context"
	"time"

	"github.com/antclaw/antclaw/internal/service/backtest"
	"github.com/antclaw/antclaw/internal/service/signals"
)

type Service struct {
	signals  *signals.Service
	backtest *backtest.Service
}

func NewService(signalsSvc *signals.Service, backtestSvc *backtest.Service) *Service {
	return &Service{signals: signalsSvc, backtest: backtestSvc}
}

type Request struct {
	Symbol        string   `json:"symbol"`
	Sections      []string `json:"sections"`
	WithAISummary bool     `json:"with_ai_summary"`
}

type Response struct {
	Symbol          string                       `json:"symbol"`
	GeneratedAt     string                       `json:"generated_at"`
	Bias            *signals.BiasData            `json:"bias,omitempty"`
	Unified         *signals.UnifiedSignalRecord `json:"unified,omitempty"`
	Accuracy1W      float64                      `json:"accuracy_1w,omitempty"`
	Accuracy1M      float64                      `json:"accuracy_1m,omitempty"`
	MissingSections []string                     `json:"missing_sections,omitempty"`
}

func (s *Service) GetReport(ctx context.Context, req Request) (*Response, error) {
	out := &Response{Symbol: req.Symbol, GeneratedAt: time.Now().Format(time.RFC3339)}
	bias, err := s.signals.GetBias(ctx, req.Symbol, "1d")
	if err == nil {
		out.Bias = bias
	} else {
		out.MissingSections = append(out.MissingSections, "bias")
	}
	unified, err := s.signals.GetUnified(ctx, req.Symbol)
	if err == nil {
		out.Unified = unified
	} else {
		out.MissingSections = append(out.MissingSections, "unified")
	}

	acc1w, err := s.backtest.GetAccuracy(ctx, "unified:"+req.Symbol+":1W", nil)
	if err == nil && acc1w.Metrics != nil {
		out.Accuracy1W = acc1w.Metrics.DirectionalAccuracy
	} else {
		out.MissingSections = append(out.MissingSections, "accuracy_1w")
	}
	acc1m, err := s.backtest.GetAccuracy(ctx, "unified:"+req.Symbol+":1M", nil)
	if err == nil && acc1m.Metrics != nil {
		out.Accuracy1M = acc1m.Metrics.DirectionalAccuracy
	} else {
		out.MissingSections = append(out.MissingSections, "accuracy_1m")
	}
	return out, nil
}
