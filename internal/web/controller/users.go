package controller

import (
	"net/http"
	"strconv"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/middleware"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service/panel"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	userService panel.UserService
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
	// management: only owner/admin
	mgmt := grp.Group("")
	mgmt.Use(middleware.RequireRole(model.RoleOwner, model.RoleAdmin))
	mgmt.POST("/list", a.list)
	mgmt.POST("/create", a.create)
	mgmt.POST("/update/:id", a.update)
	mgmt.POST("/delete/:id", a.deleteUser)
	mgmt.POST("/resetPassword/:id", a.resetPassword)
}

func (a *UsersController) me(c *gin.Context) {
	u := session.GetLoginUser(c)
	if u == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "unauthorized"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{
		"id": u.Id, "username": u.Username, "role": u.Role, "enabled": u.Enabled, "displayName": u.DisplayName, "inboundIds": u.InboundIds, "quotaGB": u.QuotaGB,
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
		Id          int    `json:"id"`
		Username    string `json:"username"`
		Role        string `json:"role"`
		Enabled     bool   `json:"enabled"`
		DisplayName string `json:"displayName"`
		InboundIds  string `json:"inboundIds"`
		QuotaGB     int64  `json:"quotaGB"`
		QuotaUsed   int64  `json:"quotaUsed"`
		QuotaPct    int    `json:"quotaPct"`
	}
	// Compute per-admin traffic usage for quota display (sum of client_traffics for clients created by each admin).
	quotaUsedByUser := map[int]int64{}
	if len(users) > 0 {
		type row struct { CreatedBy int `gorm:"column:created_by"`; Used int64 `gorm:"column:used"` }
		var rows []row
		// Only count traffic for admins that have clients; missing admins stay 0.
		_ = a.userService.DB().Table("clients c").Select("c.created_by as created_by, COALESCE(SUM(ct.up + ct.down),0) as used").Joins("JOIN client_traffics ct ON ct.email = c.email").Group("c.created_by").Scan(&rows).Error
		for _, r := range rows {
			quotaUsedByUser[r.CreatedBy] = r.Used
		}
	}
	out := make([]safeUser, 0, len(users))
	for _, u := range users {
		if u.Role == "" { u.Role = model.RoleOwner }
		used := quotaUsedByUser[u.Id]
		pct := 0
		if u.QuotaGB > 0 {
			quotaBytes := u.QuotaGB * 1024 * 1024 * 1024
			if quotaBytes > 0 {
				pct = int(used * 100 / quotaBytes)
				if pct > 100 { pct = 100 }
			}
		}
		out = append(out, safeUser{u.Id, u.Username, u.Role, u.Enabled, u.DisplayName, u.InboundIds, u.QuotaGB, used, pct})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": out})
}

type createUserForm struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Role        string `json:"role"`
	DisplayName string `json:"displayName"`
	InboundIds  string `json:"inboundIds"`
	QuotaGB     int64  `json:"quotaGB"`
}

func (a *UsersController) create(c *gin.Context) {
	var f createUserForm
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	// only owner can create owner/admin
	caller := session.GetLoginUser(c)
	if caller.Role != model.RoleOwner && (f.Role == model.RoleOwner || f.Role == model.RoleAdmin) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can create owner/admin"})
		return
	}
	u, err := a.userService.CreateUser(f.Username, f.Password, f.Role, f.DisplayName, f.InboundIds, f.QuotaGB)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": gin.H{"id": u.Id, "username": u.Username, "role": u.Role}})
}

type updateUserForm2 struct {
	Role        string  `json:"role"`
	Enabled     *bool   `json:"enabled"`
	DisplayName *string `json:"displayName"`
	InboundIds  *string `json:"inboundIds"`
	QuotaGB     *int64  `json:"quotaGB"`
}

func (a *UsersController) update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var f updateUserForm2
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	caller := session.GetLoginUser(c)
	target, err := a.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "user not found"})
		return
	}
	if caller.Role != model.RoleOwner && (target.Role == model.RoleOwner || f.Role == model.RoleOwner) {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can modify owner"})
		return
	}
	if err := a.userService.UpdateUserRole(id, f.Role, f.Enabled, f.DisplayName, f.InboundIds, f.QuotaGB); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (a *UsersController) deleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	caller := session.GetLoginUser(c)
	target, err := a.userService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "not found"})
		return
	}
	if caller.Id == id {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "cannot delete yourself"})
		return
	}
	if caller.Role != model.RoleOwner && target.Role == model.RoleOwner {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "only owner can delete owner"})
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
	if caller.Id != id && caller.Role != model.RoleOwner && caller.Role != model.RoleAdmin {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": "forbidden"})
		return
	}
	if err := a.userService.UpdateUserPassword(id, f.Password); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
