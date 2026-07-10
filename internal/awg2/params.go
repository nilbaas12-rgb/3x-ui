package awg2

import (
	"fmt"
	"regexp"
	"strings"
)

// ObfuscationParams holds AmneziaWG 2.0 obfuscation parameters
type ObfuscationParams struct {
	// Jitter parameters
	Jc   int // Jitter correction (default 6)
	Jmin int // Minimum jitter (default 55)
	Jmax int // Maximum jitter (default 205)

	// Packet size parameters
	S1 int // First packet size (default 72)
	S2 int // Second packet size (default 56)
	S3 int // Third packet size (default 32)
	S4 int // Fourth packet size (default 16)

	// Header obfuscation ranges
	H1 string // Header 1 range (e.g., "234567-345678")
	H2 string // Header 2 range
	H3 string // Header 3 range
	H4 string // Header 4 range

	// Initial packet parameters
	I1 string // Initial packet format (e.g., "<r 128>")
}

// DefaultObfuscationParams returns the default AWG2 obfuscation parameters
func DefaultObfuscationParams() ObfuscationParams {
	return ObfuscationParams{
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
	}
}

// ValidateParams validates obfuscation parameters
func ValidateParams(p ObfuscationParams) error {
	if p.Jc < 0 || p.Jc > 255 {
		return fmt.Errorf("invalid Jc value: %d (must be 0-255)", p.Jc)
	}
	if p.Jmin < 0 || p.Jmin > 255 {
		return fmt.Errorf("invalid Jmin value: %d (must be 0-255)", p.Jmin)
	}
	if p.Jmax < 0 || p.Jmax > 255 {
		return fmt.Errorf("invalid Jmax value: %d (must be 0-255)", p.Jmax)
	}
	if p.Jmin > p.Jmax {
		return fmt.Errorf("Jmin (%d) cannot be greater than Jmax (%d)", p.Jmin, p.Jmax)
	}

	// Validate S parameters
	for i, s := range []int{p.S1, p.S2, p.S3, p.S4} {
		if s < 0 || s > 255 {
			return fmt.Errorf("invalid S%d value: %d (must be 0-255)", i+1, s)
		}
	}

	// Validate H parameters (range format: "num1-num2")
	for i, h := range []string{p.H1, p.H2, p.H3, p.H4} {
		if !isValidRange(h) {
			return fmt.Errorf("invalid H%d range format: %s (expected format: 'num1-num2')", i+1, h)
		}
	}

	// Validate I1 parameter
	if p.I1 == "" || !strings.HasPrefix(p.I1, "<r") {
		return fmt.Errorf("invalid I1 format: %s (expected format: '<r N>')", p.I1)
	}

	return nil
}

// isValidRange checks if a range string is in valid format
func isValidRange(r string) bool {
	if r == "" {
		return false
	}
	reg := regexp.MustCompile(`^\d+-\d+$`)
	return reg.MatchString(r)
}

// ParseRange parses a range string and returns start and end
func ParseRange(r string) (int64, int64, error) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid range format: %s", r)
	}

	var start, end int64
	fmt.Sscanf(parts[0], "%d", &start)
	fmt.Sscanf(parts[1], "%d", &end)

	if start > end {
		return 0, 0, fmt.Errorf("invalid range: start (%d) > end (%d)", start, end)
	}

	return start, end, nil
}
