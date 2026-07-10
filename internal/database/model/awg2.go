package model

import (
	"time"
)

// AWG2Inbound represents an AmneziaWG 2.0 inbound configuration in database
type AWG2Inbound struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	InboundID          int       `gorm:"uniqueIndex" json:"inboundId"`
	Port               uint16    `gorm:"index" json:"port"`
	Subnet             string    `json:"subnet"`
	MTU                uint16    `json:"mtu"`
	OutboundInterface  string    `json:"outboundInterface"`
	DisableIPv6        bool      `json:"disableIpv6"`
	AllowedIPsMode     int       `json:"allowedIpsMode"`
	ServerPrivateKey   string    `json:"serverPrivateKey"`
	ServerPublicKey    string    `json:"serverPublicKey"`

	// Obfuscation parameters
	Jc   uint8  `json:"jc"`
	Jmin uint8  `json:"jmin"`
	Jmax uint8  `json:"jmax"`
	S1   uint8  `json:"s1"`
	S2   uint8  `json:"s2"`
	S3   uint8  `json:"s3"`
	S4   uint8  `json:"s4"`
	H1   string `json:"h1"`
	H2   string `json:"h2"`
	H3   string `json:"h3"`
	H4   string `json:"h4"`
	I1   string `json:"i1"`

	// Status
	Preset   string `json:"preset"`
	Status   string `json:"status"`   // created, running, stopped, error
	Error    string `json:"error"`    // error message if any
	Remarks  string `json:"remarks"`

	// Metadata
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// AWG2Client represents an AmneziaWG 2.0 client in database
type AWG2Client struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	InboundID    uint      `gorm:"index" json:"inboundId"`
	Email        string    `gorm:"index" json:"email"`
	PublicKey    string    `json:"publicKey"`
	PrivateKey   string    `json:"privateKey"`
	AllowedIPs   string    `json:"allowedIps"`
	Name         string    `json:"name"`
	Enable       bool      `gorm:"default:true" json:"enable"`
	ExpiryTime   int64     `json:"expiryTime"`     // unix timestamp in milliseconds
	UsedTraffic  uint64    `gorm:"default:0" json:"usedTraffic"`
	Download     uint64    `gorm:"default:0" json:"download"`
	Upload       uint64    `gorm:"default:0" json:"upload"`
	TrafficLimit int64     `gorm:"default:0" json:"trafficLimit"`
	IPLimit      int       `gorm:"default:0" json:"ipLimit"`

	// Metadata
	CreatedAt int64 `json:"createdAt"` // unix timestamp in milliseconds
	UpdatedAt int64 `json:"updatedAt"`
}

// TableName specifies the table name for AWG2Inbound
func (AWG2Inbound) TableName() string {
	return "awg2_inbounds"
}

// TableName specifies the table name for AWG2Client
func (AWG2Client) TableName() string {
	return "awg2_clients"
}
