package middleware

import (
	"net/http"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

// RequireRole ensures login user has one of the given roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles { allowed[r] = true }
	return func(c *gin.Context) {
		u := session.GetLoginUser(c)
		if u == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "unauthorized"})
			return
		}
		if !u.Enabled {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "account disabled"})
			return
		}
		if !allowed[u.Role] && !allowed["*"] {
			// owner always passes
			if u.Role != model.RoleOwner {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "forbidden: insufficient role"})
				return
			}
		}
		c.Next()
	}
}

// RequirePermission checks HasPermission for the logged-in user.
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := session.GetLoginUser(c)
		if u == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "msg": "unauthorized"})
			return
		}
		if !model.HasPermission(u.Role, perm) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "forbidden: missing permission " + perm})
			return
		}
		c.Next()
	}
}
