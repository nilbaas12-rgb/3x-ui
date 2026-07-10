package awg2

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ObfuscationParams holds all AmneziaWG 2.0 obfuscation parameters
type ObfuscationParams struct {
	Jc   uint8 `json:"jc"`
	Jmin uint8 `json:"jmin"`
	Jmax uint8 `json:"jmax"`
	S1   uint8 `json:"s1"`
	S2   uint8 `json:"s2"`
	S3   uint8 `json:"s3"`
	S4   uint8 `json:"s4"`
	H1   string `json:"h1"`
	H2   string `json:"h2"`
	H3   string `json:"h3"`
	H4   string `json:"h4"`
	I1   string `json:"i1"`
}

// Preset represents a pre-configured obfuscation profile
type Preset struct {
	Name   string
	Params ObfuscationParams
	Desc   string
}

// Available presets
var Presets = map[string]Preset{
	"default": {
		Name: "default",
		Params: ObfuscationParams{
			Jc:   6,
			Jmin: 55,
			Jmax: 205,
			S1:   72,
			S2:   56,
			S3:   32,
			S4:   16,
			H1:   "234567-345678",
			H2:   "3456789-4567890",
			H3:   "56789012-67890123",
			H4:   "456789012-567890123",
			I1:   "<r 128>",
		},
		Desc: "Standard obfuscation (recommended)",
	},
	"heavy": {
		Name: "heavy",
		Params: ObfuscationParams{
			Jc:   8,
			Jmin: 40,
			Jmax: 250,
			S1:   100,
			S2:   80,
			S3:   40,
			S4:   20,
			H1:   "100000-200000",
			H2:   "200001-300000",
			H3:   "300001-400000",
			H4:   "400001-500000",
			I1:   "<r 256>",
		},
		Desc: "Heavy obfuscation (higher CPU, better DPI bypass)",
	},
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s (value=%v)", e.Field, e.Message, e.Value)
}

// Validate checks if obfuscation parameters are valid
func (p *ObfuscationParams) Validate() error {
	if p.Jmin > p.Jmax {
		return ValidationError{"Jmin/Jmax", p.Jmin, "Jmin must be <= Jmax"}
	}

	for i, h := range []string{p.H1, p.H2, p.H3, p.H4} {
		if !isValidRange(h) {
			return ValidationError{fmt.Sprintf("H%d", i+1), h, "invalid range format (expected 'num-num')"}
		}
	}

	if !strings.HasPrefix(p.I1, "<r") || !strings.HasSuffix(p.I1, ">") {
		return ValidationError{"I1", p.I1, "invalid format (expected '<r N>')"}
	}

	return nil
}

// isValidRange validates range format "num1-num2"
func isValidRange(r string) bool {
	reg := regexp.MustCompile(`^\d+-\d+$`)
	return reg.MatchString(r)
}

// ParseRange parses range string and returns start, end
func ParseRange(r string) (uint64, uint64, error) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: %s", r)
	}

	start, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid start value: %s", parts[0])
	}

	end, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid end value: %s", parts[1])
	}

	if start > end {
		return 0, 0, fmt.Errorf("start (%d) cannot be greater than end (%d)", start, end)
	}

	return start, end, nil
}

// GetPreset returns a preset by name
func GetPreset(name string) (*Preset, error) {
	p, ok := Presets[name]
	if !ok {
		available := make([]string, 0, len(Presets))
		for k := range Presets {
			available = append(available, k)
		}
		return nil, fmt.Errorf("unknown preset: %s (available: %v)", name, available)
	}
	return &p, nil
}

// Clone creates a deep copy of ObfuscationParams
func (p *ObfuscationParams) Clone() *ObfuscationParams {
	return &ObfuscationParams{
		Jc:   p.Jc,
		Jmin: p.Jmin,
		Jmax: p.Jmax,
		S1:   p.S1,
		S2:   p.S2,
		S3:   p.S3,
		S4:   p.S4,
		H1:   p.H1,
		H2:   p.H2,
		H3:   p.H3,
		H4:   p.H4,
		I1:   p.I1,
	}
}

// RandomizeParams generates randomized parameters for additional obfuscation
func (p *ObfuscationParams) Randomize() {
	// Randomize jitter within reasonable bounds
	p.Jc = uint8(rand.Intn(10) + 4)   // 4-13
	p.Jmin = uint8(rand.Intn(50) + 30) // 30-79
	p.Jmax = uint8(rand.Intn(100) + 150) // 150-249

	// Ensure Jmin <= Jmax
	if p.Jmin > p.Jmax {
		p.Jmin, p.Jmax = p.Jmax, p.Jmin
	}

	// Randomize packet sizes
	p.S1 = uint8(rand.Intn(80) + 60)  // 60-139
	p.S2 = uint8(rand.Intn(60) + 40)  // 40-99
	p.S3 = uint8(rand.Intn(30) + 20)  // 20-49
	p.S4 = uint8(rand.Intn(20) + 10)  // 10-29

	// Randomize header ranges
	p.H1 = randomRange(100000, 500000)
	p.H2 = randomRange(500001, 1000000)
	p.H3 = randomRange(1000001, 2000000)
	p.H4 = randomRange(2000001, 3000000)

	// Randomize initial packet size
	rSize := rand.Intn(256) + 64 // 64-319
	p.I1 = fmt.Sprintf("<r %d>", rSize)
}

// randomRange generates a random range string "min-max"
func randomRange(minBound, maxBound int) string {
	start := rand.Intn(maxBound-minBound) + minBound
	end := start + rand.Intn(500000) + 100000
	return fmt.Sprintf("%d-%d", start, end)
}
