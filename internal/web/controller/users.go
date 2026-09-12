package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/middleware"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service/panel"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"
)

type UsersController struct {
	userService   panel.UserService
	clientService service.ClientService
	inboundSvc    service.InboundService
	xraySvc       service.XrayService
}

func NewUsersController(g *gin.RouterGroup) *UsersController {
	a := &UsersController{}
	a.initRouter(g)
	return a
}

func (a *UsersController) initRouter(g *gin.RouterGroup) {
	grp := g.Group("/users")
	// /me is available to any logged-in user (for RBAC UI)
	grp.GET("/me", a.me)
	// stats for admin header cards
	grp.GET("/stats", a.stats)
	// management: only owner/admin
	mgmt := grp.Group("")
	mgmt.Use(middleware.RequireRole(model.RoleOwner, model.RoleAdmin))
	mgmt.POST("/list", a.list)
	mgmt.POST("/create", a.create)
	mgmt.POST("/update/:id", a.update)
	mgmt.POST("/delete/:id", a.deleteUser)
	mgmt.POST("/resetPassword/:id", a.resetPassword)
	mgmt.POST("/resetUsage/:id", a.resetUsage)
	mgmt.POST("/clients/enable/:id", a.enableClients)
	mgmt.POST("/clients/disable/:id", a.disableClients)
	mgmt.POST("/clients/remove/:id", a.removeClients)
}

func (a *UsersController) me(c *gin.Context) {
	u := session.GetLoginUser(c)
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "unauthorized"})
		return
	}
	roleID := u.RoleID
	roleName := u.Role
	if r := panel.GetRoleForUser(u); r != nil {
		roleName = r.Name
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{
		"id": u.Id, "username": u.Username, "role": panel.EffectiveTierFor(u),
		"roleName": roleName, "roleId": roleID,
		"enabled": u.Enabled, "displayName": u.DisplayName, "inboundIds": u.InboundIds,
		"quotaGB": u.QuotaGB, "telegramId": u.TelegramID, "supportUrl": u.SupportURL,
		"profileTitle": u.ProfileTitle, "subDomain": u.SubDomain, "note": u.Note,
	}})
}

func (a *UsersController) stats(c *gin.Context) {
	total, active, disabled, limited := a.userService.AdminStats()
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{
		"totalAdmins": total, "activeAdmins": active,
		"disabledAdmins": disabled, "limitedAdmins": limited,
	}})
}

func (a *UsersController) list(c *gin.Context) {
	users, err := a.userService.ListUsers()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	// hide password hash
	type safeUser struct {
		Id           int    `json:"id"`
		Username     string `json:"username"`
		Role         string `json:"role"`
		RoleID       int    `json:"roleId"`
		RoleName     string `json:"roleName"`
		Enabled      bool   `json:"enabled"`
		DisplayName  string `json:"displayName"`
		InboundIds   string `json:"inboundIds"`
		QuotaGB      int64  `json:"quotaGB"`
		QuotaUsed    int64  `json:"quotaUsed"`
		QuotaPct     int    `json:"quotaPct"`
		TotalClients int64  `json:"totalClients"`
		TelegramID   string `json:"telegramId"`
		SupportURL   string `json:"supportUrl"`
		ProfileTitle string `json:"profileTitle"`
		SubDomain    string `json:"subDomain"`
		Note         string `json:"note"`
		CreatedAt    int64  `json:"createdAt"`
		UpdatedAt    int64  `json:"updatedAt"`
	}
	// Compute per-admin traffic usage for quota display (sum of client_traffics for clients created by each admin).
	quotaUsedByUser := map[int]int64{}
	clientCountByUser := map[int]int64{}
	if len(users) > 0 {
		type row struct {
			CreatedBy int   `gorm:"column:created_by"`
			Used      int64 `gorm:"column:used"`
		}
		var rows []row
		// Only count traffic for admins that have clients; missing admins stay 0.
		_ = a.userService.DB().Table("clients c").Select("c.created_by as created_by, COALESCE(SUM(ct.up + ct.down),0) as used").Joins("JOIN client_traffics ct ON ct.email = c.email").Group("c.created_by").Scan(&rows).Error
		for _, r := range rows {
			quotaUsedByUser[r.CreatedBy] = r.Used
		}
		type crow struct {
			CreatedBy int   `gorm:"column:created_by"`
			N         int64 `gorm:"column:n"`
		}
		var crows []crow
		_ = a.userService.DB().Table("clients").Select("created_by, COUNT(*) as n").Group("created_by").Scan(&crows).Error
		for _, r := range crows {
			clientCountByUser[r.CreatedBy] = r.N
		}
	}
	// resolve role names
	roleNames := map[int]string{}
	for _, u := range users {
		if u.RoleID > 0 {
			if _, ok := roleNames[u.RoleID]; !ok {
				if rv, err := (&panel.AdminRoleService{}).Get(u.RoleID); err == nil && rv != nil {
					roleNames[u.RoleID] = rv.Name
				}
			}
		}
	}
	out := make([]safeUser, 0, len(users))
	for _, u := range users {
		if u.Role == "" {
			u.Role = model.RoleOwner
		}
		used := quotaUsedByUser[u.Id]
		pct := 0
		if u.QuotaGB > 0 {
			quotaBytes := u.QuotaGB * 1024 * 1024 * 1024
			if quotaBytes > 0 {
				pct = int(used * 100 / quotaBytes)
				if pct > 100 {
					pct = 100
				}
			}
		}
		rn := roleNames[u.RoleID]
		if rn == "" {
			rn = u.Role
		}
		out = append(out, safeUser{u.Id, u.Username, panel.EffectiveTierFor(&u), u.RoleID, rn, u.Enabled, u.DisplayName, u.InboundIds, u.QuotaGB, used, pct, clientCountByUser[u.Id], u.TelegramID, u.SupportURL, u.ProfileTitle, u.SubDomain, u.Note, u.CreatedAt, u.UpdatedAt})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "obj": out})
}

type createUserForm struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Role         string `json:"role"`
	RoleID       int    `json:"roleId"`
	DisplayName  string `json:"displayName"`
	InboundIds   string `json:"inboundIds"`
	QuotaGB      int64  `json:"quotaGB"`
	TelegramID   string `json:"telegramId"`
	SupportURL   string `json:"supportUrl"`
	ProfileTitle string `json:"profileTitle"`
	SubDomain    string `json:"subDomain"`
	Note         string `json:"note"`
}

func (a *UsersController) create(c *gin.Context) {
	var f createUserForm
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	// only owner can create owner/admin
	caller := session.GetLoginUser(c)
	callerTier := panel.EffectiveTierFor(caller)
	if callerTier != model.RoleOwner && (f.Role == model.RoleOwner || f.Role == model.RoleAdmin) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can create owner/admin"})
		return
	}
	u, err := a.userService.CreateUserFull(f.Username, f.Password, f.Role, f.RoleID, f.DisplayName, f.InboundIds, f.QuotaGB, f.TelegramID, f.SupportURL, f.ProfileTitle, f.SubDomain, f.Note)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"id": u.Id, "username": u.Username, "role": u.Role, "roleId": u.RoleID}})
}

type updateUserForm2 struct {
	Role         string  `json:"role"`
	RoleID       *int    `json:"roleId"`
	Enabled      *bool   `json:"enabled"`
	DisplayName  *string `json:"displayName"`
	InboundIds   *string `json:"inboundIds"`
	QuotaGB      *int64  `json:"quotaGB"`
	TelegramID   *string `json:"telegramId"`
	SupportURL   *string `json:"supportUrl"`
	ProfileTitle *string `json:"profileTitle"`
	SubDomain    *string `json:"subDomain"`
	Note         *string `json:"note"`
}

func (a *UsersController) update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f updateUserForm2
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	caller := session.GetLoginUser(c)
	callerTier := panel.EffectiveTierFor(caller)
	target, err := a.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "user not found"})
		return
	}
	targetTier := panel.EffectiveTierFor(target)
	if callerTier != model.RoleOwner && (targetTier == model.RoleOwner || f.Role == model.RoleOwner) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can modify owner"})
		return
	}
	if err := a.userService.UpdateUserFull(id, f.Role, f.RoleID, f.Enabled, f.DisplayName, f.InboundIds, f.QuotaGB, f.TelegramID, f.SupportURL, f.ProfileTitle, f.SubDomain, f.Note); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	// If quota raised above usage, re-enable eligible owned clients (Heimdall-style recovery).
	if f.QuotaGB != nil && *f.QuotaGB >= 0 {
		if u2, err := a.userService.GetUserByID(id); err == nil && u2.Enabled {
			var used int64
			_ = a.userService.DB().Table("clients c").Select("COALESCE(SUM(ct.up + ct.down),0)").Joins("JOIN client_traffics ct ON ct.email = c.email").Where("c.created_by = ?", id).Scan(&used).Error
			if *f.QuotaGB == 0 || used < *f.QuotaGB*1024*1024*1024 {
				var emails []string
				_ = a.userService.DB().Model(&model.ClientRecord{}).Where("created_by = ?", id).Pluck("email", &emails).Error
				if len(emails) > 0 {
					if _, needRestart, err := a.clientService.BulkSetEnable(&a.inboundSvc, emails, true); err == nil && needRestart {
						a.xraySvc.SetToNeedRestart()
					}
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *UsersController) deleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	caller := session.GetLoginUser(c)
	callerTier := panel.EffectiveTierFor(caller)
	target, err := a.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "not found"})
		return
	}
	if caller.Id == id {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "cannot delete yourself"})
		return
	}
	if callerTier != model.RoleOwner && panel.EffectiveTierFor(target) == model.RoleOwner {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can delete owner"})
		return
	}
	if n := a.userService.CountAdminClients(id); n > 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "admin owns clients and cannot be deleted"})
		return
	}
	if err := a.userService.DeleteUser(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type resetPwForm struct {
	Password string `json:"password" binding:"required"`
}

func (a *UsersController) resetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f resetPwForm
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	caller := session.GetLoginUser(c)
	callerTier := panel.EffectiveTierFor(caller)
	if caller.Id != id && callerTier != model.RoleOwner && callerTier != model.RoleAdmin {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "forbidden"})
		return
	}
	if err := a.userService.UpdateUserPassword(id, f.Password); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *UsersController) resetUsage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if _, err := a.userService.GetUserByID(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "not found"})
		return
	}
	err := a.userService.ResetAdminUsage(
		func(emails []string) (int, error) {
			return a.clientService.BulkResetTraffic(&a.inboundSvc, emails)
		},
		func() { a.xraySvc.SetToNeedRestart() },
		id,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *UsersController) enableClients(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	n, err := a.userService.SetAdminClientsEnable(
		func(emails []string, enable bool) (int, bool, error) {
			r, needRestart, err := a.clientService.BulkSetEnable(&a.inboundSvc, emails, enable)
			return r.Changed, needRestart, err
		},
		func() { a.xraySvc.SetToNeedRestart() },
		id, true,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"count": n}})
}

func (a *UsersController) disableClients(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	n, err := a.userService.SetAdminClientsEnable(
		func(emails []string, enable bool) (int, bool, error) {
			r, needRestart, err := a.clientService.BulkSetEnable(&a.inboundSvc, emails, enable)
			return r.Changed, needRestart, err
		},
		func() { a.xraySvc.SetToNeedRestart() },
		id, false,
	)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"count": n}})
}

func (a *UsersController) removeClients(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var emails []string
	_ = a.userService.DB().Model(&model.ClientRecord{}).Where("created_by = ?", id).Pluck("email", &emails).Error
	if len(emails) == 0 {
		c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"count": 0}})
		return
	}
	res, needRestart, err := a.clientService.BulkDelete(&a.inboundSvc, emails, false)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	if needRestart {
		a.xraySvc.SetToNeedRestart()
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"count": res.Deleted}})
}
