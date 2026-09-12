package job

import (
	"github.com/ksgamer31/neon-x-panel/v3/internal/database"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/logger"
	"gorm.io/gorm"
)

// AdminQuotaJob enforces per-admin traffic quotas: when the sum of up+down
// across all clients created by an admin reaches quotaGB, the admin account
// and all of those clients are disabled automatically.
// A quotaGB of 0 means unlimited (no enforcement).
// Re-enabling is manual via the Admins page after increasing the quota
// or resetting traffic.
type AdminQuotaJob struct{}

func NewAdminQuotaJob() *AdminQuotaJob { return &AdminQuotaJob{} }

func (j *AdminQuotaJob) Run() {
	db := database.GetDB()
	if db == nil {
		return
	}
	var admins []model.User
	if err := db.Where("quota_gb > 0 AND enabled = ?", true).Find(&admins).Error; err != nil {
		logger.Warningf("admin quota: list admins failed: %v", err)
		return
	}
	for _, adm := range admins {
		quotaBytes := adm.QuotaGB * 1024 * 1024 * 1024
		if quotaBytes <= 0 {
			continue
		}
		// Sum traffic for all clients owned by this admin.
		var used int64
		// clients.created_by -> client_traffics.email -> up+down
		err := db.Table("clients c").
			Select("COALESCE(SUM(ct.up + ct.down), 0)").
			Joins("JOIN client_traffics ct ON ct.email = c.email").
			Where("c.created_by = ?", adm.Id).
			Scan(&used).Error
		if err != nil {
			logger.Warningf("admin quota: sum traffic for %s failed: %v", adm.Username, err)
			continue
		}
		if used < quotaBytes {
			continue
		}
		logger.Infof("admin quota: %s exceeded %d GB (used %d bytes) — disabling admin and %d-owned clients", adm.Username, adm.QuotaGB, used, adm.Id)
		if err := disableAdminAndClients(db, &adm); err != nil {
			logger.Warningf("admin quota: disable %s failed: %v", adm.Username, err)
		}
	}
}

func disableAdminAndClients(db *gorm.DB, adm *model.User) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// Disable admin account + bump login_epoch so active sessions are invalidated.
		if err := tx.Model(&model.User{}).Where("id = ?", adm.Id).Updates(map[string]any{
			"enabled":     false,
			"login_epoch": gorm.Expr("login_epoch + 1"),
		}).Error; err != nil {
			return err
		}
		// Disable all client_records owned by this admin (those still enabled).
		if err := tx.Model(&model.ClientRecord{}).Where("created_by = ? AND enable = ?", adm.Id, true).Update("enable", false).Error; err != nil {
			return err
		}
		// Also flip client_traffics.enable so Xray stops serving them even before next restart.
		// Find owned emails first (chunked to avoid huge IN).
		var emails []string
		if err := tx.Model(&model.ClientRecord{}).Where("created_by = ?", adm.Id).Pluck("email", &emails).Error; err != nil {
			return err
		}
		const chunk = 400
		for i := 0; i < len(emails); i += chunk {
			end := i + chunk
			if end > len(emails) {
				end = len(emails)
			}
			if err := tx.Table("client_traffics").Where("email IN ?", emails[i:end]).Update("enable", false).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
