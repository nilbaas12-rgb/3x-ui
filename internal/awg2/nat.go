package awg2

import (
	"fmt"
	"net"
	"strings"
)

// NATRule represents a NAT forwarding rule
type NATRule struct {
	Name        string // identifier for the rule
	Interface   string // network interface (e.g., "eth0")
	SourceNet   string // source network (e.g., "10.9.9.0/24")
	IptablesCmd string // iptables command
}

// GenerateNATRules generates NAT rules that don't conflict with existing rules
func GenerateNATRules(awgInterface, sourceNet, outboundInterface string) []NATRule {
	rules := make([]NATRule, 0)

	// Forward rule: Allow packets from AWG subnet through forwarding
	rules = append(rules, NATRule{
		Name:      "awg_forward",
		Interface: awgInterface,
		SourceNet: sourceNet,
		IptablesCmd: fmt.Sprintf(
			"iptables -I FORWARD -i %s -j ACCEPT",
			awgInterface,
		),
	})

	// NAT rule: Masquerade packets from AWG subnet going out
	rules = append(rules, NATRule{
		Name:      "awg_nat",
		Interface: outboundInterface,
		SourceNet: sourceNet,
		IptablesCmd: fmt.Sprintf(
			"iptables -t nat -I POSTROUTING -s %s -o %s -j MASQUERADE",
			sourceNet,
			outboundInterface,
		),
	})

	return rules
}

// GenerateNATCleanup generates cleanup commands to remove NAT rules
func GenerateNATCleanup(rules []NATRule) []string {
	cleanupCmds := make([]string, len(rules))

	for i, rule := range rules {
		// Replace -I (insert) with -D (delete)
		cleanupCmds[i] = strings.ReplaceAll(rule.IptablesCmd, " -I ", " -D ")
	}

	return cleanupCmds
}

// BuildPostUp builds the PostUp command from rules
func BuildPostUp(rules []NATRule) string {
	commands := make([]string, len(rules))
	for i, rule := range rules {
		commands[i] = rule.IptablesCmd
	}
	return strings.Join(commands, "; ")
}

// BuildPostDown builds the PostDown command from rules
func BuildPostDown(rules []NATRule) string {
	commands := make([]string, len(rules))
	for i, rule := range rules {
		commands[i] = strings.ReplaceAll(rule.IptablesCmd, " -I ", " -D ")
	}
	return strings.Join(commands, "; ")
}

// ValidateNAT validates NAT configuration
func ValidateNAT(subnet, outboundInterface string) error {
	// Validate subnet is valid CIDR
	_, _, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet CIDR: %w", err)
	}

	// Validate interface name (basic check)
	if outboundInterface == "" {
		return fmt.Errorf("outbound interface cannot be empty")
	}

	if len(outboundInterface) > 16 {
		return fmt.Errorf("interface name too long: %s", outboundInterface)
	}

	return nil
}

// IsValidInterface checks if interface name follows Linux conventions
func IsValidInterface(name string) bool {
	if len(name) == 0 || len(name) > 15 {
		return false
	}
	// Allow alphanumeric and some special chars
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}
