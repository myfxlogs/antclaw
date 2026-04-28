package factors

import (
	"context"
	"time"
)

type Factor interface {
	Name() string
	Compute(ctx context.Context, symbol string, asOf time.Time) (raw float64, ok bool, err error)
}
