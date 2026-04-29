package calibration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store 缓存内存中的 Calibrator + 持久化到 signal_calibrations。
type Store struct {
	pool *pgxpool.Pool
	mu   sync.RWMutex
	mem  map[string]Calibrator
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, mem: make(map[string]Calibrator)}
}

// Save 写库 + 内存。fittedAt 自动 NOW()。
func (s *Store) Save(ctx context.Context, modelID string, c Calibrator, nSamples int, brier float64) error {
	if modelID == "" || c == nil {
		return fmt.Errorf("calibration store: empty modelID or calibrator")
	}
	params, err := serialize(c)
	if err != nil {
		return err
	}
	if s.pool != nil {
		_, err = s.pool.Exec(ctx, `
			INSERT INTO signal_calibrations(model_id, type, params, n_samples, brier, fitted_at)
			VALUES ($1,$2,$3,$4,$5,NOW())
			ON CONFLICT (model_id) DO UPDATE SET
				type = EXCLUDED.type,
				params = EXCLUDED.params,
				n_samples = EXCLUDED.n_samples,
				brier = EXCLUDED.brier,
				fitted_at = NOW()`,
			modelID, c.Type(), params, nSamples, brier)
		if err != nil {
			return err
		}
	}
	s.mu.Lock()
	s.mem[modelID] = c
	s.mu.Unlock()
	return nil
}

// Load 内存命中优先；未命中则从 DB 反序列化。
func (s *Store) Load(ctx context.Context, modelID string) (Calibrator, error) {
	s.mu.RLock()
	c, ok := s.mem[modelID]
	s.mu.RUnlock()
	if ok {
		return c, nil
	}
	if s.pool == nil {
		return nil, fmt.Errorf("calibration store: model %q not in memory and no DB", modelID)
	}
	var typ string
	var raw []byte
	err := s.pool.QueryRow(ctx, `SELECT type, params FROM signal_calibrations WHERE model_id = $1`, modelID).
		Scan(&typ, &raw)
	if err != nil {
		return nil, fmt.Errorf("calibration store: %w", err)
	}
	c, err = deserialize(typ, raw)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.mem[modelID] = c
	s.mu.Unlock()
	return c, nil
}

// List 返回所有已保存校准模型的元数据。
type Summary struct {
	ModelID  string
	Type     string
	NSamples int
	Brier    float64
	FittedAt time.Time
}

func (s *Store) List(ctx context.Context) ([]Summary, error) {
	if s.pool == nil {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `SELECT model_id, type, COALESCE(n_samples,0), COALESCE(brier,0), fitted_at FROM signal_calibrations ORDER BY fitted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Summary{}
	for rows.Next() {
		var s Summary
		if err := rows.Scan(&s.ModelID, &s.Type, &s.NSamples, &s.Brier, &s.FittedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// serialize / deserialize：JSON 形式存 params。
func serialize(c Calibrator) ([]byte, error) {
	switch v := c.(type) {
	case *Platt:
		return json.Marshal(map[string]any{"A": v.A, "B": v.B, "n": v.NSample})
	case *Isotonic:
		return json.Marshal(map[string]any{"X": v.X, "P": v.P, "n": v.N})
	}
	return nil, fmt.Errorf("calibration: unknown type %s", c.Type())
}

func deserialize(typ string, raw []byte) (Calibrator, error) {
	switch typ {
	case "platt":
		var m struct {
			A, B float64
			N    int `json:"n"`
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		return &Platt{A: m.A, B: m.B, NSample: m.N}, nil
	case "isotonic":
		var m struct {
			X []float64 `json:"X"`
			P []float64 `json:"P"`
			N int       `json:"n"`
		}
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, err
		}
		return &Isotonic{X: m.X, P: m.P, N: m.N}, nil
	}
	return nil, fmt.Errorf("calibration: unknown type %q", typ)
}
