// Package rpc provides the AdminSocial Connect-RPC handler (A13-P0-04, A13-P0-05).
package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

// AdminSocialHandler implements AdminSocialServiceHandler.
type AdminSocialHandler struct {
	repo postgres.ModerationRepository
}

func NewAdminSocialHandler(repo postgres.ModerationRepository) *AdminSocialHandler {
	return &AdminSocialHandler{repo: repo}
}

func (h *AdminSocialHandler) ListPosts(ctx context.Context, req *connect.Request[alfqv1.AdminListPostsRequest]) (*connect.Response[alfqv1.AdminListPostsResponse], error) {
	msg := req.Msg
	limit := msg.PageSize
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var cursor *postgres.SocialCursor
	if msg.Cursor != "" {
		c, err := postgres.DecodeSocialCursor(msg.Cursor)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		cursor = c
	}
	rows, total, nextCursor, err := h.repo.ListPostsAdmin(ctx, msg.Keyword, msg.AuthorId, msg.PostType, msg.Visibility, msg.Status, cursor, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	posts := make([]*alfqv1.AdminPostSummary, len(rows))
	for i, r := range rows {
		preview := r.Content
		if len([]rune(preview)) > 200 {
			preview = string([]rune(preview)[:200])
		}
		posts[i] = &alfqv1.AdminPostSummary{
			PostId:         r.ID,
			AuthorId:       r.AuthorID,
			AuthorName:     r.AuthorName,
			ContentPreview: preview,
			PostType:       r.PostType,
			Visibility:     r.Visibility,
			LikeCount:      r.LikeCount,
			CommentCount:   r.CommentCount,
			ShareCount:     r.ShareCount,
			ReportCount:    int32(r.Score), // reused field
			CreatedAt:      r.CreatedAt.Unix(),
		}
	}
	return connect.NewResponse(&alfqv1.AdminListPostsResponse{
		Posts:      posts,
		NextCursor: postgres.EncodeSocialCursor(nextCursor),
		TotalCount: total,
	}), nil
}

func (h *AdminSocialHandler) GetPostDetail(ctx context.Context, req *connect.Request[alfqv1.AdminGetPostDetailRequest]) (*connect.Response[alfqv1.AdminPostDetail], error) {
	row, comments, _, err := h.repo.GetPostDetail(ctx, req.Msg.PostId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	p := &alfqv1.Post{
		Id:          row.ID,
		AuthorId:    row.AuthorID,
		AuthorName:  row.AuthorName,
		Content:     row.Content,
		PostType:    row.PostType,
		Visibility:  row.Visibility,
		LikeCount:   row.LikeCount,
		CommentCount: row.CommentCount,
		ShareCount:  row.ShareCount,
		CreatedAt:   row.CreatedAt.Unix(),
	}
	cmts := make([]*alfqv1.Comment, len(comments))
	for i, c := range comments {
		cmt := &alfqv1.Comment{Id: c.ID, PostId: c.PostID, AuthorId: c.AuthorID, AuthorName: c.AuthorName, Content: c.Content, CreatedAt: c.CreatedAt.Unix()}
		if c.ParentCommentID != nil {
			cmt.ParentCommentId = *c.ParentCommentID
		}
		cmts[i] = cmt
	}
	return connect.NewResponse(&alfqv1.AdminPostDetail{Post: p, Comments: cmts}), nil
}

func (h *AdminSocialHandler) ListComments(ctx context.Context, req *connect.Request[alfqv1.AdminListCommentsRequest]) (*connect.Response[alfqv1.AdminListCommentsResponse], error) {
	msg := req.Msg
	limit := msg.PageSize
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var c *postgres.SocialCursor
	if msg.Cursor != "" {
		dec, _ := postgres.DecodeSocialCursor(msg.Cursor)
		c = dec
	}
	rows, next, err := h.repo.ListCommentsAdmin(ctx, msg.PostId, msg.AuthorId, msg.Status, c, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	cmts := make([]*alfqv1.AdminCommentSummary, len(rows))
	for i, r := range rows {
		cmts[i] = &alfqv1.AdminCommentSummary{
			CommentId:  r.ID,
			PostId:     r.PostID,
			AuthorId:   r.AuthorID,
			AuthorName: r.AuthorName,
			Content:    r.Content,
			CreatedAt:  r.CreatedAt.Unix(),
		}
	}
	return connect.NewResponse(&alfqv1.AdminListCommentsResponse{Comments: cmts, NextCursor: postgres.EncodeSocialCursor(next)}), nil
}

func (h *AdminSocialHandler) ModerateContent(ctx context.Context, req *connect.Request[alfqv1.ModerateContentRequest]) (*connect.Response[alfqv1.ModerateContentResponse], error) {
	msg := req.Msg
	if msg.Reason == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("reason is required"))
	}
	if err := h.repo.UpdateContentStatus(ctx, msg.TargetType, msg.TargetId, msg.Action); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	c, _ := h.repo.CreateModerationCase(ctx, &postgres.ModerationCaseRow{
		Source: "manual", TargetType: msg.TargetType, TargetID: msg.TargetId,
		Reason: msg.Reason, Priority: "normal", Status: "actioned", Notes: &msg.Notes,
	})
	caseID := ""
	if c != nil {
		caseID = c.ID
	}
	return connect.NewResponse(&alfqv1.ModerateContentResponse{Success: true, CaseId: caseID}), nil
}

func (h *AdminSocialHandler) ListReports(ctx context.Context, req *connect.Request[alfqv1.AdminListReportsRequest]) (*connect.Response[alfqv1.AdminListReportsResponse], error) {
	msg := req.Msg
	limit := msg.PageSize
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var c *postgres.SocialCursor
	if msg.Cursor != "" {
		dec, _ := postgres.DecodeSocialCursor(msg.Cursor)
		c = dec
	}
	rows, total, next, err := h.repo.ListReportsAdmin(ctx, msg.Status, msg.Priority, msg.TargetType, msg.AssigneeId, c, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	reports := make([]*alfqv1.AdminReportSummary, len(rows))
	for i, r := range rows {
		reports[i] = &alfqv1.AdminReportSummary{
			ReportId:   r.ID,
			TargetType: r.TargetType,
			TargetId:   r.TargetID,
			ReporterId: strPtr(r.ReporterID),
			Reason:     r.Reason,
			Priority:   r.Priority,
			Status:     r.Status,
			AssigneeId: strPtr(r.AssigneeID),
			CreatedAt:  r.CreatedAt.Unix(),
		}
	}
	return connect.NewResponse(&alfqv1.AdminListReportsResponse{Reports: reports, NextCursor: postgres.EncodeSocialCursor(next), TotalCount: total}), nil
}

func strPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *AdminSocialHandler) GetReportDetail(ctx context.Context, req *connect.Request[alfqv1.AdminGetReportDetailRequest]) (*connect.Response[alfqv1.AdminReportDetail], error) {
	r, err := h.repo.GetReportDetail(ctx, req.Msg.ReportId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	return connect.NewResponse(&alfqv1.AdminReportDetail{
		Report: &alfqv1.AdminReportSummary{
			ReportId: r.ID, TargetType: r.TargetType, TargetId: r.TargetID,
			ReporterId: strPtr(r.ReporterID), Reason: r.Reason, Priority: r.Priority,
			Status: r.Status, AssigneeId: strPtr(r.AssigneeID), CreatedAt: r.CreatedAt.Unix(),
		},
	}), nil
}

func (h *AdminSocialHandler) ReviewReport(ctx context.Context, req *connect.Request[alfqv1.ReviewReportRequest]) (*connect.Response[alfqv1.ReviewReportResponse], error) {
	msg := req.Msg
	if msg.Reason == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("reason is required"))
	}
	if err := h.repo.UpdateReportStatus(ctx, msg.ReportId, msg.Action, msg.AssigneeId, msg.AdminId); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&alfqv1.ReviewReportResponse{Success: true}), nil
}

func (h *AdminSocialHandler) AssignReport(ctx context.Context, req *connect.Request[alfqv1.AssignReportRequest]) (*connect.Response[alfqv1.AssignReportResponse], error) {
	if err := h.repo.UpdateReportStatus(ctx, req.Msg.ReportId, "", req.Msg.AssigneeId, ""); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&alfqv1.AssignReportResponse{Success: true}), nil
}

func (h *AdminSocialHandler) ListModerationCases(ctx context.Context, req *connect.Request[alfqv1.AdminListModerationCasesRequest]) (*connect.Response[alfqv1.AdminListModerationCasesResponse], error) {
	msg := req.Msg
	limit := msg.PageSize
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	var c *postgres.SocialCursor
	if msg.Cursor != "" {
		dec, _ := postgres.DecodeSocialCursor(msg.Cursor)
		c = dec
	}
	rows, _, next, err := h.repo.ListReportsAdmin(ctx, msg.Status, "", msg.TargetType, msg.AssigneeId, c, limit)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	cases := make([]*alfqv1.ModerationCaseSummary, len(rows))
	for i, r := range rows {
		cases[i] = &alfqv1.ModerationCaseSummary{
			CaseId:     r.ID,
			Source:     r.Source,
			TargetType: r.TargetType,
			TargetId:   r.TargetID,
			ReporterId: strPtr(r.ReporterID),
			Reason:     r.Reason,
			Priority:   r.Priority,
			Status:     r.Status,
			AssigneeId: strPtr(r.AssigneeID),
			CreatedAt:  r.CreatedAt.Unix(),
		}
	}
	return connect.NewResponse(&alfqv1.AdminListModerationCasesResponse{Cases: cases, NextCursor: postgres.EncodeSocialCursor(next)}), nil
}

var _ antclawv1connect.AdminSocialServiceHandler = (*AdminSocialHandler)(nil)
