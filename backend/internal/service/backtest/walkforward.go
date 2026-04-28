package backtest

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 价格序列（按时间升序）
type priceSeries struct {
	Times  []time.Time
	Closes []float64
}

// loadDailyCloses 从 price_daily 拉取 [from, to] 区间内 symbol 的日线收盘价。
func loadDailyCloses(ctx context.Context, pool *pgxpool.Pool, symbol string, from, to time.Time) (*priceSeries, error) {
	rows, err := pool.Query(ctx, `
		SELECT time, close
		  FROM price_daily
		 WHERE symbol = $1 AND time BETWEEN $2 AND $3 AND close IS NOT NULL AND close > 0
		 ORDER BY time ASC`, symbol, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := &priceSeries{}
	for rows.Next() {
		var t time.Time
		var c float64
		if err := rows.Scan(&t, &c); err != nil {
			return nil, err
		}
		out.Times = append(out.Times, t)
		out.Closes = append(out.Closes, c)
	}
	return out, nil
}

// smaCrossoverReturns 计算 SMA(short)/SMA(long) 双均线策略在序列上的逐日收益。
// 持仓规则：short SMA > long SMA → 全仓多；否则空仓。
// 输入 closes 必须时间升序。
func smaCrossoverReturns(closes []float64, short, long int) []float64 {
	n := len(closes)
	if n <= long+1 || short <= 0 || long <= short {
		return nil
	}
	rets := make([]float64, 0, n-long-1)
	for i := long; i < n-1; i++ {
		ssum, lsum := 0.0, 0.0
		for j := i - short + 1; j <= i; j++ {
			ssum += closes[j]
		}
		for j := i - long + 1; j <= i; j++ {
			lsum += closes[j]
		}
		signal := 0.0
		if ssum/float64(short) > lsum/float64(long) {
			signal = 1.0
		}
		dailyRet := (closes[i+1] - closes[i]) / closes[i]
		rets = append(rets, signal*dailyRet)
	}
	return rets
}

// sharpe 估算年化 Sharpe（日频，rf=0，ann=252）。样本不足时返回 0。
func sharpe(rets []float64) float64 {
	if len(rets) < 5 {
		return 0
	}
	mean := 0.0
	for _, r := range rets {
		mean += r
	}
	mean /= float64(len(rets))
	variance := 0.0
	for _, r := range rets {
		variance += (r - mean) * (r - mean)
	}
	variance /= float64(len(rets) - 1)
	std := math.Sqrt(variance)
	if std == 0 {
		return 0
	}
	return (mean / std) * math.Sqrt(252)
}

// pickBestSMA 在参数网格上选择 IS Sharpe 最高的 (short,long)。
func pickBestSMA(closes []float64, shorts, longs []int) (int, int, float64) {
	bestS, bestL := shorts[0], longs[0]
	best := math.Inf(-1)
	for _, s := range shorts {
		for _, l := range longs {
			if l <= s {
				continue
			}
			sh := sharpe(smaCrossoverReturns(closes, s, l))
			if sh > best {
				best, bestS, bestL = sh, s, l
			}
		}
	}
	if math.IsInf(best, -1) {
		best = 0
	}
	return bestS, bestL, best
}

// runWalkforwardSMA 在给定价格序列上做 K 折滚动训练-测试。
// trainRatio: 训练窗占整折比例（默认 0.7）。
type wfFoldResult struct {
	FoldIdx                          int
	TrainFrom, TrainTo               time.Time
	TestFrom, TestTo                 time.Time
	BestShort, BestLong              int
	InSampleSharpe, OutOfSampleSharpe float64
}

func runWalkforwardSMA(ps *priceSeries, folds int, trainRatio float64) ([]wfFoldResult, error) {
	if folds <= 0 {
		folds = 5
	}
	if trainRatio <= 0 || trainRatio >= 1 {
		trainRatio = 0.7
	}
	n := len(ps.Closes)
	if n < folds*60 { // 每折至少 60 根 K 线，避免过拟合噪声
		return nil, fmt.Errorf("price bars insufficient: have %d, need >= %d", n, folds*60)
	}
	foldSize := n / folds
	shorts := []int{5, 10, 20}
	longs := []int{30, 50, 100}
	out := make([]wfFoldResult, 0, folds)
	for i := 0; i < folds; i++ {
		startIdx := i * foldSize
		endIdx := startIdx + foldSize
		if i == folds-1 {
			endIdx = n
		}
		trainEnd := startIdx + int(float64(endIdx-startIdx)*trainRatio)
		trainCloses := ps.Closes[startIdx:trainEnd]
		testCloses := ps.Closes[trainEnd:endIdx]
		s, l, isSharpe := pickBestSMA(trainCloses, shorts, longs)
		oosSharpe := sharpe(smaCrossoverReturns(testCloses, s, l))
		out = append(out, wfFoldResult{
			FoldIdx:           i + 1,
			TrainFrom:         ps.Times[startIdx],
			TrainTo:           ps.Times[trainEnd-1],
			TestFrom:          ps.Times[trainEnd],
			TestTo:            ps.Times[endIdx-1],
			BestShort:         s,
			BestLong:          l,
			InSampleSharpe:    isSharpe,
			OutOfSampleSharpe: oosSharpe,
		})
	}
	return out, nil
}

