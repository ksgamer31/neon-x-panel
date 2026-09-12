package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/middleware"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service/panel"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"
)

type RolesController struct {
	roleService panel.AdminRoleService
}

func NewRolesController(g *gin.RouterGroup) *RolesController {
	a := &RolesController{}
	a.initRouter(g)
	return a
}

func (a *RolesController) initRouter(g *gin.RouterGroup) {
	grp := g.Group("/roles")
	grp.GET("/list", a.list)
	grp.GET("/get/:id", a.get)
	mgmt := grp.Group("")
	mgmt.Use(middleware.RequireRole(model.RoleOwner))
	mgmt.POST("/create", a.create)
	mgmt.POST("/update/:id", a.update)
	mgmt.POST("/duplicate/:id", a.duplicate)
	mgmt.POST("/remove/:id", a.removeRole)
}

func roleOf(c *gin.Context) string {
	if u := session.GetLoginUser(c); u != nil {
		return panel.EffectiveTierFor(u)
	}
	return ""
}

func (a *RolesController) list(c *gin.Context) {
	tier := roleOf(c)
	if tier != model.RoleOwner && tier != model.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"success": false, "msg": "forbidden: insufficient role"})
		return
	}
	rows, err := a.roleService.List()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": rows})
}

func (a *RolesController) get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	row, err := a.roleService.Get(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func (a *RolesController) create(c *gin.Context) {
	var p panel.AdminRolePayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	row, err := a.roleService.Create(p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func (a *RolesController) update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p panel.AdminRolePayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	row, err := a.roleService.Update(id, p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func (a *RolesController) duplicate(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	row, err := a.roleService.Duplicate(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "obj": row})
}

func (a *RolesController) removeRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := a.roleService.Delete(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
