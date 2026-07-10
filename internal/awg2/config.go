package awg2

import (
	"fmt"
	"strings"
)

// ServerConfig represents an AWG2 server configuration
type ServerConfig struct {
	PrivateKey string
	Address    string
	Port       int
	MTU        int
	Params     ObfuscationParams
	Peers      []PeerConfig
}

// PeerConfig represents an AWG2 peer (client)
type PeerConfig struct {
	PublicKey  string
	AllowedIPs string
	Name       string // Comment
}

// GenerateServerConfig generates an INI-format server configuration
func GenerateServerConfig(sc *ServerConfig) (string, error) {
	if err := ValidateParams(sc.Params); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}

	var sb strings.Builder

	// [Interface] section
	sb.WriteString("[Interface]\n")
	fmt.Fprintf(&sb, "PrivateKey = %s\n", sc.PrivateKey)
	fmt.Fprintf(&sb, "Address = %s\n", sc.Address)
	fmt.Fprintf(&sb, "ListenPort = %d\n", sc.Port)
	fmt.Fprintf(&sb, "MTU = %d\n", sc.MTU)

	// PostUp/PostDown (example for Linux with iptables)
	sb.WriteString("PostUp = iptables -I FORWARD -i awg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE\n")
	sb.WriteString("PostDown = iptables -D FORWARD -i awg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE\n")

	// Obfuscation parameters
	fmt.Fprintf(&sb, "Jc = %d\n", sc.Params.Jc)
	fmt.Fprintf(&sb, "Jmin = %d\n", sc.Params.Jmin)
	fmt.Fprintf(&sb, "Jmax = %d\n", sc.Params.Jmax)
	fmt.Fprintf(&sb, "S1 = %d\n", sc.Params.S1)
	fmt.Fprintf(&sb, "S2 = %d\n", sc.Params.S2)
	fmt.Fprintf(&sb, "S3 = %d\n", sc.Params.S3)
	fmt.Fprintf(&sb, "S4 = %d\n", sc.Params.S4)
	fmt.Fprintf(&sb, "H1 = %s\n", sc.Params.H1)
	fmt.Fprintf(&sb, "H2 = %s\n", sc.Params.H2)
	fmt.Fprintf(&sb, "H3 = %s\n", sc.Params.H3)
	fmt.Fprintf(&sb, "H4 = %s\n", sc.Params.H4)
	fmt.Fprintf(&sb, "I1 = %s\n", sc.Params.I1)

	// Peers section
	for _, peer := range sc.Peers {
		sb.WriteString("\n[Peer]\n")
		if peer.Name != "" {
			fmt.Fprintf(&sb, "#_Name = %s\n", peer.Name)
		}
		fmt.Fprintf(&sb, "PublicKey = %s\n", peer.PublicKey)
		fmt.Fprintf(&sb, "AllowedIPs = %s\n", peer.AllowedIPs)
	}

	return sb.String(), nil
}

// GenerateClientConfig generates a client configuration for a peer
func GenerateClientConfig(serverPublicKey, serverEndpoint, clientPrivateKey, clientAllowedIP string) (string, error) {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	fmt.Fprintf(&sb, "PrivateKey = %s\n", clientPrivateKey)
	fmt.Fprintf(&sb, "Address = %s\n", clientAllowedIP)
	sb.WriteString("DNS = 8.8.8.8, 1.1.1.1\n")
	sb.WriteString("MTU = 1280\n")

	sb.WriteString("\n[Peer]\n")
	fmt.Fprintf(&sb, "PublicKey = %s\n", serverPublicKey)
	fmt.Fprintf(&sb, "Endpoint = %s\n", serverEndpoint)
	sb.WriteString("AllowedIPs = 0.0.0.0/0\n")
	sb.WriteString("PersistentKeepalive = 25\n")

	return sb.String(), nil
}
