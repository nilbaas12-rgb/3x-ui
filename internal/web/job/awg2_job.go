package job

import (
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service"
	"github.com/mhsanaei/3x-ui/v3/internal/xray"
)

// AWG2Job manages AWG2 instances and collects traffic statistics
type AWG2Job struct {
	inboundService service.InboundService
}

// NewAWG2Job creates a new AWG2 job instance
func NewAWG2Job() *AWG2Job {
	return &AWG2Job{}
}

// Run reconciles AWG2 instances and records traffic
func (j *AWG2Job) Run() {
	logger.Debug("[awg2] Running AWG2 reconciliation job")

	// TODO: Implement AWG2 instance reconciliation
	// 1. Get all enabled AWG2 inbounds from DB
	// 2. Start/stop instances as needed
	// 3. Collect traffic statistics
	// 4. Update client traffics in DB
	// 5. Handle per-client traffic and online status

	traffics := make([]*xray.Traffic, 0)
	clientTraffics := make([]*xray.ClientTraffic, 0)

	if len(traffics) > 0 || len(clientTraffics) > 0 {
		if _, _, err := j.inboundService.AddTraffic(traffics, clientTraffics); err != nil {
			logger.Warning("[awg2] Failed to add traffic:", err)
		}
	}

	logger.Debug("[awg2] AWG2 job completed")
}
