package service

import (
	"encoding/json"
	"strings"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/xray"
)

// Creator isolation (Neon X v1.2.1): a creator-tier admin sees and touches
// only the clients it created itself (clients.created_by = users.id).
// Every client/group read takes an optional createdBy scope; nil = unscoped
// (owner/admin/editor/viewer keep the full panel view). Legacy rows with
// created_by = 0 belong to no creator, so creators never see them.

// OwnerUserID returns the first owner account id (inbounds belong to it).
func OwnerUserID() int {
	var id int
	if err := database.GetDB().Model(&model.User{}).
		Where("role = ?", model.RoleOwner).
		Order("id ASC").Limit(1).
		Pluck("id", &id).Error; err != nil || id == 0 {
		return 1
	}
	return id
}

// OwnedClientEmails returns the email set created by one admin.
func OwnedClientEmails(createdBy int) (map[string]struct{}, error) {
	var emails []string
	if err := database.GetDB().Model(&model.ClientRecord{}).
		Where("created_by = ?", createdBy).
		Pluck("email", &emails).Error; err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(emails))
	for _, e := range emails {
		if e != "" {
			out[e] = struct{}{}
		}
	}
	return out, nil
}

// GetRecordBySubID loads a client row by its subscription id.
func (s *ClientService) GetRecordBySubID(subID string) (*model.ClientRecord, error) {
	row := &model.ClientRecord{}
	if err := database.GetDB().Where("sub_id = ?", subID).First(row).Error; err != nil {
		return nil, err
	}
	return row, nil
}


// FilterInboundClientsToOwned strips every client a creator did not create
// from inbound settings payloads and traffic stats, in place.
func FilterInboundClientsToOwned(inbounds []*model.Inbound, owned map[string]struct{}) {
	for _, ib := range inbounds {
		if ib == nil {
			continue
		}
		if strings.TrimSpace(ib.Settings) != "" {
			var raw map[string]any
			if err := json.Unmarshal([]byte(ib.Settings), &raw); err == nil {
				if arr, ok := raw["clients"].([]any); ok {
					kept := make([]any, 0, len(arr))
					for _, e := range arr {
						m, ok := e.(map[string]any)
						if !ok {
							continue
						}
						em, _ := m["email"].(string)
						if _, ok := owned[em]; ok {
							kept = append(kept, e)
						}
					}
					raw["clients"] = kept
					if out, err := json.Marshal(raw); err == nil {
						ib.Settings = string(out)
					}
				}
			}
		}
		if len(ib.ClientStats) > 0 {
			kept := make([]xray.ClientTraffic, 0, len(ib.ClientStats))
			for _, st := range ib.ClientStats {
				if _, ok := owned[st.Email]; ok {
					kept = append(kept, st)
				}
			}
			ib.ClientStats = kept
		}
	}
}
