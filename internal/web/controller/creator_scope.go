package controller

import (
	"errors"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

var errClientNotFound = errors.New("client not found")

// creatorScope returns the admin id when the caller is creator-tier, else 0.
// 0 = unscoped (owner/admin/editor/viewer keep the full panel view).
// u.Role already carries the base tier (custom roles resolve BaseTier into
// Role at creation), so a direct comparison covers DB roles too.
func creatorScope(c *gin.Context) int {
	u := session.GetLoginUser(c)
	if u == nil || u.Id == 0 {
		return 0
	}
	if u.Role == model.RoleCreator {
		return u.Id
	}
	return 0
}

// denyForeignClient aborts with a generic not-found when a creator touches a
// client it did not create. Returns true when the request was denied.
func denyForeignClient(c *gin.Context, createdBy int) bool {
	if scope := creatorScope(c); scope != 0 && createdBy != scope {
		jsonMsg(c, I18nWeb(c, "pages.inbounds.toasts.obtain"), errClientNotFound)
		return true
	}
	return false
}
