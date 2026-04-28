package firecrawl

import (
	"context"
	"encoding/json"
	"time"
)

// FedWatchSnapshot CME FedWatch 隐含利率概率快照。
type FedWatchSnapshot struct {
	NextMeetingDate    string
	HoldProbability    float64
	Cut25Probability   float64
	Cut50Probability   float64
	Hike25Probability  float64
	ImpliedYearEndRate float64
	MeetingCount       int
	FetchedAt          time.Time
}

// FetchFedWatch 抓取 cmegroup.com/markets/interest-rates/cme-fedwatch-tool.html 关键概率字段。
func (c *Client) FetchFedWatch(ctx context.Context) (*FedWatchSnapshot, error) {
	schema := json.RawMessage(`{
        "type":"object",
        "properties":{
          "next_meeting_date":{"type":"string"},
          "hold_probability":{"type":"number"},
          "cut_25_probability":{"type":"number"},
          "cut_50_probability":{"type":"number"},
          "hike_25_probability":{"type":"number"},
          "implied_year_end_rate":{"type":"number"},
          "meeting_count":{"type":"integer"}
        }
    }`)
	prompt := "From CME FedWatch tool, extract for the next FOMC meeting: meeting date (YYYY-MM-DD), hold probability %, 25bp cut probability %, 50bp cut probability %, 25bp hike probability %, year-end implied rate (bps from December futures), remaining meeting count this year."
	var raw struct {
		NextMeetingDate    string  `json:"next_meeting_date"`
		HoldProbability    float64 `json:"hold_probability"`
		Cut25Probability   float64 `json:"cut_25_probability"`
		Cut50Probability   float64 `json:"cut_50_probability"`
		Hike25Probability  float64 `json:"hike_25_probability"`
		ImpliedYearEndRate float64 `json:"implied_year_end_rate"`
		MeetingCount       int     `json:"meeting_count"`
	}
	url := "https://www.cmegroup.com/markets/interest-rates/cme-fedwatch-tool.html"
	if err := c.ScrapeJSON(ctx, url, prompt, schema, 8000, &raw); err != nil {
		return nil, err
	}
	return &FedWatchSnapshot{
		NextMeetingDate:    raw.NextMeetingDate,
		HoldProbability:    raw.HoldProbability,
		Cut25Probability:   raw.Cut25Probability,
		Cut50Probability:   raw.Cut50Probability,
		Hike25Probability:  raw.Hike25Probability,
		ImpliedYearEndRate: raw.ImpliedYearEndRate,
		MeetingCount:       raw.MeetingCount,
		FetchedAt:          time.Now().UTC(),
	}, nil
}
