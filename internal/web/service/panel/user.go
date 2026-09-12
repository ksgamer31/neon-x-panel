package panel

import (
	"errors"

	"github.com/xlzd/gotp"
	"gorm.io/gorm"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/logger"
	"github.com/ksgamer31/neon-x-panel/v3/internal/util/crypto"
	ldaputil "github.com/ksgamer31/neon-x-panel/v3/internal/util/ldap"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service"
)

// UserService provides business logic for user management and authentication.
// It handles user creation, login, password management, and 2FA operations.
type UserService struct {
	settingService service.SettingService
}

// GetFirstUser retrieves the first user from the database.
// This is typically used for initial setup or when there's only one admin user.
func (s *UserService) GetFirstUser() (*model.User, error) {
	db := database.GetDB()

	user := &model.User{}
	err := db.Model(model.User{}).
		First(user).
		Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) CheckUser(username string, password string, twoFactorCode string) (*model.User, error) {
	db := database.GetDB()

	user := &model.User{}

	err := db.Model(model.User{}).
		Where("username = ?", username).
		First(user).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("invalid credentials")
	} else if err != nil {
		logger.Warning("check user err:", err)
		return nil, err
	}
	if !user.Enabled && user.Role != "" {
		return nil, errors.New("account disabled")
	}

	if !crypto.CheckPasswordHash(user.Password, password) {
		ldapEnabled, _ := s.settingService.GetLdapEnable()
		if !ldapEnabled {
			return nil, errors.New("invalid credentials")
		}

		host, _ := s.settingService.GetLdapHost()
		port, _ := s.settingService.GetLdapPort()
		useTLS, _ := s.settingService.GetLdapUseTLS()
		skipVerify, _ := s.settingService.GetLdapInsecureSkipVerify()
		bindDN, _ := s.settingService.GetLdapBindDN()
		ldapPass, _ := s.settingService.GetLdapPassword()
		baseDN, _ := s.settingService.GetLdapBaseDN()
		userFilter, _ := s.settingService.GetLdapUserFilter()
		userAttr, _ := s.settingService.GetLdapUserAttr()

		cfg := ldaputil.Config{
			Host:               host,
			Port:               port,
			UseTLS:             useTLS,
			InsecureSkipVerify: skipVerify,
			BindDN:             bindDN,
			Password:           ldapPass,
			BaseDN:             baseDN,
			UserFilter:         userFilter,
			UserAttr:           userAttr,
		}
		ok, err := ldaputil.AuthenticateUser(cfg, username, password)
		if err != nil || !ok {
			return nil, errors.New("invalid credentials")
		}
	}

	twoFactorEnable, err := s.settingService.GetTwoFactorEnable()
	if err != nil {
		logger.Warning("check two factor err:", err)
		return nil, err
	}

	if twoFactorEnable {
		twoFactorToken, err := s.settingService.GetTwoFactorToken()
		if err != nil {
			logger.Warning("check two factor token err:", err)
			return nil, err
		}

		if gotp.NewDefaultTOTP(twoFactorToken).Now() != twoFactorCode {
			return nil, errors.New("invalid 2fa code")
		}
	}

	return user, nil
}

func (s *UserService) BumpLoginEpoch() error {
	db := database.GetDB()
	return db.Model(model.User{}).
		Where("1 = 1").
		Update("login_epoch", gorm.Expr("login_epoch + 1")).
		Error
}

func (s *UserService) UpdateUser(id int, username string, password string) error {
	db := database.GetDB()
	hashedPassword, err := crypto.HashPasswordAsBcrypt(password)
	if err != nil {
		return err
	}

	twoFactorEnable, err := s.settingService.GetTwoFactorEnable()
	if err != nil {
		return err
	}

	if twoFactorEnable {
		_ = s.settingService.SetTwoFactorEnable(false)
		_ = s.settingService.SetTwoFactorToken("")
	}

	return db.Model(model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"username":    username,
			"password":    hashedPassword,
			"login_epoch": gorm.Expr("login_epoch + 1"),
		}).
		Error
}

func (s *UserService) UpdateFirstUser(username string, password string) error {
	if username == "" {
		return errors.New("username can not be empty")
	} else if password == "" {
		return errors.New("password can not be empty")
	}
	hashedPassword, er := crypto.HashPasswordAsBcrypt(password)
	if er != nil {
		return er
	}
	db := database.GetDB()
	user := &model.User{}
	err := db.Model(model.User{}).First(user).Error
	if database.IsNotFound(err) {
		user.Username = username
		user.Password = hashedPassword
		if user.Role == "" {
			user.Role = model.RoleOwner
		}
		user.Enabled = true
		return db.Model(model.User{}).Create(user).Error
	} else if err != nil {
		return err
	}
	user.Username = username
	user.Password = hashedPassword
	user.LoginEpoch++
	return db.Save(user).Error
}

// --- Neon X multi-admin RBAC ---

func (s *UserService) DB() *gorm.DB { return database.GetDB() }

func (s *UserService) ListUsers() ([]model.User, error) {
	db := database.GetDB()
	var users []model.User
	if err := db.Order("id asc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *UserService) GetUserByID(id int) (*model.User, error) {
	db := database.GetDB()
	u := &model.User{}
	if err := db.First(u, id).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) CreateUser(username, password, role, displayName, inboundIds string, quotaGB int64) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password required")
	}
	if role == "" {
		role = model.RoleViewer
	}
	if !model.IsValidRole(role) {
		return nil, errors.New("invalid role")
	}
	hashed, err := crypto.HashPasswordAsBcrypt(password)
	if err != nil {
		return nil, err
	}
	db := database.GetDB()
	u := &model.User{Username: username, Password: hashed, Role: role, Enabled: true, DisplayName: displayName, InboundIds: inboundIds, QuotaGB: quotaGB}
	if err := db.Create(u).Error; err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) DeleteUser(id int) error {
	db := database.GetDB()
	// prevent deleting the last owner
	var ownerCount int64
	db.Model(&model.User{}).Where("role = ?", model.RoleOwner).Count(&ownerCount)
	u, err := s.GetUserByID(id)
	if err != nil {
		return err
	}
	if u.Role == model.RoleOwner && ownerCount <= 1 {
		return errors.New("cannot delete the last owner")
	}
	return db.Delete(&model.User{}, id).Error
}

func (s *UserService) UpdateUserRole(id int, role string, enabled *bool, displayName *string, inboundIds *string, quotaGB *int64) error {
	if role != "" && !model.IsValidRole(role) {
		return errors.New("invalid role")
	}
	db := database.GetDB()
	updates := map[string]any{}
	if role != "" {
		updates["role"] = role
	}
	if enabled != nil {
		updates["enabled"] = *enabled
		if !*enabled {
			updates["login_epoch"] = gorm.Expr("login_epoch + 1")
		}
	}
	if displayName != nil {
		updates["display_name"] = *displayName
	}
	if inboundIds != nil {
		updates["inbound_ids"] = *inboundIds
	}
	if quotaGB != nil {
		if *quotaGB < 0 {
			return errors.New("quota must be >= 0")
		}
		updates["quota_gb"] = *quotaGB
	}
	if len(updates) == 0 {
		return nil
	}
	return db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *UserService) UpdateUserPassword(id int, newPassword string) error {
	if newPassword == "" {
		return errors.New("password can not be empty")
	}
	hashed, err := crypto.HashPasswordAsBcrypt(newPassword)
	if err != nil {
		return err
	}
	db := database.GetDB()
	return db.Model(&model.User{}).Where("id = ?", id).Updates(map[string]any{"password": hashed, "login_epoch": gorm.Expr("login_epoch + 1")}).Error
}

