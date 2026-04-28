package ai

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	aiv1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
	"github.com/antclaw/antclaw/internal/service/systemai"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service implements AI-powered analysis business logic.
type Service struct {
	sysAI    *systemai.Service
	pool     *pgxpool.Pool // 可选：用于读取 ai_cache 与 BuildContext 数据源
	sessions map[string]*ChatSession
}

// ChatSession stores conversation history.
type ChatSession struct {
	Messages []*aiv1.ChatMessage
	Created  time.Time
}

// NewService 创建 AIService。pool 可为 nil，此时缓存层与 BuildContext 数据查询会被跳过。
func NewService(sysAI *systemai.Service, pool *pgxpool.Pool) *Service {
	return &Service{
		sysAI:    sysAI,
		pool:     pool,
		sessions: make(map[string]*ChatSession),
	}
}

// fingerprint 计算 prompt 的 SHA-256 指纹，作为 ai_cache 主键。
func fingerprint(operation, prompt string) string {
	h := sha256.Sum256([]byte(operation + "|" + prompt))
	return hex.EncodeToString(h[:])
}

// cacheLookup 在 ai_cache 表内按指纹+未过期取结果；未命中返回空串。
func (s *Service) cacheLookup(ctx context.Context, fp string) string {
	if s.pool == nil {
		return ""
	}
	var result string
	err := s.pool.QueryRow(ctx, `
		SELECT result FROM ai_cache
		 WHERE fingerprint = $1 AND (expires_at IS NULL OR expires_at > NOW())
		 LIMIT 1`, fp).Scan(&result)
	if err != nil {
		return ""
	}
	return result
}

// cacheStore 将结果写入 ai_cache，TTL 默认 1 小时。
func (s *Service) cacheStore(ctx context.Context, fp, op, model, result string, ttl time.Duration) {
	if s.pool == nil {
		return
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	_, _ = s.pool.Exec(ctx, `
		INSERT INTO ai_cache(fingerprint, operation, model, result, created_at, expires_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW() + ($5 || ' seconds')::INTERVAL)
		ON CONFLICT (fingerprint) DO UPDATE
		SET result = EXCLUDED.result, created_at = EXCLUDED.created_at, expires_at = EXCLUDED.expires_at`,
		fp, op, model, result, fmt.Sprintf("%d", int(ttl.Seconds())))
}

// Interpret provides AI interpretation of market data。
// 流程：构造 prompt → 命中 ai_cache 直接返回（cache_hit=true）→ 否则调 LLM 并写入缓存。
func (s *Service) Interpret(ctx context.Context, dataType, rawData, question, locale string) (*aiv1.InterpretResponse, error) {
	if locale == "" {
		locale = "en"
	}

	prompt := buildInterpretPrompt(dataType, rawData, question, locale)
	fp := fingerprint("interpret", prompt)
	if cached := s.cacheLookup(ctx, fp); cached != "" {
		return &aiv1.InterpretResponse{
			Interpretation: cached,
			KeyPoints:      extractKeyPoints(cached),
			Confidence:     0.78,
			CacheHit:       true,
		}, nil
	}

	content, err := s.callAI(ctx, prompt, "analysis")
	if err != nil {
		return nil, fmt.Errorf("ai interpret: %w", err)
	}
	s.cacheStore(ctx, fp, "interpret", "", content, time.Hour)

	return &aiv1.InterpretResponse{
		Interpretation: content,
		KeyPoints:      extractKeyPoints(content),
		Confidence:     0.78,
		CacheHit:       false,
	}, nil
}

// Outlook provides AI-generated market outlook.
func (s *Service) Outlook(ctx context.Context, pair, timeframe, locale string) (*aiv1.OutlookResponse, error) {
	if locale == "" {
		locale = "en"
	}
	if timeframe == "" {
		timeframe = "1d"
	}

	prompt := buildOutlookPrompt(pair, timeframe, locale)
	fp := fingerprint("outlook", prompt)
	if cached := s.cacheLookup(ctx, fp); cached != "" {
		return outlookFromContent(pair, cached), nil
	}
	content, err := s.callAI(ctx, prompt, "analysis")
	if err != nil {
		return nil, fmt.Errorf("ai outlook: %w", err)
	}
	s.cacheStore(ctx, fp, "outlook", "", content, 30*time.Minute)
	return outlookFromContent(pair, content), nil
}

// outlookFromContent 把 LLM 文本拆解为各字段；保持与原同步行为一致。
func outlookFromContent(pair, content string) *aiv1.OutlookResponse {
	return &aiv1.OutlookResponse{
		Pair:        pair,
		Summary:     extractSummary(content),
		BullishCase: extractBullish(content),
		BearishCase: extractBearish(content),
		KeyLevels:   extractKeyLevels(content),
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

// GetSession retrieves or creates a chat session.
func (s *Service) GetSession(sessionID string) *ChatSession {
	session, ok := s.sessions[sessionID]
	if !ok {
		session = &ChatSession{
			Messages: []*aiv1.ChatMessage{},
			Created:  time.Now(),
		}
		s.sessions[sessionID] = session
	}
	return session
}

// AddMessage adds a message to a session.
func (s *Service) AddMessage(sessionID string, role, content string) *aiv1.ChatMessage {
	session := s.GetSession(sessionID)
	msg := &aiv1.ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
	session.Messages = append(session.Messages, msg)
	return msg
}

// callAI 调用配置的 AI 提供商（OpenAI 兼容格式）
func (s *Service) callAI(ctx context.Context, prompt, purpose string) (string, error) {
	// 1. 获取配置
	cfg, secret, err := s.sysAI.GetPrimaryForPurpose(ctx, purpose)
	if err != nil {
		return "", fmt.Errorf("no ai config for purpose %q: %w", purpose, err)
	}
	if cfg == nil || secret == "" {
		return "", fmt.Errorf("ai provider not configured or missing secret for purpose %q", purpose)
	}
	if cfg.BaseURL == "" {
		return "", fmt.Errorf("ai provider base_url is empty")
	}

	// 2. 构造请求
	model := cfg.DefaultModel
	if model == "" && len(cfg.Models) > 0 {
		model = cfg.Models[0]
	}
	if model == "" {
		model = "gpt-4"
	}

	reqBody := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a professional financial analyst. Provide concise, data-driven analysis."},
			{"role": "user", "content": prompt},
		},
		"temperature": cfg.Temperature,
	}
	if cfg.MaxTokens > 0 {
		reqBody["max_tokens"] = cfg.MaxTokens
	}

	body, _ := json.Marshal(reqBody)

	// 3. 发送请求
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if !strings.HasSuffix(baseURL, "/v1") && !strings.HasSuffix(baseURL, "/v1/") {
		baseURL = baseURL + "/v1"
	}
	url := baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+secret)
	if cfg.Organization != "" {
		req.Header.Set("OpenAI-Organization", cfg.Organization)
	}

	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("ai api error %d: %s", resp.StatusCode, string(respBody))
	}

	// 4. 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode ai response: %w", err)
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty ai response")
	}

	return result.Choices[0].Message.Content, nil
}

// buildInterpretPrompt 构建数据解读提示词
func buildInterpretPrompt(dataType, rawData, question, locale string) string {
	if question == "" {
		question = "What does this data indicate?"
	}
	return fmt.Sprintf("Analyze the following %s data and answer: %s\n\nData: %s\n\nProvide your analysis in %s.", dataType, question, rawData, locale)
}

// buildOutlookPrompt 构建市场展望提示词
func buildOutlookPrompt(pair, timeframe, locale string) string {
	return fmt.Sprintf("Provide a technical outlook for %s on %s timeframe. Include: summary, bullish case, bearish case, and key support/resistance levels. Respond in %s.", pair, timeframe, locale)
}

func extractKeyPoints(content string) []string {
	sentences := strings.Split(content, ". ")
	var points []string
	for i, s := range sentences {
		if i >= 3 {
			break
		}
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			points = append(points, trimmed)
		}
	}
	return points
}

func extractSummary(content string) string {
	if idx := strings.Index(content, "\n"); idx > 0 {
		return strings.TrimSpace(content[:idx])
	}
	return content
}

func extractBullish(content string) string {
	if idx := strings.Index(strings.ToLower(content), "bullish"); idx >= 0 {
		end := strings.Index(content[idx:], "\n")
		if end < 0 {
			return content[idx:]
		}
		return content[idx : idx+end]
	}
	return ""
}

func extractBearish(content string) string {
	if idx := strings.Index(strings.ToLower(content), "bearish"); idx >= 0 {
		end := strings.Index(content[idx:], "\n")
		if end < 0 {
			return content[idx:]
		}
		return content[idx : idx+end]
	}
	return ""
}

func extractKeyLevels(content string) string {
	if idx := strings.Index(strings.ToLower(content), "key level"); idx >= 0 {
		end := strings.Index(content[idx:], "\n")
		if end < 0 {
			return content[idx:]
		}
		return content[idx : idx+end]
	}
	return ""
}
