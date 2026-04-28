package firecrawl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 Firecrawl /v1/scrape JSON 抽取能力。
// 用于反爬/无公开 API 的页面（MyFXBook、Finviz、FedWatch、TradingEconomics 等）。
// 需要 FIRECRAWL_API_KEY 环境变量；未设置时 IsAvailable() 返回 false，所有 Scrape 调用直接返回错误，
// 由调用方决定是否优雅降级。
type Client struct {
	src    apiclient.Source
	base   string
	apiKey string
}

func NewClient(src apiclient.Source) *Client {
	return &Client{
		src:    src,
		base:   "https://api.firecrawl.dev/v1/scrape",
		apiKey: os.Getenv("FIRECRAWL_API_KEY"),
	}
}

// IsAvailable 报告是否已配置 API Key。调用方可据此跳过抓取分支。
func (c *Client) IsAvailable() bool { return c.apiKey != "" }

// ScrapeJSON 使用 prompt + schema 对页面做结构化抽取，把抽取出的 JSON 写入 out。
// out 必须是指针；schema 以 JSON 字节串形式提供（参见 firecrawl 官方协议）。
func (c *Client) ScrapeJSON(ctx context.Context, pageURL, prompt string, schema json.RawMessage, waitForMs int, out any) error {
	if c.apiKey == "" {
		return fmt.Errorf("firecrawl: FIRECRAWL_API_KEY not set")
	}
	type jsonOpts struct {
		Prompt string          `json:"prompt"`
		Schema json.RawMessage `json:"schema"`
	}
	body, err := json.Marshal(struct {
		URL         string    `json:"url"`
		Formats     []string  `json:"formats"`
		WaitFor     int       `json:"waitFor,omitempty"`
		JSONOptions *jsonOpts `json:"jsonOptions,omitempty"`
	}{
		URL:         pageURL,
		Formats:     []string{"json"},
		WaitFor:     waitForMs,
		JSONOptions: &jsonOpts{Prompt: prompt, Schema: schema},
	})
	if err != nil {
		return fmt.Errorf("firecrawl: marshal: %w", err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.base, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("firecrawl: do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("firecrawl: http %d", resp.StatusCode)
	}
	// Firecrawl 响应：{ "success":bool, "data": { "json": <user-defined> } }
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			JSON json.RawMessage `json:"json"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("firecrawl: decode envelope: %w", err)
	}
	if !envelope.Success || len(envelope.Data.JSON) == 0 {
		return fmt.Errorf("firecrawl: scrape unsuccessful or empty payload")
	}
	if err := json.Unmarshal(envelope.Data.JSON, out); err != nil {
		return fmt.Errorf("firecrawl: decode json: %w", err)
	}
	return nil
}
