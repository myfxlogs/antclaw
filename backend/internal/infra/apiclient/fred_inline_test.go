package apiclient

import "testing"

// TestFredClient_SetBaseURL_Normalize 验证 SetBaseURL 对常见错填的兜底归一化。
func TestFredClient_SetBaseURL_Normalize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"missing /fred", "https://api.stlouisfed.org", "https://api.stlouisfed.org/fred"},
		{"missing /fred + trailing slash", "https://api.stlouisfed.org/", "https://api.stlouisfed.org/fred"},
		{"already with /fred", "https://api.stlouisfed.org/fred", "https://api.stlouisfed.org/fred"},
		{"with /fred and trailing slash", "https://api.stlouisfed.org/fred/", "https://api.stlouisfed.org/fred"},
		{"v2 path preserved", "https://api.stlouisfed.org/fred/v2", "https://api.stlouisfed.org/fred/v2"},
		{"empty ignored", "", ""}, // 由内部 default 兜住，不被覆盖
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewFredClient("")
			c.SetBaseURL(tc.in)
			got := c.getBaseURL()
			// 空入参时使用 default
			expect := tc.want
			if tc.in == "" {
				expect = fredDefaultBaseURL
			}
			if got != expect {
				t.Fatalf("in=%q got=%q want=%q", tc.in, got, expect)
			}
		})
	}
}
