package migrations

import (
	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"gorm.io/gorm"
)

// MigrateAWG2Tables creates AWG2 tables if they don't exist
func MigrateAWG2Tables(db *gorm.DB) error {
	logger.Info("[awg2] Migrating AWG2 tables...")

	// Create AWG2 inbounds table
	if !db.Migrator().HasTable(&model.AWG2Inbound{}) {
		if err := db.Migrator().CreateTable(&model.AWG2Inbound{}); err != nil {
			logger.Error("[awg2] Failed to create AWG2Inbound table:", err)
			return err
		}
		logger.Info("[awg2] Created AWG2Inbound table")
	}

	// Create AWG2 clients table
	if !db.Migrator().HasTable(&model.AWG2Client{}) {
		if err := db.Migrator().CreateTable(&model.AWG2Client{}); err != nil {
			logger.Error("[awg2] Failed to create AWG2Client table:", err)
			return err
		}
		logger.Info("[awg2] Created AWG2Client table")
	}

	// Add indexes
	db.Migrator().CreateIndex(&model.AWG2Inbound{}, "inbound_id")
	db.Migrator().CreateIndex(&model.AWG2Client{}, "inbound_id")
	db.Migrator().CreateIndex(&model.AWG2Client{}, "email")

	return nil
}
