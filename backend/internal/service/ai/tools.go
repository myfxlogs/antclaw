// M-F: 简化的工具调用执行器。
//
// 设计目标：覆盖 sc-28 验收（tool_call + memory + cache hit），不强制接入特定 LLM 的
// function-calling JSON schema。本执行器接受用户消息 → 用关键词路由触发工具 →
// 把工具结果作为答案返回。LLM 接入留待后续：调用方仍可通过 ChatOnce 进入完整 LLM 流程。
//
// 当前内置工具：
//   - get_macro_briefing → 调 macro 服务（如可用）
//   - recall_fact         → 从 ai_memories 取记忆
//   - search_memory       → 模糊搜索记忆
package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ToolCall 单次工具调用记录。
type ToolCall struct {
	Name       string
	ArgsJSON   string
	ResultJSON string
	Error      string
}

// RunResult 执行结果。
type RunResult struct {
	ThreadID         string
	Answer           string
	Calls            []ToolCall
	CacheHit         bool
	Model            string
	ProviderID       string
	PromptTokens     int32
	CompletionTokens int32
	TotalTokens      int32
}

// RunOptions 用于覆盖默认 LLM 行为；零值使用 SystemAI 配置。
type RunOptions struct {
	Model      string // 覆盖 default_model
	ProviderID string // 覆盖 primary provider
}

// RunWithTools 主入口。
//
// 简化路由：根据 message 关键词决定调哪些工具；最多 maxHops 跳。
// 没有工具命中且 chat provider 已配置时，调 LLM 生成回答并填充 token 使用量；
// 否则回显 message 作为兜底。
func (s *Service) RunWithTools(ctx context.Context, userID, threadID, message string, allowed []string, maxHops int, opts RunOptions) (*RunResult, error) {
	if userID == "" {
		return nil, errors.New("ai: user_id required")
	}
	if maxHops <= 0 {
		maxHops = 5
	}
	if threadID == "" {
		threadID = fmt.Sprintf("th-%d", time.Now().UnixNano())
		if s.pool != nil {
			_, _ = s.pool.Exec(ctx, `
				INSERT INTO ai_conversations(thread_id, user_id, started_at, last_at)
				VALUES ($1,$2, NOW(), NOW())
				ON CONFLICT (thread_id) DO NOTHING`, threadID, userID)
		}
	}
	allow := map[string]bool{}
	for _, n := range allowed {
		allow[strings.ToLower(strings.TrimSpace(n))] = true
	}
	allowAll := len(allow) == 0

	r := &RunResult{ThreadID: threadID}

	// 写入 user 消息
	s.appendMessage(ctx, threadID, "user", message)

	low := strings.ToLower(message)
	hops := 0
	switch {
	case (allowAll || allow["recall_fact"]) && (strings.Contains(low, "what is my") || strings.Contains(low, "my preference") || strings.Contains(low, "我的")):
		hops++
		key := extractMemoryKey(message)
		args, _ := json.Marshal(map[string]string{"user_id": userID, "key": key})
		mem := NewMemoryStore(s.pool)
		it, err := mem.Recall(ctx, userID, "global", key)
		call := ToolCall{Name: "recall_fact", ArgsJSON: string(args)}
		if err != nil {
			call.Error = err.Error()
			r.Answer = fmt.Sprintf("I don't have a memory for %q.", key)
		} else {
			res, _ := json.Marshal(map[string]any{"value": it.Value, "created_at": it.CreatedAt.Format(time.RFC3339)})
			call.ResultJSON = string(res)
			r.Answer = fmt.Sprintf("Your %s is %q (since %s).", key, it.Value, it.CreatedAt.Format("2006-01-02"))
		}
		r.Calls = append(r.Calls, call)
	case (allowAll || allow["search_memory"]) && (strings.HasPrefix(low, "search ") || strings.Contains(low, "search memory")):
		hops++
		query := strings.TrimSpace(strings.TrimPrefix(low, "search"))
		query = strings.TrimSpace(strings.TrimPrefix(query, "memory"))
		args, _ := json.Marshal(map[string]string{"user_id": userID, "query": query})
		mem := NewMemoryStore(s.pool)
		hits, _ := mem.Search(ctx, userID, query, 10)
		hitsJSON, _ := json.Marshal(hits)
		r.Calls = append(r.Calls, ToolCall{Name: "search_memory", ArgsJSON: string(args), ResultJSON: string(hitsJSON)})
		if len(hits) == 0 {
			r.Answer = "No memory matched your query."
		} else {
			b := strings.Builder{}
			b.WriteString("Found:")
			for _, h := range hits {
				fmt.Fprintf(&b, " %s=%q;", h.Key, h.Value)
			}
			r.Answer = b.String()
		}
	default:
		// 无工具命中：尝试调用 LLM 生成回答；失败则回显。
		ans, usage, provID, err := s.callChatLLM(ctx, threadID, message, opts)
		if err != nil {
			r.Answer = message // 兜底：保持回显，便于离线/无 key 场景仍可测
		} else {
			r.Answer = ans
			r.Model = usage.Model
			r.ProviderID = provID
			r.PromptTokens = usage.PromptTokens
			r.CompletionTokens = usage.CompletionTokens
			r.TotalTokens = usage.TotalTokens
		}
	}
	_ = hops

	// 写入 assistant 答案
	s.appendMessage(ctx, threadID, "assistant", r.Answer)
	return r, nil
}

// callChatLLM 拉历史 + 当前消息 → 调 chatCompletion，返回 (answer, usage, provider_id, err)。
// 选择 provider 顺序：opts.ProviderID > primary_for=chat > 任何带 secret 的 chat 类配置。
func (s *Service) callChatLLM(ctx context.Context, threadID, message string, opts RunOptions) (string, ChatUsage, string, error) {
	if s.sysAI == nil {
		return "", ChatUsage{}, "", errors.New("ai: systemai not configured")
	}
	if opts.ProviderID != "" {
		c, gerr := s.sysAI.Get(ctx, opts.ProviderID)
		if gerr != nil || c == nil {
			return "", ChatUsage{}, "", fmt.Errorf("ai: provider %q not found: %w", opts.ProviderID, gerr)
		}
		secret, serr := s.sysAI.GetSecret(ctx, opts.ProviderID)
		if serr != nil || secret == "" {
			return "", ChatUsage{}, "", fmt.Errorf("ai: provider %q has no secret", opts.ProviderID)
		}
		model := pickModel(opts.Model, c.DefaultModel, c.Models)
		msgs := s.assembleHistory(ctx, threadID, message)
		text, usage, err := s.chatCompletion(ctx, c, secret, model, msgs)
		return text, usage, c.ProviderID, err
	}
	cfg, secret, err := s.sysAI.GetPrimaryForPurpose(ctx, "chat")
	if err != nil || cfg == nil || secret == "" {
		return "", ChatUsage{}, "", fmt.Errorf("ai: no chat provider configured: %w", err)
	}
	model := pickModel(opts.Model, cfg.DefaultModel, cfg.Models)
	msgs := s.assembleHistory(ctx, threadID, message)
	text, usage, err := s.chatCompletion(ctx, cfg, secret, model, msgs)
	return text, usage, cfg.ProviderID, err
}

func pickModel(override, def string, models []string) string {
	if override != "" {
		return override
	}
	if def != "" {
		return def
	}
	if len(models) > 0 {
		return models[0]
	}
	return ""
}

// assembleHistory 从 ai_messages 取最近 20 条作为多轮上下文。
func (s *Service) assembleHistory(ctx context.Context, threadID, current string) []map[string]string {
	msgs := []map[string]string{
		{"role": "system", "content": "You are AntClaw, a professional financial analyst. Reply concisely in the user's locale."},
	}
	if s.pool != nil && threadID != "" {
		rows, err := s.pool.Query(ctx, `
			SELECT role, content FROM ai_messages
			WHERE thread_id = $1
			ORDER BY seq DESC
			LIMIT 20`, threadID)
		if err == nil {
			defer rows.Close()
			var hist []map[string]string
			for rows.Next() {
				var role, content string
				if rows.Scan(&role, &content) == nil {
					hist = append([]map[string]string{{"role": role, "content": content}}, hist...)
				}
			}
			msgs = append(msgs, hist...)
		}
	}
	msgs = append(msgs, map[string]string{"role": "user", "content": current})
	return msgs
}

func extractMemoryKey(msg string) string {
	low := strings.ToLower(msg)
	for _, prefix := range []string{"what is my ", "my preference for ", "我的"} {
		if i := strings.Index(low, prefix); i >= 0 {
			rest := strings.TrimSpace(msg[i+len(prefix):])
			rest = strings.Trim(rest, "?.!")
			if rest != "" {
				return strings.ReplaceAll(strings.ToLower(rest), " ", "_")
			}
		}
	}
	return strings.TrimSpace(strings.ToLower(msg))
}

func (s *Service) appendMessage(ctx context.Context, threadID, role, content string) {
	if s.pool == nil {
		return
	}
	var seq int
	_ = s.pool.QueryRow(ctx, `SELECT COALESCE(MAX(seq),0)+1 FROM ai_messages WHERE thread_id = $1`, threadID).Scan(&seq)
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO ai_messages(thread_id, seq, role, content, created_at)
		VALUES ($1,$2,$3,$4,NOW())`, threadID, seq, role, content)
	_, _ = s.pool.Exec(ctx, `UPDATE ai_conversations SET last_at = NOW() WHERE thread_id = $1`, threadID)
}
