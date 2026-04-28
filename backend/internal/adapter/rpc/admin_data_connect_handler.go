package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"connectrpc.com/connect"
	admindatav1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AdminDataConnectHandler 采集数据汇总与预览 Connect 实现。
type AdminDataConnectHandler struct {
	pool *pgxpool.Pool
}

// NewAdminDataConnectHandler 创建 handler。
func NewAdminDataConnectHandler(pool *pgxpool.Pool) *AdminDataConnectHandler {
	return &AdminDataConnectHandler{pool: pool}
}

func (h *AdminDataConnectHandler) GetDataSummary(ctx context.Context, _ *connect.Request[admindatav1.GetDataSummaryRequest]) (*connect.Response[admindatav1.GetDataSummaryResponse], error) {
	summaries := CollectDataSummaries(ctx, h.pool)
	items := make([]*admindatav1.DataSummaryItem, 0, len(summaries))
	for _, s := range summaries {
		items = append(items, &admindatav1.DataSummaryItem{
			JobId:      s.JobID,
			Name:       s.Name,
			Table:      s.Table,
			Count:      s.Count,
			LatestTime: s.LatestTime,
			Error:      s.Error,
		})
	}
	return connect.NewResponse(&admindatav1.GetDataSummaryResponse{
		Items:     items,
		UpdatedAt: time.Now().Unix(),
	}), nil
}

func (h *AdminDataConnectHandler) GetDataPreview(ctx context.Context, req *connect.Request[admindatav1.GetDataPreviewRequest]) (*connect.Response[admindatav1.GetDataPreviewResponse], error) {
	limit := int(req.Msg.GetLimit())
	if limit == 0 {
		limit = 50
	}
	prev, err := FetchDataPreview(ctx, h.pool, req.Msg.GetJobId(), limit)
	if err != nil {
		switch {
		case errors.Is(err, ErrPreviewMissingJobID), errors.Is(err, ErrPreviewUnknownJobID):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	b, err := json.Marshal(prev.Rows)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&admindatav1.GetDataPreviewResponse{
		JobId:        prev.JobID,
		Table:        prev.Table,
		TimeCol:      prev.TimeCol,
		Columns:      prev.Columns,
		RowsJson:     string(b),
		TotalSampled: int32(prev.TotalSampled),
	}), nil
}

var _ antclawv1connect.AdminDataServiceHandler = (*AdminDataConnectHandler)(nil)
