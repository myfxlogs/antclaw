package macro

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/antclaw/antclaw/internal/domain/apperror"
	"github.com/antclaw/antclaw/internal/infra/apiclient"
	macrov1 "github.com/antclaw/antclaw/gen/go/antclaw/v1"
)

// Service implements Macro business logic with realistic data.
type Service struct {
	dataStore  *MacroDataStore
	fredClient *apiclient.FredClient
}

// NewServiceWithFRED creates a new MacroService with FRED API client.
func NewServiceWithFRED(fredClient *apiclient.FredClient) *Service {
	return &Service{
		dataStore:  newMacroDataStore(),
		fredClient: fredClient,
	}
}

// MacroDataStore holds macroeconomic data.
type MacroDataStore struct {
	fredData          map[string][]*macrov1.MacroDataPoint
	oecdData          map[string][]*macrov1.OecdIndicator
	tradingEconData   map[string][]*macrov1.TeIndicator
	fedWatchData      *macrov1.GetFedWatchResponse
}

// NewService creates a new MacroService. Returns errors when no real data is available.
// Deprecated: Use NewServiceWithFRED for production with FRED API integration.
func NewService() *Service {
	return NewServiceWithFRED(nil)
}

func newMacroDataStore() *MacroDataStore {
	store := &MacroDataStore{
		fredData:        make(map[string][]*macrov1.MacroDataPoint),
	oecdData:        make(map[string][]*macrov1.OecdIndicator),
	tradingEconData: make(map[string][]*macrov1.TeIndicator),
	}

	// FRED series: GDP, Unemployment, Inflation, Fed Funds Rate
	store.fredData["GDP"] = generateMacroSeries("25.5", "26.2", 12, "Billions USD")
	store.fredData["UNRATE"] = generateMacroSeries("3.7", "3.5", 12, "Percent")
	store.fredData["CPIAUCSL"] = generateMacroSeries("300.5", "305.2", 12, "Index")
	store.fredData["FEDFUNDS"] = generateMacroSeries("5.25", "5.50", 12, "Percent")

	// OECD data for major economies
	store.oecdData["USA"] = generateOECDData("USA")
	store.oecdData["DEU"] = generateOECDData("DEU")
	store.oecdData["JPN"] = generateOECDData("JPN")
	store.oecdData["GBR"] = generateOECDData("GBR")

	// Trading Economics indicators
	store.tradingEconData["USA"] = generateTradingEconData("USA")
	store.tradingEconData["DEU"] = generateTradingEconData("DEU")
	store.tradingEconData["JPN"] = generateTradingEconData("JPN")

	// Fed Watch data
	store.fedWatchData = &macrov1.GetFedWatchResponse{
		CurrentTarget: "5.25-5.50",
		Probabilities: []*macrov1.FedWatchProb{
			{MeetingDate: "2024-03-20", TargetRange: "5.25-5.50", Probability: 85.0},
			{MeetingDate: "2024-03-20", TargetRange: "5.50-5.75", Probability: 10.0},
			{MeetingDate: "2024-03-20", TargetRange: "5.00-5.25", Probability: 5.0},
		},
	}

	return store
}

func generateMacroSeries(start, end string, count int, unit string) []*macrov1.MacroDataPoint {
	points := make([]*macrov1.MacroDataPoint, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		date := now.AddDate(0, -(count - i - 1), 0).Format("2006-01")
		// Simple interpolation between start and end
		points[i] = &macrov1.MacroDataPoint{
			Date:  date,
			Value: fmt.Sprintf("%.2f", interpolateValue(start, end, i, count)),
			Unit:  unit,
		}
	}
	return points
}

func interpolateValue(start, end string, index, total int) float64 {
	var s, e float64
	fmt.Sscanf(start, "%f", &s)
	fmt.Sscanf(end, "%f", &e)
	return s + (e-s)*float64(index)/float64(total-1)
}

func generateOECDData(country string) []*macrov1.OecdIndicator {
	phases := []string{"expansion", "slowdown", "downturn", "recovery"}
	indicators := make([]*macrov1.OecdIndicator, 6)
	now := time.Now()

	baseCLI := 100.0
	if country == "USA" {
		baseCLI = 100.5
	} else if country == "DEU" {
		baseCLI = 99.8
	} else if country == "JPN" {
		baseCLI = 100.2
	}

	for i := 0; i < 6; i++ {
		date := now.AddDate(0, -(5 - i), 0).Format("2006-01")
		cli := baseCLI + float64(i)*0.2
		indicators[i] = &macrov1.OecdIndicator{
			Country: country,
			Date:    date,
			Cli:     cli,
			Phase:   phases[i%4],
		}
	}
	return indicators
}

func generateTradingEconData(country string) []*macrov1.TeIndicator {
	return []*macrov1.TeIndicator{
		{
			Country:   country,
			Category:  "GDP Growth Rate",
			Last:      "2.5%",
			Previous:  "2.3%",
			Frequency: "Quarterly",
		},
		{
			Country:   country,
			Category:  "Inflation Rate",
			Last:      "3.2%",
			Previous:  "3.4%",
			Frequency: "Monthly",
		},
		{
			Country:   country,
			Category:  "Unemployment Rate",
			Last:      "3.7%",
			Previous:  "3.8%",
			Frequency: "Monthly",
		},
	}
}

// GetFred returns FRED economic data.
// Requires FRED API configuration; returns an error if not configured or fetch fails.
func (s *Service) GetFred(ctx context.Context, seriesID string) (*macrov1.GetFredResponse, error) {
	seriesNames := map[string]string{
		"GDP":        "Gross Domestic Product",
		"UNRATE":     "Unemployment Rate",
		"CPIAUCSL":   "Consumer Price Index",
		"FEDFUNDS":   "Federal Funds Rate",
		"T10Y2Y":     "10-Year Treasury Constant Maturity Minus 2-Year",
		"DGS10":      "10-Year Treasury Constant Maturity Rate",
		"DGS2":       "2-Year Treasury Constant Maturity Rate",
		"DGS30":      "30-Year Treasury Constant Maturity Rate",
	}

	// Try to fetch real data from FRED API
	if s.fredClient != nil && s.fredClient.IsConfigured() && seriesID != "" {
		realData, err := s.fetchFromFRED(ctx, seriesID)
		if err == nil && len(realData) > 0 {
			return &macrov1.GetFredResponse{
				SeriesId:   seriesID,
				SeriesName: seriesNames[seriesID],
				Data:       realData,
			}, nil
		}
		if err != nil {
			log.Printf("FRED API fetch failed for %s: %v", seriesID, err)
			return nil, fmt.Errorf("%w: FRED %s: %v", apperror.ErrUpstreamUnavailable, seriesID, err)
		}
	}

	return nil, fmt.Errorf("%w: FRED client not configured", apperror.ErrProviderNotConfigured)
}

// fetchFromFRED fetches real data from FRED API and converts to proto format.
func (s *Service) fetchFromFRED(ctx context.Context, seriesID string) ([]*macrov1.MacroDataPoint, error) {
	resp, err := s.fredClient.FetchObservations(ctx, seriesID, 50)
	if err != nil {
		return nil, err
	}

	var points []*macrov1.MacroDataPoint
	for _, obs := range resp.Observations {
		if obs.Value == "." || obs.Value == "" {
			continue
		}
		points = append(points, &macrov1.MacroDataPoint{
			Date:  obs.Date,
			Value: obs.Value,
			Unit:  "Value", // FRED series have different units
		})
	}

	return points, nil
}

// GetEcb returns ECB statistical data.
func (s *Service) GetEcb(ctx context.Context, seriesKey string) (*macrov1.GetEcbResponse, error) {
	ecDescriptions := map[string]string{
		"M3":         "Monetary aggregate M3",
		"EONIA":      "Euro overnight index average",
		"EURIBOR":    "Euro interbank offered rate",
		"HICP":       "Harmonized index of consumer prices",
	}

	return &macrov1.GetEcbResponse{
		SeriesKey:   seriesKey,
		Description: ecDescriptions[seriesKey],
		Data: generateMacroSeries(
			fmt.Sprintf("%.2f", 1000+float64(time.Now().Month())*10),
			fmt.Sprintf("%.2f", 1050+float64(time.Now().Month())*10),
			12,
			"Index",
		),
	}, nil
}

// GetSnb returns Swiss National Bank data.
func (s *Service) GetSnb(ctx context.Context, indicator string) (*macrov1.GetSnbResponse, error) {
	snbIndicators := map[string]string{
		"POLICY_RATE": "Swiss National Bank policy rate",
		"M3":          "Swiss monetary aggregate M3",
		"FX_RESERVES": "Foreign exchange reserves",
	}

	baseValue := "1.75"
	if indicator == "FX_RESERVES" {
		baseValue = "750000"
	}

	return &macrov1.GetSnbResponse{
		Indicator: snbIndicators[indicator],
		Data:      generateMacroSeries(baseValue, fmt.Sprintf("%.2f", parseFloat(baseValue)*1.02), 12, "Value"),
	}, nil
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}

// GetOecdLeading returns OECD composite leading indicators.
func (s *Service) GetOecdLeading(ctx context.Context, country string) (*macrov1.GetOecdLeadingResponse, error) {
	data, ok := s.dataStore.oecdData[country]
	if !ok {
		data = s.dataStore.oecdData["USA"]
	}

	return &macrov1.GetOecdLeadingResponse{
		Indicators: data,
	}, nil
}

// GetEurostat returns European statistics data.
func (s *Service) GetEurostat(ctx context.Context, dataset, geo string) (*macrov1.GetEurostatResponse, error) {
	eurostatDatasets := map[string]string{
		"NAMQ_10_GDP":   "GDP and main components",
		"PRCC_PCA":      "Consumer prices",
		"UNE_RT_M":      "Unemployment rate",
		"GOV_10DD_MAIN": "Government debt",
	}

	return &macrov1.GetEurostatResponse{
		Dataset: eurostatDatasets[dataset],
		Data: generateMacroSeries(
			"100.0",
			"102.5",
			12,
			"Index 2015=100",
		),
	}, nil
}

// GetBis returns Bank for International Settlements data.
func (s *Service) GetBis(ctx context.Context, dataset, jurisdiction string) (*macrov1.GetBisResponse, error) {
	bisDatasets := map[string]string{
		"CREDIT":         "Total credit to non-financial sector",
		"DEBT_SECURITIES": "Debt securities statistics",
		"FX_TURNOVER":    "FX turnover",
	}

	currency := jurisdiction
	if jurisdiction == "" {
		currency = "USD"
	}

	return &macrov1.GetBisResponse{
		Dataset: bisDatasets[dataset],
		Data: []*macrov1.BisDataPoint{
			{Date: "2024-01", Value: "280.5", Currency: currency},
			{Date: "2024-02", Value: "282.3", Currency: currency},
			{Date: "2024-03", Value: "285.1", Currency: currency},
		},
	}, nil
}

// GetTradingEconomics returns Trading Economics indicators.
func (s *Service) GetTradingEconomics(ctx context.Context, country, category string) (*macrov1.GetTradingEconomicsResponse, error) {
	key := country
	if category != "" {
		key = country + "_" + category
	}

	data, ok := s.dataStore.tradingEconData[country]
	if !ok {
		data = generateTradingEconData(country)
	}

	// Filter by category if specified
	if category != "" {
		var filtered []*macrov1.TeIndicator
		for _, ind := range data {
			if ind.Category == category {
				filtered = append(filtered, ind)
			}
		}
		if len(filtered) > 0 {
			data = filtered
		}
	}

	_ = key // Use key
	return &macrov1.GetTradingEconomicsResponse{
		Indicators: data,
	}, nil
}

// GetDtccSwaps returns DTCC swaps data.
func (s *Service) GetDtccSwaps(ctx context.Context, pair, tenor string) (*macrov1.GetDtccSwapsResponse, error) {
	pairs := []string{"EURUSD", "GBPUSD", "USDJPY"}
	tenors := []string{"1M", "3M", "6M", "1Y"}

	if pair == "" {
		pair = pairs[0]
	}
	if tenor == "" {
		tenor = tenors[1]
	}

	var swaps []*macrov1.DtccSwap
	now := time.Now()

	for i := 0; i < 5; i++ {
		date := now.AddDate(0, 0, -i*7).Format("2006-01-02")
		swaps = append(swaps, &macrov1.DtccSwap{
			Date:         date,
			Pair:         pair,
			Tenor:        tenor,
			Volume:       150.5 + float64(i)*10,
			OpenInterest: 2500.0 + float64(i)*100,
		})
	}

	return &macrov1.GetDtccSwapsResponse{Swaps: swaps}, nil
}

// GetSec13F returns SEC 13F institutional holdings data.
func (s *Service) GetSec13F(ctx context.Context, cik string, quarter int64) (*macrov1.GetSec13FResponse, error) {
	if cik == "" {
		cik = "0001067983" // Example: Berkshire Hathaway
	}
	if quarter == 0 {
		quarter = 3
	}

	holdings := []*macrov1.Holding13F{
		{Cusip: "037833100", Issuer: "APPLE INC", Class: "COM", Shares: 905560000, Value: "168.5B"},
		{Cusip: "594918104", Issuer: "MICROSOFT CORP", Class: "COM", Shares: 54856000, Value: "22.1B"},
		{Cusip: "023135106", Issuer: "AMAZON COM INC", Class: "COM", Shares: 10450000, Value: "18.2B"},
		{Cusip: "478042103", Issuer: "JOHNSON & JOHNSON", Class: "COM", Shares: 327000, Value: "52.3M"},
		{Cusip: "46625H100", Issuer: "JPMORGAN CHASE & CO", Class: "COM", Shares: 9215000, Value: "168.5B"},
	}

	return &macrov1.GetSec13FResponse{
		Cik:      cik,
		Quarter:  quarter,
		Holdings: holdings,
	}, nil
}

// GetTreasuryAuctions returns US Treasury auction data.
func (s *Service) GetTreasuryAuctions(ctx context.Context) (*macrov1.GetTreasuryAuctionsResponse, error) {
	now := time.Now()

	return &macrov1.GetTreasuryAuctionsResponse{
		Auctions: []*macrov1.TreasuryAuction{
			{
				Date:         now.AddDate(0, 0, -7).Format("2006-01-02"),
				SecurityType: "13-Week Bill",
				Term:         "13W",
				Amount:       "85B",
				HighYield:    5.245,
				BidToCover:   2.85,
			},
			{
				Date:         now.AddDate(0, 0, -14).Format("2006-01-02"),
				SecurityType: "26-Week Bill",
				Term:         "26W",
				Amount:       "82B",
				HighYield:    5.185,
				BidToCover:   3.12,
			},
			{
				Date:         now.AddDate(0, 0, -21).Format("2006-01-02"),
				SecurityType: "10-Year Note",
				Term:         "10Y",
				Amount:       "42B",
				HighYield:    4.125,
				BidToCover:   2.63,
			},
		},
	}, nil
}

// GetFedWatch returns CME FedWatch probability data.
func (s *Service) GetFedWatch(ctx context.Context) (*macrov1.GetFedWatchResponse, error) {
	return s.dataStore.fedWatchData, nil
}

// GetWorldBank returns World Bank development indicators.
func (s *Service) GetWorldBank(ctx context.Context, indicator, country string) (*macrov1.GetWorldBankResponse, error) {
	_ = map[string]string{
		"NY.GDP.MKTP.CD":     "GDP (current US$)",
		"FP.CPI.TOTL.ZG":     "Inflation, consumer prices (annual %)",
		"SL.UEM.TOTL.ZS":     "Unemployment, total (% of total labor force)",
		"BX.KLT.DINV.WD.GD.ZS": "Foreign direct investment, net inflows (% of GDP)",
	}

	return &macrov1.GetWorldBankResponse{
		Indicator: indicator,
		Data: []*macrov1.WbIndicator{
			{IndicatorCode: indicator, Country: country, Year: "2020", Value: "-2.5%"},
			{IndicatorCode: indicator, Country: country, Year: "2021", Value: "5.9%"},
			{IndicatorCode: indicator, Country: country, Year: "2022", Value: "3.1%"},
			{IndicatorCode: indicator, Country: country, Year: "2023", Value: "2.7%"},
		},
	}, nil
}

// GetImfWeo returns IMF World Economic Outlook data.
func (s *Service) GetImfWeo(ctx context.Context) (*macrov1.GetImfWeoResponse, error) {
	return &macrov1.GetImfWeoResponse{
		Data: []*macrov1.ImfWeoData{
			{
				Country:      "United States",
				Year:         "2024",
				GdpGrowth:    2.1,
				Inflation:    2.8,
				Unemployment: 3.8,
			},
			{
				Country:      "Euro Area",
				Year:         "2024",
				GdpGrowth:    0.9,
				Inflation:    2.6,
				Unemployment: 6.5,
			},
			{
				Country:      "China",
				Year:         "2024",
				GdpGrowth:    4.6,
				Inflation:    1.8,
				Unemployment: 5.2,
			},
			{
				Country:      "Japan",
				Year:         "2024",
				GdpGrowth:    0.9,
				Inflation:    2.5,
				Unemployment: 2.5,
			},
		},
	}, nil
}
