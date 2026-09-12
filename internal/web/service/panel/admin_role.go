package panel

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"gorm.io/gorm"
)

type AdminRoleService struct{}

type AdminRolePayload struct {
	Name     string         `json:"name"`
	BaseTier string         `json:"baseTier"`
	QuotaGB  int64          `json:"quotaGB"`
	Perms    map[string]any `json:"permissions"`
	Limits   map[string]any `json:"limits"`
	Features map[string]any `json:"features"`
	Access   map[string]any `json:"access"`
}

type AdminRoleView struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	BaseTier   string `json:"baseTier"`
	BuiltIn    bool   `json:"builtIn"`
	OwnerRole  bool   `json:"ownerRole"`
	QuotaGB    int64  `json:"quotaGB"`
	Perms      any    `json:"permissions"`
	Limits     any    `json:"limits"`
	Features   any    `json:"features"`
	Access     any    `json:"access"`
	AdminCount int64  `json:"adminCount"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

func slugify(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	dash := false
	for _, r := range name {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
		if !ok {
			if !dash {
				b.WriteRune('-')
				dash = true
			}
			continue
		}
		if r == '-' {
			if dash {
				continue
			}
			dash = true
		} else {
			dash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func mapToJSON(v map[string]any) string {
	if len(v) == 0 {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func jsonToAny(raw string) any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	var out any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]any{}
	}
	return out
}

func roleToView(r *model.AdminRole, n int64) *AdminRoleView {
	if r == nil {
		return nil
	}
	return &AdminRoleView{
		Id: r.Id, Name: r.Name, Slug: r.Slug, BaseTier: r.BaseTier,
		BuiltIn: r.BuiltIn, OwnerRole: r.OwnerRole, QuotaGB: r.QuotaGB,
		Perms: jsonToAny(r.PermsJSON), Limits: jsonToAny(r.LimitsJSON),
		Features: jsonToAny(r.FeaturesJSON), Access: jsonToAny(r.AccessJSON),
		AdminCount: n, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func (s *AdminRoleService) List() ([]*AdminRoleView, error) {
	db := database.GetDB()
	var rows []*model.AdminRole
	if err := db.Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	counts := map[int]int64{}
	if len(rows) > 0 {
		type cr struct {
			RoleID int
			Count  int64
		}
		var g []cr
		ids := make([]int, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.Id)
		}
		_ = db.Model(&model.User{}).Select("role_id as role_id, COUNT(*) as count").Where("role_id IN ?", ids).Group("role_id").Scan(&g).Error
		for _, x := range g {
			counts[x.RoleID] = x.Count
		}
	}
	out := make([]*AdminRoleView, 0, len(rows))
	for _, r := range rows {
		out = append(out, roleToView(r, counts[r.Id]))
	}
	return out, nil
}

func (s *AdminRoleService) Get(id int) (*AdminRoleView, error) {
	db := database.GetDB()
	var r model.AdminRole
	if err := db.Where("id = ?", id).First(&r).Error; err != nil {
		return nil, err
	}
	var n int64
	_ = db.Model(&model.User{}).Where("role_id = ?", id).Count(&n).Error
	return roleToView(&r, n), nil
}

// GetRole resolves DB role row for a user (nil when legacy-only).
func GetRoleForUser(u *model.User) *model.AdminRole {
	if u == nil || u.RoleID <= 0 {
		return nil
	}
	var r model.AdminRole
	if err := database.GetDB().Where("id = ?", u.RoleID).First(&r).Error; err != nil {
		return nil
	}
	return &r
}

// EffectiveTierFor resolves enforcement tier via DB role, falling back to legacy string.
func EffectiveTierFor(u *model.User) string {
	return model.EffectiveTier(u, GetRoleForUser(u))
}

// RoleHasFeature checks a bool-ish feature flag on the DB role (e.g. blockLimitedAdmins).
func RoleHasFeature(u *model.User, key string) bool {
	r := GetRoleForUser(u)
	if r == nil || r.FeaturesJSON == "" {
		return true // legacy roles: keep old behavior (enforce)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(r.FeaturesJSON), &m); err != nil {
		return true
	}
	v, ok := m[key]
	if !ok {
		return true
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x == "true" || x == "1" || x == "yes"
	case float64:
		return x != 0
	default:
		return true
	}
}

func (s *AdminRoleService) Create(p AdminRolePayload) (*AdminRoleView, error) {
	name := strings.TrimSpace(p.Name)
	if name == "" {
		return nil, errors.New("role name is required")
	}
	tier := p.BaseTier
	if !model.IsValidRole(tier) {
		tier = model.RoleViewer
	}
	if tier == model.RoleOwner {
		return nil, errors.New("cannot create another owner tier; assign Owner role instead")
	}
	slug := slugify(name)
	if slug == "" {
		return nil, errors.New("role slug is invalid")
	}
	db := database.GetDB()
	var n int64
	_ = db.Model(&model.AdminRole{}).Where("name = ? OR slug = ?", name, slug).Count(&n).Error
	if n > 0 {
		return nil, errors.New("role already exists")
	}
	row := &model.AdminRole{
		Name: name, Slug: slug, BaseTier: tier, QuotaGB: p.QuotaGB,
		PermsJSON: mapToJSON(p.Perms), LimitsJSON: mapToJSON(p.Limits),
		FeaturesJSON: mapToJSON(nonEmptyFeat(p.Features)), AccessJSON: mapToJSON(p.Access),
	}
	if err := db.Create(row).Error; err != nil {
		return nil, err
	}
	return roleToView(row, 0), nil
}

func nonEmptyFeat(v map[string]any) map[string]any {
	if len(v) > 0 {
		return v
	}
	return map[string]any{"blockLimitedAdmins": true, "disconnectUsersWhenLimited": true}
}

func (s *AdminRoleService) Update(id int, p AdminRolePayload) (*AdminRoleView, error) {
	db := database.GetDB()
	var row model.AdminRole
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		return nil, err
	}
	if row.OwnerRole {
		return nil, errors.New("owner role is read-only")
	}
	updates := map[string]any{}
	if p.BaseTier != "" {
		if !model.IsValidRole(p.BaseTier) || p.BaseTier == model.RoleOwner {
			return nil, errors.New("invalid base tier")
		}
		updates["base_tier"] = p.BaseTier
	}
	if p.QuotaGB >= 0 && p.Perms != nil {
		// quota update bundled with full update; allow 0
	}
	updates["quota_gb"] = p.QuotaGB
	updates["permissions"] = mapToJSON(p.Perms)
	updates["limits"] = mapToJSON(p.Limits)
	updates["features"] = mapToJSON(p.Features)
	updates["access"] = mapToJSON(p.Access)
	if !row.BuiltIn {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return nil, errors.New("role name is required")
		}
		slug := slugify(name)
		var n int64
		_ = db.Model(&model.AdminRole{}).Where("(name = ? OR slug = ?) AND id <> ?", name, slug, id).Count(&n).Error
		if n > 0 {
			return nil, errors.New("role already exists")
		}
		updates["name"] = name
		updates["slug"] = slug
	}
	if err := db.Model(&model.AdminRole{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *AdminRoleService) Duplicate(id int) (*AdminRoleView, error) {
	db := database.GetDB()
	var src model.AdminRole
	if err := db.Where("id = ?", id).First(&src).Error; err != nil {
		return nil, err
	}
	base := strings.TrimSpace(src.Name) + " (copy)"
	name := base
	for i := 2; ; i++ {
		slug := slugify(name)
		var n int64
		_ = db.Model(&model.AdminRole{}).Where("name = ? OR slug = ?", name, slug).Count(&n).Error
		if n == 0 {
			row := &model.AdminRole{
				Name: name, Slug: slug, BaseTier: src.BaseTier, QuotaGB: src.QuotaGB,
				PermsJSON: src.PermsJSON, LimitsJSON: src.LimitsJSON,
				FeaturesJSON: src.FeaturesJSON, AccessJSON: src.AccessJSON,
			}
			if err := db.Create(row).Error; err != nil {
				return nil, err
			}
			return roleToView(row, 0), nil
		}
		name = base + " " + itoa(i)
	}
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}

func (s *AdminRoleService) Delete(id int) error {
	db := database.GetDB()
	var row model.AdminRole
	if err := db.Where("id = ?", id).First(&row).Error; err != nil {
		return err
	}
	if row.OwnerRole || row.BuiltIn {
		return errors.New("built-in roles cannot be deleted")
	}
	var n int64
	_ = db.Model(&model.User{}).Where("role_id = ?", id).Count(&n).Error
	if n > 0 {
		return errors.New("role is assigned to admins")
	}
	return db.Where("id = ?", id).Delete(&model.AdminRole{}).Error
}

var _ = gorm.ErrRecordNotFound
