// Package systemai provides system-level AI configuration management.
package systemai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	cryptopkg "github.com/antclaw/antclaw/internal/crypto"
)

const maxDiscoverPages = 10

// Config represents a system AI provider configuration.
type Config struct {
	ProviderID     string    `json:"provider_id"`
	Name           string    `json:"name"`
	BaseURL        string    `json:"base_url"`
	Organization   string    `json:"organization"`
	Models         []string  `json:"models"`
	DefaultModel   string    `json:"default_model"`
	Temperature    float64   `json:"temperature"`
	TimeoutSeconds int       `json:"timeout_seconds"`
	MaxTokens      int       `json:"max_tokens"`
	Purposes       []string  `json:"purposes"`
	PrimaryFor     []string  `json:"primary_for"`
	Enabled        bool      `json:"enabled"`
	HasSecret      bool      `json:"has_secret"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UpdatedBy      string    `json:"updated_by"`
}

// Service provides CRUD for system AI configs.
type Service struct {
	pool      *pgxpool.Pool
	secretBox *cryptopkg.SecretBox
}

// NewService creates a new system AI service.
func NewService(pool *pgxpool.Pool, box *cryptopkg.SecretBox) *Service {
	return &Service{pool: pool, secretBox: box}
}

// List returns all configs.
func (s *Service) List(ctx context.Context) ([]Config, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT provider_id, name, base_url, organization, models, default_model,
		       temperature, timeout_seconds, max_tokens, purposes, primary_for,
		       enabled, has_secret, created_at, updated_at, updated_by
		FROM system_ai_configs ORDER BY provider_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []Config
	for rows.Next() {
		var c Config
		err := rows.Scan(&c.ProviderID, &c.Name, &c.BaseURL, &c.Organization, &c.Models, &c.DefaultModel,
			&c.Temperature, &c.TimeoutSeconds, &c.MaxTokens, &c.Purposes, &c.PrimaryFor,
			&c.Enabled, &c.HasSecret, &c.CreatedAt, &c.UpdatedAt, &c.UpdatedBy)
		if err != nil {
			continue
		}
		configs = append(configs, c)
	}
	return configs, nil
}

// Get returns a single config.
func (s *Service) Get(ctx context.Context, providerID string) (*Config, error) {
	var c Config
	err := s.pool.QueryRow(ctx, `
		SELECT provider_id, name, base_url, organization, models, default_model,
		       temperature, timeout_seconds, max_tokens, purposes, primary_for,
		       enabled, has_secret, created_at, updated_at, updated_by
		FROM system_ai_configs WHERE provider_id = $1`, providerID).Scan(
		&c.ProviderID, &c.Name, &c.BaseURL, &c.Organization, &c.Models, &c.DefaultModel,
		&c.Temperature, &c.TimeoutSeconds, &c.MaxTokens, &c.Purposes, &c.PrimaryFor,
		&c.Enabled, &c.HasSecret, &c.CreatedAt, &c.UpdatedAt, &c.UpdatedBy)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateConfig updates non-sensitive fields.
func (s *Service) UpdateConfig(ctx context.Context, c *Config, by string) error {
	models := c.Models
	if models == nil {
		models = []string{}
	}
	purposes := c.Purposes
	if purposes == nil {
		purposes = []string{}
	}
	primaryFor := c.PrimaryFor
	if primaryFor == nil {
		primaryFor = []string{}
	}

	_, err := s.pool.Exec(ctx, `
		UPDATE system_ai_configs SET
			name = $1, base_url = $2, organization = $3, models = $4, default_model = $5,
			temperature = $6, timeout_seconds = $7, max_tokens = $8, purposes = $9, primary_for = $10,
			enabled = $11, updated_at = NOW(), updated_by = $12
		WHERE provider_id = $13`,
		c.Name, c.BaseURL, c.Organization, models, c.DefaultModel,
		c.Temperature, c.TimeoutSeconds, c.MaxTokens, purposes, primaryFor,
		c.Enabled, by, c.ProviderID)
	return err
}

// UpdateSecret encrypts and stores the API key.
func (s *Service) UpdateSecret(ctx context.Context, providerID, secret, by string) error {
	if strings.TrimSpace(secret) == "" {
		_, err := s.pool.Exec(ctx, `
			UPDATE system_ai_configs SET
				secret_ciphertext = NULL, secret_salt = NULL, secret_nonce = NULL, has_secret = FALSE,
				updated_at = NOW(), updated_by = $1
			WHERE provider_id = $2`,
			by, providerID)
		return err
	}
	if s.secretBox == nil {
		return errors.New("secretBox not initialized")
	}

	ct, salt, nonce, err := s.secretBox.Seal([]byte(secret))
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, `
		UPDATE system_ai_configs SET
			secret_ciphertext = $1, secret_salt = $2, secret_nonce = $3, has_secret = TRUE,
			updated_at = NOW(), updated_by = $4
		WHERE provider_id = $5`,
		ct, salt, nonce, by, providerID)
	return err
}

// GetSecret decrypts and returns the API key.
func (s *Service) GetSecret(ctx context.Context, providerID string) (string, error) {
	if s.secretBox == nil {
		return "", errors.New("secretBox not initialized")
	}

	var ct, salt, nonce []byte
	err := s.pool.QueryRow(ctx, `
		SELECT secret_ciphertext, secret_salt, secret_nonce
		FROM system_ai_configs WHERE provider_id = $1 AND has_secret = TRUE`, providerID).Scan(&ct, &salt, &nonce)
	if err != nil {
		return "", err
	}

	plaintext, err := s.secretBox.Open(ct, salt, nonce)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// DiscoverModels retrieves model IDs from an OpenAI-compatible /models endpoint.
func (s *Service) DiscoverModels(ctx context.Context, providerID string) ([]string, error) {
	cfg, err := s.Get(ctx, providerID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("base_url is empty")
	}
	secret, _ := s.GetSecret(ctx, providerID)

	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")

	// Special-case: Zhipu GLM might use different paging/shape. Try a dedicated fetch first.
	if providerID == "zhipu" {
		if all, err := fetchZhipuModelsAll(ctx, base, secret); err == nil && len(all) > 0 {
			return all, nil
		}
	}

	ids := make([]string, 0, 64)
	seen := map[string]struct{}{}
	after := ""
	for page := 0; page < maxDiscoverPages; page++ {
		pageIDs, nextAfter, hasMore, err := fetchModelsPage(ctx, base, secret, after)
		if err != nil {
			return nil, err
		}
		for _, id := range pageIDs {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
		if !hasMore || nextAfter == "" {
			break
		}
		after = nextAfter
	}
	if len(ids) == 0 {
		return nil, errors.New("no models returned by upstream /models")
	}
	return ids, nil
}

func fetchModelsPage(ctx context.Context, baseURL, secret, after string) ([]string, string, bool, error) {
	values := neturl.Values{}
	values.Set("limit", "100")
	// Be liberal in what we accept: some providers use page_size or size
	values.Set("page_size", "100")
	values.Set("size", "100")
	if after != "" {
		values.Set("after", after)
	}
	modelsURL := baseURL + "/models"
	if encoded := values.Encode(); encoded != "" {
		modelsURL += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsURL, nil)
	if err != nil {
		return nil, "", false, err
	}
	if strings.TrimSpace(secret) != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		var urlErr *neturl.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return nil, "", false, errors.New("upstream timeout while requesting /models")
		}
		return nil, "", false, fmt.Errorf("upstream unreachable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, "", false, errors.New("upstream unauthorized: check api key/secret")
	}
	// 借鉴 anttrader：403/429 时读取响应体，根据厂商真实文案识别"免费档耗尽 / 配额耗尽 / 限流"
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		lower := strings.ToLower(strings.TrimSpace(string(body)))
		switch {
		case strings.Contains(lower, "freetieronly"), strings.Contains(lower, "free tier"), strings.Contains(lower, "free-tier"):
			return nil, "", false, fmt.Errorf("upstream free-tier exhausted: status %d %s", resp.StatusCode, lower)
		case strings.Contains(lower, "quota"), strings.Contains(lower, "exhaust"), strings.Contains(lower, "allocation"),
			strings.Contains(lower, "rate limit"), strings.Contains(lower, "too many requests"),
			resp.StatusCode == http.StatusTooManyRequests:
			return nil, "", false, fmt.Errorf("upstream quota exhausted: status %d %s", resp.StatusCode, lower)
		case resp.StatusCode == http.StatusForbidden:
			return nil, "", false, errors.New("upstream unauthorized: check api key/secret")
		}
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, "", false, errors.New("upstream endpoint not found: base_url must be openai-compatible and include /v1")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			msg = fmt.Sprintf("status %d", resp.StatusCode)
		}
		return nil, "", false, fmt.Errorf("upstream returned error: %s", msg)
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		HasMore bool   `json:"has_more"`
		Next    string `json:"next"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", false, errors.New("invalid /models response: cannot parse json")
	}
	ids := make([]string, 0, len(payload.Data))
	lastID := ""
	for _, m := range payload.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" {
			continue
		}
		ids = append(ids, id)
		lastID = id
	}
	if payload.Next != "" && payload.HasMore {
		return ids, payload.Next, true, nil
	}
	return ids, lastID, payload.HasMore, nil
}

// fetchZhipuModelsAll tries to retrieve all GLM models accounting for alternative response shapes.
func fetchZhipuModelsAll(ctx context.Context, baseURL, secret string) ([]string, error) {
	q := neturl.Values{}
	q.Set("limit", "1000")
	q.Set("page_size", "1000")
	q.Set("size", "1000")
	url := baseURL + "/models"
	if enc := q.Encode(); enc != "" {
		url += "?" + enc
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(secret) != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("upstream returned error: status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))

	// Shape A: { data: [ { id | model | model_id } ] }
	var shapeA struct {
		Data []struct {
			ID      string `json:"id"`
			Model   string `json:"model"`
			ModelID string `json:"model_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &shapeA); err == nil && len(shapeA.Data) > 0 {
		out := make([]string, 0, len(shapeA.Data))
		for _, m := range shapeA.Data {
			id := firstNonEmpty(m.ID, m.Model, m.ModelID)
			id = strings.TrimSpace(id)
			if id != "" {
				out = append(out, id)
			}
		}
		return dedup(out), nil
	}

	// Shape B: { data: { list: [ { id | model | model_id } ], has_more, total } }
	var shapeB struct {
		Data struct {
			List []struct {
				ID      string `json:"id"`
				Model   string `json:"model"`
				ModelID string `json:"model_id"`
			} `json:"list"`
			HasMore bool `json:"has_more"`
			Total   int  `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &shapeB); err == nil && len(shapeB.Data.List) > 0 {
		out := make([]string, 0, len(shapeB.Data.List))
		for _, m := range shapeB.Data.List {
			id := firstNonEmpty(m.ID, m.Model, m.ModelID)
			id = strings.TrimSpace(id)
			if id != "" {
				out = append(out, id)
			}
		}
		return dedup(out), nil
	}
	return nil, errors.New("invalid /models response: cannot parse json")
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func dedup(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func FriendlyDiscoverError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "base_url is empty"):
		return "请先填写 Base URL（模型服务地址）。"
	case strings.Contains(msg, "free-tier exhausted"):
		return "免费额度已耗尽：请在厂商控制台关闭「仅使用免费档」或更换付费 Key（错误来自厂商响应）。"
	case strings.Contains(msg, "quota exhausted"):
		return "配额受限或被限流：厂商已拒绝调用（quota/429）。请检查计费/配额或稍后重试。"
	case strings.Contains(msg, "unauthorized"):
		return "鉴权失败：当前地址返回 401/403，请检查 API Key/Secret 或确认该网关是否允许匿名访问。"
	case strings.Contains(msg, "endpoint not found"):
		return "模型端点不存在：请确认 Base URL 与服务协议匹配（部分服务需要 /v1）。"
	case strings.Contains(msg, "timeout"):
		return "请求超时：请检查网络连通性或稍后重试。"
	case strings.Contains(msg, "upstream unreachable"):
		return "无法连接到模型服务：请检查 Base URL、网络或网关。"
	case strings.Contains(msg, "invalid /models response"):
		return "模型服务返回格式不兼容 /models 协议。"
	case strings.Contains(msg, "no models returned"):
		return "模型服务未返回可用模型，请检查账号权限或服务配置。"
	default:
		return "拉取模型失败，请检查 Base URL 与密钥配置。"
	}
}

// GetPrimaryForPurpose returns the primary provider for a given purpose.
func (s *Service) GetPrimaryForPurpose(ctx context.Context, purpose string) (*Config, string, error) {
	var c Config
	var ct, salt, nonce []byte
	err := s.pool.QueryRow(ctx, `
		SELECT provider_id, name, base_url, organization, models, default_model,
		       temperature, timeout_seconds, max_tokens, purposes, primary_for,
		       enabled, has_secret, created_at, updated_at, updated_by,
		       secret_ciphertext, secret_salt, secret_nonce
		FROM system_ai_configs WHERE enabled = TRUE AND $1 = ANY(primary_for) LIMIT 1`, purpose).Scan(
		&c.ProviderID, &c.Name, &c.BaseURL, &c.Organization, &c.Models, &c.DefaultModel,
		&c.Temperature, &c.TimeoutSeconds, &c.MaxTokens, &c.Purposes, &c.PrimaryFor,
		&c.Enabled, &c.HasSecret, &c.CreatedAt, &c.UpdatedAt, &c.UpdatedBy,
		&ct, &salt, &nonce)
	if err != nil {
		return nil, "", err
	}

	var secret string
	if s.secretBox != nil && len(ct) > 0 {
		plaintext, err := s.secretBox.Open(ct, salt, nonce)
		if err == nil {
			secret = string(plaintext)
		}
	}
	return &c, secret, nil
}
