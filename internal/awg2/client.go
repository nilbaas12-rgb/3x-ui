package awg2

import (
	"time"
)

// Client represents an AWG2 client
type Client struct {
	ID           uint
	InboundID    uint
	Email        string
	PublicKey    string
	PrivateKey   string
	AllowedIPs   string
	Enable       bool
	ExpiryTime   time.Time
	UsedTraffic  uint64
	Download     uint64
	Upload       uint64
	Limit        int64 // Total limit in bytes, 0 = unlimited
	LimitIP      int   // Concurrent IP limit, 0 = unlimited
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsExpired checks if the client has expired
func (c *Client) IsExpired() bool {
	if c.ExpiryTime.IsZero() {
		return false
	}
	return time.Now().After(c.ExpiryTime)
}

// IsTrafficExceeded checks if the client has exceeded traffic limit
func (c *Client) IsTrafficExceeded() bool {
	if c.Limit <= 0 {
		return false
	}
	return c.UsedTraffic >= uint64(c.Limit)
}

// IsEnabled checks if client should be active
func (c *Client) IsEnabled() bool {
	return c.Enable && !c.IsExpired() && !c.IsTrafficExceeded()
}
