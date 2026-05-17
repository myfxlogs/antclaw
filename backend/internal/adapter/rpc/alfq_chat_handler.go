package rpc

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	alfqv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/gen/go/antclaw/v1/antclawv1connect"
	"github.com/antclaw/antclaw/internal/infra/postgres"
)

type ChatHandler struct{ pool *pgxpool.Pool }

func NewChatHandler(pool *pgxpool.Pool) *ChatHandler { return &ChatHandler{pool: pool} }

func (h *ChatHandler) SendMessage(ctx context.Context, req *connect.Request[alfqv1.SendMessageRequest]) (*connect.Response[alfqv1.Message], error) {
	uid := userIDFromHTTP(ctx, req)
	var name string
	h.pool.QueryRow(ctx, "SELECT "+postgres.PublicDisplayNameExpr+" FROM users WHERE id=$1", uid).Scan(&name)
	r := req.Msg
	m := &alfqv1.Message{SenderId: uid, SenderName: name, Content: r.Content, MessageType: r.MessageType, SignalData: r.SignalData, ConversationId: r.ConversationId}
	var ts time.Time
	h.pool.QueryRow(ctx, `
		INSERT INTO alfq_messages (conversation_id, sender_id, sender_name, content, message_type, signal_data)
		VALUES ($1,$2,$3,$4,$5,$6) RETURNING id, created_at
	`, r.ConversationId, uid, name, r.Content, r.MessageType, r.SignalData).Scan(&m.Id, &ts)
	m.CreatedAt = ts.Unix()
	return connect.NewResponse(m), nil
}

func (h *ChatHandler) GetConversation(ctx context.Context, req *connect.Request[alfqv1.GetConversationRequest]) (*connect.Response[alfqv1.Conversation], error) {
	r := req.Msg
	rows, _ := h.pool.Query(ctx, `
		SELECT id, conversation_id, sender_id, sender_name, content, message_type, signal_data, created_at
		FROM alfq_messages WHERE conversation_id=$1 ORDER BY created_at DESC LIMIT $2
	`, r.ConversationId, r.PageSize)
	defer rows.Close()
	var msgs []*alfqv1.Message
	for rows.Next() {
		m := &alfqv1.Message{}
		var ts time.Time
		rows.Scan(&m.Id, &m.ConversationId, &m.SenderId, &m.SenderName, &m.Content, &m.MessageType, &m.SignalData, &ts)
		m.CreatedAt = ts.Unix()
		msgs = append(msgs, m)
	}
	conv := &alfqv1.Conversation{Id: r.ConversationId}
	if len(msgs) > 0 {
		conv.LastMessage = msgs[0].Content
		conv.LastMessageAt = msgs[0].CreatedAt
	}
	return connect.NewResponse(conv), nil
}

func (h *ChatHandler) ListConversations(ctx context.Context, req *connect.Request[alfqv1.ListConversationsRequest]) (*connect.Response[alfqv1.ConversationList], error) {
	uid := userIDFromHTTP(ctx, req)
	rows, _ := h.pool.Query(ctx, `
		SELECT DISTINCT ON (conversation_id) conversation_id, content, created_at
		FROM alfq_messages
		WHERE conversation_id IN (SELECT conversation_id FROM alfq_messages WHERE sender_id=$1)
		   OR conversation_id IN (SELECT conversation_id FROM alfq_messages WHERE sender_id=$1)
		ORDER BY conversation_id, created_at DESC
	`, uid)
	defer rows.Close()
	var convs []*alfqv1.Conversation
	for rows.Next() {
		c := &alfqv1.Conversation{}
		var ts time.Time
		rows.Scan(&c.Id, &c.LastMessage, &ts)
		c.LastMessageAt = ts.Unix()
		convs = append(convs, c)
	}
	return connect.NewResponse(&alfqv1.ConversationList{Conversations: convs}), nil
}

func (h *ChatHandler) MarkRead(ctx context.Context, req *connect.Request[alfqv1.ChatMarkReadRequest]) (*connect.Response[alfqv1.ChatMarkReadResponse], error) {
	return connect.NewResponse(&alfqv1.ChatMarkReadResponse{}), nil
}

var _ antclawv1connect.ChatServiceHandler = (*ChatHandler)(nil)
