package service

import (
	"encoding/json"
	"fmt"

	"github.com/mhsanaei/3x-ui/v3/internal/awg2"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
)

// AWG2Settings represents AWG2-specific inbound settings
type AWG2Settings struct {
	Port           int                    `json:"port"`
	Subnet         string                 `json:"subnet"`
	MTU            int                    `json:"mtu"`
	DisableIPv6    bool                   `json:"disableIPv6"`
	AllowedIPsMode int                    `json:"allowedIpsMode"`
	Params         awg2.ObfuscationParams `json:"params"`
	Clients        []model.AWG2Client     `json:"clients"`
}

// GetAWG2Inbound retrieves an AWG2 inbound by ID
func (s *InboundService) GetAWG2Inbound(inboundID int) (*model.AWG2Inbound, error) {
	db := database.GetDB()
	var awg2Inbound model.AWG2Inbound
	err := db.Where("inbound_id = ?", inboundID).First(&awg2Inbound).Error
	if err != nil {
		return nil, err
	}
	return &awg2Inbound, nil
}

// CreateAWG2Inbound creates a new AWG2 inbound
func (s *InboundService) CreateAWG2Inbound(inbound *model.Inbound, settings AWG2Settings) (*model.AWG2Inbound, error) {
	db := database.GetDB()

	// Validate parameters
	if err := awg2.ValidateParams(settings.Params); err != nil {
		return nil, fmt.Errorf("invalid AWG2 parameters: %w", err)
	}

	// Generate server keys
	keyPair, err := awg2.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	awg2Inbound := &model.AWG2Inbound{
		InboundID:        uint(inbound.Id),
		Port:             uint16(settings.Port),
		Subnet:           settings.Subnet,
		MTU:              uint16(settings.MTU),
		DisableIPv6:      settings.DisableIPv6,
		AllowedIPsMode:   settings.AllowedIPsMode,
		Jc:               settings.Params.Jc,
		Jmin:             settings.Params.Jmin,
		Jmax:             settings.Params.Jmax,
		S1:               settings.Params.S1,
		S2:               settings.Params.S2,
		S3:               settings.Params.S3,
		S4:               settings.Params.S4,
		H1:               settings.Params.H1,
		H2:               settings.Params.H2,
		H3:               settings.Params.H3,
		H4:               settings.Params.H4,
		I1:               settings.Params.I1,
		ServerPrivateKey: keyPair.PrivateKey,
		ServerPublicKey:  keyPair.PublicKey,
		Preset:           "default",
		Status:           "created",
	}

	if err := db.Create(awg2Inbound).Error; err != nil {
		logger.Error("[awg2] Failed to create AWG2 inbound:", err)
		return nil, err
	}

	logger.Info("[awg2] Created new AWG2 inbound", awg2Inbound.ID)
	return awg2Inbound, nil
}

// AddAWG2Client adds a new client to an AWG2 inbound
func (s *InboundService) AddAWG2Client(inboundID int, email string) (*model.AWG2Client, error) {
	db := database.GetDB()

	// Get AWG2 inbound
	awg2Inbound, err := s.GetAWG2Inbound(inboundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get AWG2 inbound: %w", err)
	}

	// Generate client keys
	keyPair, err := awg2.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client keys: %w", err)
	}

	// Allocate IP for client
	// This is simplified; in production, implement proper IP allocation
	clientIP := fmt.Sprintf("%s.%d/32", awg2Inbound.Subnet[:len(awg2Inbound.Subnet)-5], 100) // placeholder

	client := &model.AWG2Client{
		InboundID:  awg2Inbound.ID,
		Email:      email,
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
		AllowedIPs: clientIP,
		Enable:     true,
	}

	if err := db.Create(client).Error; err != nil {
		logger.Error("[awg2] Failed to add client:", err)
		return nil, err
	}

	logger.Info("[awg2] Added new client", email, "to inbound", inboundID)
	return client, nil
}

// GetAWG2Clients retrieves all clients for an AWG2 inbound
func (s *InboundService) GetAWG2Clients(inboundID int) ([]model.AWG2Client, error) {
	db := database.GetDB()
	var clients []model.AWG2Client
	err := db.Where("inbound_id = ?", inboundID).Find(&clients).Error
	return clients, err
}

// RemoveAWG2Client removes a client from AWG2 inbound
func (s *InboundService) RemoveAWG2Client(clientID uint) error {
	db := database.GetDB()
	if err := db.Delete(&model.AWG2Client{}, clientID).Error; err != nil {
		logger.Error("[awg2] Failed to remove client:", err)
		return err
	}
	logger.Info("[awg2] Removed client", clientID)
	return nil
}

// GenerateAWG2ClientConfig generates a client configuration
func (s *InboundService) GenerateAWG2ClientConfig(inboundID int, clientID uint) (string, error) {
	db := database.GetDB()

	// Get inbound and client
	awg2Inbound, err := s.GetAWG2Inbound(inboundID)
	if err != nil {
		return "", err
	}

	var client model.AWG2Client
	if err := db.First(&client, clientID).Error; err != nil {
		return "", err
	}

	// Get server inbound to get endpoint
	inbound, err := s.GetInbound(inboundID)
	if err != nil {
		return "", err
	}

	// Generate config
	endpoint := fmt.Sprintf("%s:%d", inbound.Remark, awg2Inbound.Port) // Use remark as hostname hint
	config, err := awg2.GenerateClientConfig(
		awg2Inbound.ServerPublicKey,
		endpoint,
		client.PrivateKey,
		client.AllowedIPs,
	)

	if err != nil {
		logger.Error("[awg2] Failed to generate client config:", err)
		return "", err
	}

	return config, nil
}
