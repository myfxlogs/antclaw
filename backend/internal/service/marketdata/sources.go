package marketdata

import (
	"context"

	"github.com/antclaw/antclaw/internal/infra/apiclient/alphavantage"
	"github.com/antclaw/antclaw/internal/infra/apiclient/cryptocompare"
	"github.com/antclaw/antclaw/internal/infra/apiclient/twelvedata"
)

// twelveSource adapt twelvedata.Client.
type twelveSource struct{ c *twelvedata.Client }

func NewTwelveDataSource(c *twelvedata.Client) Source { return &twelveSource{c: c} }
func (s *twelveSource) Name() string                  { return s.c.Name() }
func (s *twelveSource) Available() bool               { return s.c.Available() }
func (s *twelveSource) FetchOHLC(ctx context.Context, sym, tf string, n int) ([]Bar, error) {
	bars, err := s.c.FetchOHLC(ctx, sym, tf, n)
	if err != nil {
		return nil, err
	}
	out := make([]Bar, len(bars))
	for i, b := range bars {
		out[i] = Bar{Time: b.Time, Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Volume: b.Volume}
	}
	return out, nil
}

// alphaSource adapt alphavantage.Client.
type alphaSource struct{ c *alphavantage.Client }

func NewAlphaVantageSource(c *alphavantage.Client) Source { return &alphaSource{c: c} }
func (s *alphaSource) Name() string                       { return s.c.Name() }
func (s *alphaSource) Available() bool                    { return s.c.Available() }
func (s *alphaSource) FetchOHLC(ctx context.Context, sym, tf string, n int) ([]Bar, error) {
	bars, err := s.c.FetchOHLC(ctx, sym, tf, n)
	if err != nil {
		return nil, err
	}
	out := make([]Bar, len(bars))
	for i, b := range bars {
		out[i] = Bar{Time: b.Time, Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Volume: b.Volume}
	}
	return out, nil
}

// cryptoSource adapt cryptocompare.Client.
type cryptoSource struct{ c *cryptocompare.Client }

func NewCryptoCompareSource(c *cryptocompare.Client) Source { return &cryptoSource{c: c} }
func (s *cryptoSource) Name() string                        { return s.c.Name() }
func (s *cryptoSource) Available() bool                     { return s.c.Available() }
func (s *cryptoSource) FetchOHLC(ctx context.Context, sym, tf string, n int) ([]Bar, error) {
	bars, err := s.c.FetchOHLC(ctx, sym, tf, n)
	if err != nil {
		return nil, err
	}
	out := make([]Bar, len(bars))
	for i, b := range bars {
		out[i] = Bar{Time: b.Time, Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Volume: b.Volume}
	}
	return out, nil
}
