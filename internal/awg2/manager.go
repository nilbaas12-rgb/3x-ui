package awg2

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/mhsanaei/3x-ui/v3/internal/logger"
)

// Manager manages AWG2 sidecar processes
type Manager struct {
	mu        sync.RWMutex
	instances map[string]*Instance
	binPath   string
}

// Instance represents a running AWG2 instance
type Instance struct {
	Tag       string
	Port      int
	ConfPath  string
	Cmd       *exec.Cmd
	Running   bool
}

// NewManager creates a new AWG2 manager
func NewManager(binPath string) *Manager {
	if binPath == "" {
		binPath = "/usr/local/bin/amneziawg"
	}
	return &Manager{
		instances: make(map[string]*Instance),
		binPath:   binPath,
	}
}

// Start starts an AWG2 instance
func (m *Manager) Start(instance *Instance) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.instances[instance.Tag]; exists {
		return fmt.Errorf("instance %s already exists", instance.Tag)
	}

	// Verify configuration file exists
	if _, err := os.Stat(instance.ConfPath); err != nil {
		return fmt.Errorf("config file not found: %w", err)
	}

	// Start process (placeholder - actual implementation depends on AWG binary)
	logger.Info("[awg2] Starting instance", instance.Tag, "with config", instance.ConfPath)

	instance.Running = true
	m.instances[instance.Tag] = instance

	return nil
}

// Stop stops an AWG2 instance
func (m *Manager) Stop(tag string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	instance, exists := m.instances[tag]
	if !exists {
		return fmt.Errorf("instance %s not found", tag)
	}

	if instance.Cmd != nil && instance.Cmd.Process != nil {
		if err := instance.Cmd.Process.Kill(); err != nil {
			logger.Warning("[awg2] Failed to kill process", tag, ":", err)
		}
	}

	instance.Running = false
	logger.Info("[awg2] Stopped instance", tag)

	return nil
}

// Restart restarts an AWG2 instance
func (m *Manager) Restart(instance *Instance) error {
	if err := m.Stop(instance.Tag); err != nil {
		logger.Warning("[awg2] Error stopping instance before restart:", err)
	}

	return m.Start(instance)
}

// GetInstance gets an instance by tag
func (m *Manager) GetInstance(tag string) (*Instance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instance, exists := m.instances[tag]
	return instance, exists
}

// ListInstances returns all instances
func (m *Manager) ListInstances() []*Instance {
	m.mu.RLock()
	defer m.mu.RUnlock()

	instances := make([]*Instance, 0, len(m.instances))
	for _, instance := range m.instances {
		instances = append(instances, instance)
	}
	return instances
}
