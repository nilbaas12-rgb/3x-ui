package service

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/awg2"
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"gorm.io/gorm"
)

// AWG2Service manages AWG2 inbounds and clients
type AWG2Service struct {
	mu        sync.RWMutex
	instances map[int]*awg2.Instance // inbound ID -> instance
	settings  SettingService         // for DNS and defaults
}

// NewAWG2Service creates a new AWG2 service
func NewAWG2Service(settingService SettingService) *AWG2Service {
	return &AWG2Service{
		instances: make(map[int]*awg2.Instance),
		settings:  settingService,
	}
}

// CreateInbound creates a new AWG2 inbound
func (s *AWG2Service) CreateInbound(
	inboundID int,
	port uint16,
	subnet string,
	outboundInterface string,
	preset string,
	randomizeParams bool,
) (*model.AWG2Inbound, error) {

	// Get or validate preset
	p, err := awg2.GetPreset(preset)
	if err != nil {
		return nil, err
	}

	// Clone and potentially randomize parameters
	params := p.Params.Clone()
	if randomizeParams {
		params.Randomize()
	}

	// Generate server keys
	keyPair, err := awg2.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("key generation failed: %w", err)
	}

	// Validate NAT configuration
	if err := awg2.ValidateNAT(subnet, outboundInterface); err != nil {
		return nil, err
	}

	// Generate NAT rules
	natRules := awg2.GenerateNATRules("awg0", subnet, outboundInterface)

	now := time.Now()
	inbound := &model.AWG2Inbound{
		InboundID:         uint(inboundID),
		Port:              port,
		Subnet:            subnet,
		OutboundInterface: outboundInterface,
		MTU:               1280,
		DisableIPv6:       false,
		AllowedIPsMode:    2,
		ServerPrivateKey:  keyPair.PrivateKey,
		ServerPublicKey:   keyPair.PublicKey,
		Jc:                params.Jc,
		Jmin:              params.Jmin,
		Jmax:              params.Jmax,
		S1:                params.S1,
		S2:                params.S2,
		S3:                params.S3,
		S4:                params.S4,
		H1:                params.H1,
		H2:                params.H2,
		H3:                params.H3,
		H4:                params.H4,
		I1:                params.I1,
		Preset:            preset,
		Status:            awg2.InstanceStateCreated,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	db := database.GetDB()
	if err := db.Create(inbound).Error; err != nil {
		logger.Error("[awg2] Failed to create inbound:", err)
		return nil, err
	}

	// Create instance
	instance := awg2.NewInstance(fmt.Sprintf("inbound-%d", inboundID), port, subnet)
	instance.ID = inbound.ID
	instance.ServerPrivateKey = keyPair.PrivateKey
	instance.ServerPublicKey = keyPair.PublicKey
	instance.ObfuscationParams = params

	s.mu.Lock()
	s.instances[inboundID] = instance
	s.mu.Unlock()

	logger.Info("[awg2] Created inbound", inboundID, "on port", port, "with preset", preset)
	return inbound, nil
}

// RandomizeParams randomizes parameters for an inbound
func (s *AWG2Service) RandomizeParams(inboundID int) (*model.AWG2Inbound, error) {
	db := database.GetDB()

	var inbound model.AWG2Inbound
	if err := db.Where("inbound_id = ?", inboundID).First(&inbound).Error; err != nil {
		return nil, fmt.Errorf("inbound not found: %w", err)
	}

	// Create params and randomize
	params := awg2.ObfuscationParams{
		Jc:   inbound.Jc,
		Jmin: inbound.Jmin,
		Jmax: inbound.Jmax,
		S1:   inbound.S1,
		S2:   inbound.S2,
		S3:   inbound.S3,
		S4:   inbound.S4,
		H1:   inbound.H1,
		H2:   inbound.H2,
		H3:   inbound.H3,
		H4:   inbound.H4,
		I1:   inbound.I1,
	}

	params.Randomize()

	// Update in DB
	updates := map[string]interface{}{
		"jc":  params.Jc,
		"jmin": params.Jmin,
		"jmax": params.Jmax,
		"s1":   params.S1,
		"s2":   params.S2,
		"s3":   params.S3,
		"s4":   params.S4,
		"h1":   params.H1,
		"h2":   params.H2,
		"h3":   params.H3,
		"h4":   params.H4,
		"i1":   params.I1,
	}

	if err := db.Model(&inbound).Updates(updates).Error; err != nil {
		logger.Error("[awg2] Failed to randomize params:", err)
		return nil, err
	}

	// Update instance
	s.mu.RLock()
	instance, exists := s.instances[inboundID]
	s.mu.RUnlock()

	if exists {
		instance.ObfuscationParams = params.Clone()
	}

	logger.Info("[awg2] Randomized params for inbound", inboundID)
	return &inbound, nil
}

// AddClient adds a client to an AWG2 inbound
func (s *AWG2Service) AddClient(inboundID int, email, name string) (*model.AWG2Client, error) {
	db := database.GetDB()

	// Get inbound
	var awg2Inbound model.AWG2Inbound
	if err := db.Where("inbound_id = ?", inboundID).First(&awg2Inbound).Error; err != nil {
		return nil, fmt.Errorf("inbound not found: %w", err)
	}

	// Generate keys
	keyPair, err := awg2.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("key generation failed: %w", err)
	}

	// Allocate IP
	clientIP, err := awg2.AllocateClientIP(awg2Inbound.Subnet, 100+int(awg2Inbound.ID)%150)
	if err != nil {
		return nil, fmt.Errorf("IP allocation failed: %w", err)
	}

	now := time.Now()
	client := &model.AWG2Client{
		InboundID:  awg2Inbound.ID,
		Email:      email,
		Name:       name,
		PublicKey:  keyPair.PublicKey,
		PrivateKey: keyPair.PrivateKey,
		AllowedIPs: clientIP,
		Enable:     true,
		CreatedAt:  now.UnixMilli(),
		UpdatedAt:  now.UnixMilli(),
	}

	if err := db.Create(client).Error; err != nil {
		logger.Error("[awg2] Failed to add client:", err)
		return nil, err
	}

	logger.Info("[awg2] Added client", email, "to inbound", inboundID)
	return client, nil
}

// GetClientConfig generates client configuration
func (s *AWG2Service) GetClientConfig(inboundID int, clientID uint, endpoint string) (string, error) {
	db := database.GetDB()

	// Get inbound
	var awg2Inbound model.AWG2Inbound
	if err := db.Where("inbound_id = ?", inboundID).First(&awg2Inbound).Error; err != nil {
		return "", fmt.Errorf("inbound not found: %w", err)
	}

	// Get client
	var client model.AWG2Client
	if err := db.First(&client, clientID).Error; err != nil {
		return "", fmt.Errorf("client not found: %w", err)
	}

	// Get DNS from settings
	dns, _ := s.settings.GetXrayDNS()
	if dns == "" {
		dns = "8.8.8.8, 1.1.1.1"
	}

	// Get keepalive from settings or use default
	keepalive := 25
	if ka, err := s.settings.GetXrayKeepalive(); err == nil && ka > 0 {
		keepalive = ka
	}

	// Generate config
	cc := &awg2.ClientConfig{
		PrivateKey:          client.PrivateKey,
		Address:             client.AllowedIPs,
		ServerPublicKey:     awg2Inbound.ServerPublicKey,
		ServerEndpoint:      endpoint,
		DNS:                 dns,
		MTU:                 awg2Inbound.MTU,
		AllowedIPs:          "0.0.0.0/0",
		PersistentKeepalive: keepalive,
	}

	config, err := awg2.GenerateClientConfigConf(cc)
	if err != nil {
		logger.Error("[awg2] Config generation failed:", err)
		return "", err
	}

	return config, nil
}

// GetClientQR generates QR code for a client
func (s *AWG2Service) GetClientQR(inboundID int, clientID uint, endpoint string) ([]byte, error) {
	config, err := s.GetClientConfig(inboundID, clientID, endpoint)
	if err != nil {
		return nil, err
	}

	return awg2.GenerateQRCode(config, 400)
}

// UpdateClientTraffic updates client traffic stats
func (s *AWG2Service) UpdateClientTraffic(clientID uint, up, down uint64) error {
	db := database.GetDB()

	var client model.AWG2Client
	if err := db.First(&client, clientID).Error; err != nil {
		return err
	}

	client.UsedTraffic += up + down
	client.Upload += up
	client.Download += down
	client.UpdatedAt = time.Now().UnixMilli()

	return db.Save(&client).Error
}

// GetStats returns traffic statistics for an inbound
func (s *AWG2Service) GetStats(inboundID int) (rx, tx uint64, err error) {
	s.mu.RLock()
	instance, exists := s.instances[inboundID]
	s.mu.RUnlock()

	if !exists {
		return 0, 0, fmt.Errorf("instance not found")
	}

	return instance.GetStats()
}

// ListClients returns all clients for an inbound
func (s *AWG2Service) ListClients(inboundID int) ([]model.AWG2Client, error) {
	db := database.GetDB()

	var clients []model.AWG2Client
	if err := db.Where("inbound_id = (SELECT id FROM awg2_inbounds WHERE inbound_id = ?)", inboundID).Find(&clients).Error; err != nil {
		return nil, err
	}

	return clients, nil
}

// RemoveClient removes a client
func (s *AWG2Service) RemoveClient(clientID uint) error {
	db := database.GetDB()
	return db.Delete(&model.AWG2Client{}, clientID).Error
}

// MarshalSettings marshals AWG2 inbound as settings JSON
func (s *AWG2Service) MarshalSettings(awg2IB *model.AWG2Inbound, clients []model.AWG2Client) (string, error) {
	settings := map[string]interface{}{
		"port":              awg2IB.Port,
		"subnet":            awg2IB.Subnet,
		"mtu":               awg2IB.MTU,
		"outboundInterface": awg2IB.OutboundInterface,
		"disableIPv6":       awg2IB.DisableIPv6,
		"allowedIpsMode":    awg2IB.AllowedIPsMode,
		"serverPublicKey":   awg2IB.ServerPublicKey,
		"serverPrivateKey":  awg2IB.ServerPrivateKey,
		"params": map[string]interface{}{
			"jc":   awg2IB.Jc,
			"jmin": awg2IB.Jmin,
			"jmax": awg2IB.Jmax,
			"s1":   awg2IB.S1,
			"s2":   awg2IB.S2,
			"s3":   awg2IB.S3,
			"s4":   awg2IB.S4,
			"h1":   awg2IB.H1,
			"h2":   awg2IB.H2,
			"h3":   awg2IB.H3,
			"h4":   awg2IB.H4,
			"i1":   awg2IB.I1,
		},
		"clients": clients,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
