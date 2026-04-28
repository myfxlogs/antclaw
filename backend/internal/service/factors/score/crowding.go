package score

import (
	"math"
)

// CrowdingScore calculates crowding risk penalty score
type CrowdingScore struct{}

// NewCrowdingScore creates crowding scorer
func NewCrowdingScore() *CrowdingScore {
	return &CrowdingScore{}
}

// CrowdingInput contains crowding metrics
type CrowdingInput struct {
	COTIndex       float64 // 0-100, commercial positioning
	SpecMomentum   float64 // Speculator momentum (4W change)
	CrowdingIndex  float64 // Overall crowding measure
}

// Calculate computes crowding penalty (-1 to 0)
// Returns negative score when crowded (risk penalty)
func (c *CrowdingScore) Calculate(input CrowdingInput) float64 {
	// High COT index (>80 or <20) indicates extreme positioning
	cotExtreme := 0.0
	if input.COTIndex > 80 {
		cotExtreme = (input.COTIndex - 80) / 20 // 0 to 1
	} else if input.COTIndex < 20 {
		cotExtreme = (20 - input.COTIndex) / 20 // 0 to 1
	}

	// Spec momentum chasing
	specChasing := math.Abs(input.SpecMomentum)
	if specChasing > 1 {
		specChasing = 1
	}

	// Overall crowding index (0-1)
	crowding := input.CrowdingIndex

	// Weighted combination
	penalty := cotExtreme*0.4 + specChasing*0.3 + crowding*0.3

	// Return negative penalty (0 to -1)
	return -penalty
}

// IsCrowded returns true if asset is considered crowded
func (c *CrowdingScore) IsCrowded(input CrowdingInput, threshold float64) bool {
	score := c.Calculate(input)
	return score < -threshold
}
