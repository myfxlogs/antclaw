package secedgar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/antclaw/antclaw/internal/infra/apiclient"
)

// Client 封装 SEC EDGAR 公开 JSON 接口（无需 API Key，需 User-Agent）。
type Client struct {
	src       apiclient.Source
	base      string
	userAgent string
}

func NewClient(src apiclient.Source, ua string) *Client {
	if ua == "" {
		ua = "AntClaw/1.0 contact@antclaw.local"
	}
	return &Client{src: src, base: "https://data.sec.gov", userAgent: ua}
}

// SubmissionsResp 节选自 EDGAR submissions API 的字段（数组并行存储）。
type SubmissionsResp struct {
	CIK   string `json:"cik"`
	Name  string `json:"name"`
	Filings struct {
		Recent struct {
			AccessionNumber []string `json:"accessionNumber"`
			FilingDate      []string `json:"filingDate"`
			Form            []string `json:"form"`
			PrimaryDocument []string `json:"primaryDocument"`
		} `json:"recent"`
	} `json:"filings"`
}

func (c *Client) GetSubmissions(ctx context.Context, cik string) (*SubmissionsResp, error) {
	url := fmt.Sprintf("%s/submissions/CIK%010s.json", c.base, cik)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", c.userAgent)
	resp, err := c.src.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sec edgar http %d", resp.StatusCode)
	}
	var out SubmissionsResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
