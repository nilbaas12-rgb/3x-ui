package model

import (
	"time"
)

// AWG2Inbound represents an AmneziaWG 2.0 inbound configuration
type AWG2Inbound struct {
	ID              uint      `gorm:"primaryKey"`
	InboundID       int       `gorm:"index"`
	Port            uint16
	Subnet          string // e.g., "10.9.9.1/24"
	MTU             uint16 // default 1280
	DisableIPv6     bool
	AllowedIPsMode  int // 1, 2, 3
	AllowedIPs      string

	// Obfuscation parameters
	Jc   int
	Jmin int
	Jmax int
	S1   int
	S2   int
	S3   int
	S4   int
	H1   string
	H2   string
	H3   string
	H4   string
	I1   string

	// Server keys
	ServerPrivateKey string
	ServerPublicKey  string
	Preset           string // e.g., "default"
	Status           string // running, stopped, error
	Remarks          string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// AWG2Client represents an AmneziaWG 2.0 client
type AWG2Client struct {
	ID           uint      `gorm:"primaryKey"`
	InboundID    uint      `gorm:"index"`
	Email        string    `gorm:"index"`
	PublicKey    string
	PrivateKey   string
	AllowedIPs   string // e.g., "10.9.9.2/32"
	Name         string
	Enable       bool
	ExpiryTime   int64 // Unix timestamp in milliseconds
	UsedTraffic  uint64
	Download     uint64
	Upload       uint64
	Limit        int64 // Total limit in bytes
	LimitIP      int

	CreatedAt int64 // Unix timestamp in milliseconds
	UpdatedAt int64
}

// TableName specifies table name for AWG2Inbound
func (AWG2Inbound) TableName() string {
	return "awg2_inbounds"
}

// TableName specifies table name for AWG2Client
func (AWG2Client) TableName() string {
	return "awg2_clients"
}
