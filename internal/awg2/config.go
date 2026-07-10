package awg2

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ServerConfig represents complete server configuration
type ServerConfig struct {
	PrivateKey        string            `json:"privateKey"`
	PublicKey         string            `json:"publicKey"`
	Address           string            `json:"address"`           // e.g., "10.9.9.1/24"
	Port              uint16            `json:"port"`              // e.g., 51820
	MTU               uint16            `json:"mtu"`               // default 1280
	Interface         string            `json:"interface"`         // e.g., "awg0"
	OutboundInterface string            `json:"outboundInterface"` // e.g., "eth0"
	Params            ObfuscationParams `json:"params"`
	Clients           []ClientConfigData `json:"clients"`
	NATRules          []NATRule         `json:"natRules"`
}

// ClientConfigData represents a client in server config
type ClientConfigData struct {
	Name       string `json:"name"`
	PublicKey  string `json:"publicKey"`
	AllowedIPs string `json:"allowedIps"`
}

// ClientConfig represents a client-side configuration
type ClientConfig struct {
	PrivateKey          string `json:"privateKey"`
	Address             string `json:"address"`              // e.g., "10.9.9.2/32"
	ServerPublicKey     string `json:"serverPublicKey"`
	ServerEndpoint      string `json:"serverEndpoint"`       // "IP:PORT"
	DNS                 string `json:"dns,omitempty"`        // comma-separated
	MTU                 uint16 `json:"mtu"`
	AllowedIPs          string `json:"allowedIps"`
	PersistentKeepalive int    `json:"persistentKeepalive"`
}

// GenerateServerConfigINI generates INI format server configuration
func GenerateServerConfigINI(sc *ServerConfig) (string, error) {
	if err := sc.Params.Validate(); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	if err := ValidateNAT(sc.Address, sc.OutboundInterface); err != nil {
		return "", err
	}

	var buf strings.Builder

	// [Interface] section
	buf.WriteString("[Interface]\n")
	fmt.Fprintf(&buf, "PrivateKey = %s\n", sc.PrivateKey)
	fmt.Fprintf(&buf, "Address = %s\n", sc.Address)
	fmt.Fprintf(&buf, "ListenPort = %d\n", sc.Port)
	fmt.Fprintf(&buf, "MTU = %d\n", sc.MTU)

	// Generate and include NAT rules
	if len(sc.NATRules) > 0 {
		fmt.Fprintf(&buf, "PostUp = %s\n", BuildPostUp(sc.NATRules))
		fmt.Fprintf(&buf, "PostDown = %s\n", BuildPostDown(sc.NATRules))
	}

	// Obfuscation parameters
	buf.WriteString("\n# AmneziaWG 2.0 Obfuscation Parameters\n")
	fmt.Fprintf(&buf, "Jc = %d\n", sc.Params.Jc)
	fmt.Fprintf(&buf, "Jmin = %d\n", sc.Params.Jmin)
	fmt.Fprintf(&buf, "Jmax = %d\n", sc.Params.Jmax)
	fmt.Fprintf(&buf, "S1 = %d\n", sc.Params.S1)
	fmt.Fprintf(&buf, "S2 = %d\n", sc.Params.S2)
	fmt.Fprintf(&buf, "S3 = %d\n", sc.Params.S3)
	fmt.Fprintf(&buf, "S4 = %d\n", sc.Params.S4)
	fmt.Fprintf(&buf, "H1 = %s\n", sc.Params.H1)
	fmt.Fprintf(&buf, "H2 = %s\n", sc.Params.H2)
	fmt.Fprintf(&buf, "H3 = %s\n", sc.Params.H3)
	fmt.Fprintf(&buf, "H4 = %s\n", sc.Params.H4)
	fmt.Fprintf(&buf, "I1 = %s\n", sc.Params.I1)

	// Peers (Clients)
	for _, client := range sc.Clients {
		buf.WriteString("\n[Peer]\n")
		if client.Name != "" {
			fmt.Fprintf(&buf, "# %s\n", client.Name)
		}
		fmt.Fprintf(&buf, "PublicKey = %s\n", client.PublicKey)
		fmt.Fprintf(&buf, "AllowedIPs = %s\n", client.AllowedIPs)
	}

	return buf.String(), nil
}

// GenerateClientConfigConf generates CONF format client configuration
func GenerateClientConfigConf(cc *ClientConfig) (string, error) {
	// Validate required fields
	if cc.PrivateKey == "" || cc.ServerPublicKey == "" || cc.ServerEndpoint == "" {
		return "", fmt.Errorf("missing required fields")
	}

	// Validate address is CIDR format
	if _, _, err := net.ParseCIDR(cc.Address); err != nil {
		return "", fmt.Errorf("invalid address CIDR: %w", err)
	}

	var buf strings.Builder

	buf.WriteString("[Interface]\n")
	fmt.Fprintf(&buf, "PrivateKey = %s\n", cc.PrivateKey)
	fmt.Fprintf(&buf, "Address = %s\n", cc.Address)

	if cc.DNS != "" {
		fmt.Fprintf(&buf, "DNS = %s\n", cc.DNS)
	}

	if cc.MTU > 0 {
		fmt.Fprintf(&buf, "MTU = %d\n", cc.MTU)
	}

	buf.WriteString("\n[Peer]\n")
	fmt.Fprintf(&buf, "PublicKey = %s\n", cc.ServerPublicKey)
	fmt.Fprintf(&buf, "Endpoint = %s\n", cc.ServerEndpoint)
	fmt.Fprintf(&buf, "AllowedIPs = %s\n", cc.AllowedIPs)

	if cc.PersistentKeepalive > 0 {
		fmt.Fprintf(&buf, "PersistentKeepalive = %d\n", cc.PersistentKeepalive)
	}

	return buf.String(), nil
}

// AllocateClientIP allocates a new IP for a client from subnet
func AllocateClientIP(subnet string, index int) (string, error) {
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", fmt.Errorf("invalid subnet: %w", err)
	}

	if index < 2 || index > 254 {
		return "", fmt.Errorf("client index out of range (2-254): %d", index)
	}

	// Calculate IP address
	baseIP := ipnet.IP
	baseIP = baseIP.To4()
	if baseIP == nil {
		return "", fmt.Errorf("subnet must be IPv4")
	}

	// Modify the last octet
	clientIP := make(net.IP, len(baseIP))
	copy(clientIP, baseIP)
	clientIP[3] = byte(index)

	return clientIP.String() + "/32", nil
}

// ParseSubnet parses subnet and returns base IP and CIDR
func ParseSubnet(subnet string) (string, int, error) {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", 0, err
	}

	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		return "", 0, fmt.Errorf("must be IPv4 subnet")
	}

	return ip.String(), ones, nil
}

// ParseMTU parses MTU string to uint16
func ParseMTU(mtuStr string) (uint16, error) {
	mtu, err := strconv.ParseUint(mtuStr, 10, 16)
	if err != nil {
		return 0, err
	}
	if mtu < 576 || mtu > 65535 {
		return 0, fmt.Errorf("MTU must be between 576 and 65535, got %d", mtu)
	}
	return uint16(mtu), nil
}

// ParsePort parses port string to uint16
func ParsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	if port == 0 || port > 65535 {
		return 0, fmt.Errorf("invalid port: %d", port)
	}
	return uint16(port), nil
}
