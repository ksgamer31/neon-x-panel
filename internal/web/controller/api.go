package controller

import (
	"net/http"
	"strings"

	"github.com/ksgamer31/neon-x-panel/v3/internal/database/model"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/middleware"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service/panel"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/service/tgbot"
	"github.com/ksgamer31/neon-x-panel/v3/internal/web/session"

	"github.com/gin-gonic/gin"
)

// APIController handles the main API routes for the 3x-ui panel, including inbounds and server management.
type APIController struct {
	BaseController
	inboundController     *InboundController
	serverController      *ServerController
	nodeController        *NodeController
	hostController        *HostController
	settingController     *SettingController
	xraySettingController *XraySettingController
	userService           panel.UserService
	apiTokenService       panel.ApiTokenService
	Tgbot                 tgbot.Tgbot
}

// NewAPIController creates a new APIController instance and initializes its routes.
func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

func (a *APIController) checkAPIAuth(c *gin.Context) {
	// A verified client certificate (a completed mTLS handshake) authenticates
	// the caller, equivalent to a valid bearer token. api_authed must be set so
	// the CSRF middleware lets cert-authed mutations through.
	if c.Request.TLS != nil && len(c.Request.TLS.VerifiedChains) > 0 {
		if u, err := a.userService.GetFirstUser(); err == nil {
			session.SetAPIAuthUser(c, u)
		}
		c.Set("api_authed", true)
		c.Set("api_token_scope", model.ApiScopeNodeSync)
		c.Next()
		return
	}
	auth := c.GetHeader("Authorization")
	if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
		tok := after
		if row, ok := a.apiTokenService.MatchToken(tok); ok {
			if u, err := a.userService.GetFirstUser(); err == nil {
				session.SetAPIAuthUser(c, u)
			}
			c.Set("api_authed", true)
			c.Set("api_token_scope", row.Scope)
			c.Next()
			return
		}
	}
	if !session.IsLogin(c) {
		if c.GetHeader("X-Requested-With") == "XMLHttpRequest" {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.AbortWithStatus(http.StatusNotFound)
		}
		return
	}
	c.Next()
}

// monitorScopeAllow exposes only status/metrics routes without sensitive data.
// Keys are route patterns relative to /panel/api.
var monitorScopeAllow = map[string]struct{}{
	"/server/status":                              {},
	"/server/cpuHistory/:bucket":                  {},
	"/server/history/:metric/:bucket":             {},
	"/server/xrayMetricsState":                    {},
	"/server/xrayMetricsHistory/:metric/:bucket":  {},
	"/server/xrayObservatory":                     {},
	"/server/xrayObservatoryHistory/:tag/:bucket": {},
	"/server/getXrayVersion":                      {},
	"/server/getPanelUpdateInfo":                  {},
	"/nodes/history/:id/:metric/:bucket":          {},
}

// nodeSyncScopeAllow is the node-sync route/method allowlist relative to
// /panel/api; Gin patterns prevent concrete parameters broadening authority.
var nodeSyncScopeAllow = map[string]map[string]struct{}{
	"/server/status":               {http.MethodGet: {}},
	"/inbounds/list":               {http.MethodGet: {}},
	"/inbounds/add":                {http.MethodPost: {}},
	"/inbounds/del/:id":            {http.MethodPost: {}},
	"/inbounds/update/:id":         {http.MethodPost: {}},
	"/clients/add":                 {http.MethodPost: {}},
	"/clients/del/:email":          {http.MethodPost: {}},
	"/clients/:email/detach":       {http.MethodPost: {}},
	"/clients/update/:email":       {http.MethodPost: {}},
	"/server/restartXrayService":   {http.MethodPost: {}},
	"/server/getWebCertFiles":      {http.MethodGet: {}},
	"/server/descendants":          {http.MethodGet: {}},
	"/clients/resetTraffic/:email": {http.MethodPost: {}},
	"/inbounds/resetAllTraffics":   {http.MethodPost: {}},
	"/inbounds/:id/resetTraffic":   {http.MethodPost: {}},
	"/clients/onlinesByGuid":       {http.MethodPost: {}},
	"/clients/onlines":             {http.MethodPost: {}},
	"/clients/lastOnline":          {http.MethodPost: {}},
	"/inbounds/pushClientTraffics": {http.MethodPost: {}},
	"/server/clientIps":            {http.MethodGet: {}, http.MethodPost: {}},
	"/clients/clientIpsByGuid":     {http.MethodPost: {}},
	"/hosts/list":                  {http.MethodGet: {}},
}

// enforceTokenScope applies explicit allowlists to monitor and node-sync tokens.
// Admin tokens and session-login users retain their existing behavior.
func (a *APIController) enforceTokenScope(c *gin.Context) {
	scopeVal, ok := c.Get("api_token_scope")
	if !ok {
		c.Next()
		return
	}
	scope, _ := scopeVal.(string)
	if scope == model.ApiScopeAdmin {
		c.Next()
		return
	}
	deny := func() {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"msg":     "this API token is not permitted to access this endpoint",
		})
	}
	rel := relAPIPath(c.FullPath())
	switch scope {
	case model.ApiScopeMonitor:
		if _, allowed := monitorScopeAllow[rel]; allowed && (c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead) {
			c.Next()
			return
		}
	case model.ApiScopeNodeSync:
		if methods, allowed := nodeSyncScopeAllow[rel]; allowed {
			if _, allowedMethod := methods[c.Request.Method]; allowedMethod {
				c.Next()
				return
			}
		}
	default:
		deny()
		return
	}
	deny()
}


// enforceRBAC — Neon X tiered RBAC (owner = *):
// viewer  (view):  read-only — only GET + read-like POSTs (lists/options/onlines)
// creator (client-create-only): view + POST /clients/add only; no inbound mutate, no client edit/delete, no settings/nodes/hosts/xray
// editor  (create+edit): view + inbound create/edit + client create/edit; no delete, no settings/nodes/hosts/xray
// admin   (creator-full): view + all inbound+client create/edit/delete; no settings/nodes/hosts/xray
// owner   (): full panel (bypass)
func (a *APIController) enforceRBAC(c *gin.Context) {
	u := session.GetLoginUser(c)
	if u == nil {
		c.Next()
		return
	}
	if u.Role == model.RoleOwner || u.Role == "" {
		c.Next()
		return
	}
	rel := relAPIPath(c.FullPath())
	method := c.Request.Method
	isMutating := method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete || method == http.MethodPatch

// settings/nodes/hosts read endpoints needed by viewers — allow for all non-owner roles
	readSensitiveAllow := map[string]bool{
		"/setting/all": true, 
		"/setting/defaultSettings": true, 
		"/setting/factoryDefaults": true, 
		"/setting/getDefaultJsonConfig": true,
		"/nodes/list": true,
		"/hosts/list": true,
	}
	// All non-owner roles are blocked from settings/nodes/hosts/xray/admin management (except read-like endpoints)
	privilegedPrefixes := []string{"/setting/", "/nodes/", "/hosts/", "/xray/", "/users/"}
	if u.Role == model.RoleViewer || u.Role == model.RoleCreator || u.Role == model.RoleEditor || u.Role == model.RoleAdmin {
		if readSensitiveAllow[rel] {
			// skip privileged block for read-like setting/nodes/hosts endpoints
		} else {
			for _, p := range privilegedPrefixes {
				if len(rel) >= len(p) && rel[:len(p)] == p {
					// /users/me is allowed for everyone (handled separately — but enforceRBAC sees /users/me)
					if rel == "/users/me" {
						break
					}
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "forbidden: insufficient role"})
					return
				}
			}
		}
		// also block xray/server restart etc
		if rel == "/server/restartXrayService" || rel == "/server/getXrayVersion" || rel == "/backuptotgbot" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "forbidden: insufficient role"})
			return
		}
	}

	if u.Role == model.RoleViewer {
		if isMutating {
			readPOST := map[string]bool{
				"/setting/all": true,
				"/setting/defaultSettings": true,
				"/setting/factoryDefaults": true,
				"/setting/getDefaultJsonConfig": true,
				"/inbounds/list": true,
				"/inbounds/list/slim": true,
				"/inbounds/options": true,
				"/inbounds/allLinks": true,
				"/clients/onlines": true,
				"/clients/onlinesByGuid": true,
				"/clients/lastOnline": true,
				"/clients/clientIpsByGuid": true,
				"/server/clientIps": true,
				"/clients/list": true,
				"/clients/list/paged": true,
				"/clients/get/:email": true,
			}
			if !readPOST[rel] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "viewer: read-only"})
				return
			}
		}
		c.Next()
		return
	}
	if u.Role == model.RoleCreator {
		// creator = client:create only (+ view). Allow only POST /clients/add
		allowed := map[string]bool{
			"/clients/add": true,
			"/clients/bulkCreate": true,
			// read-like
			"/inbounds/list": true,
			"/inbounds/list/slim": true,
			"/inbounds/options": true,
			"/clients/list": true,
			"/clients/list/paged": true,
			"/clients/get/:email": true,
			"/clients/onlines": true,
			"/clients/onlinesByGuid": true,
			"/clients/lastOnline": true,
			"/setting/all": true,
			"/setting/defaultSettings": true,
			"/setting/factoryDefaults": true,
			"/setting/getDefaultJsonConfig": true,
			"/nodes/list": true,
			"/hosts/list": true,
		}
		if isMutating && !allowed[rel] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "creator: can only create clients"})
			return
		}
		c.Next()
		return
	}
	if u.Role == model.RoleEditor {
		// editor: create+edit inbound+client, no delete
		blocked := map[string]bool{
			"/inbounds/del/:id": true,
			"/inbounds/bulkDel": true,
			"/inbounds/:id/delAllClients": true,
			"/clients/del/:email": true,
			"/clients/bulkDel": true,
			"/clients/delDepleted": true,
			"/clients/delOrphans": true,
			"/clients/:email/detach": true,
			"/clients/bulkDetach": true,
		}
		if blocked[rel] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "msg": "editor: cannot delete"})
			return
		}
		c.Next()
		return
	}
	if u.Role == model.RoleAdmin {
		// admin = full inbound+client (creator-full), no settings/nodes already blocked
		c.Next()
		return
	}
	c.Next()
}

func relAPIPath(fullPath string) string {
	const marker = "/panel/api"
	_, after, ok := strings.Cut(fullPath, marker)
	if !ok {
		return ""
	}
	return after
}

// initRouter sets up the API routes for inbounds, server, and other endpoints.
func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)
	api.Use(a.enforceTokenScope)
	api.Use(a.enforceRBAC)
	// Decode + verify the node config envelope (zstd + X-Config-Sha256) and
	// advertise support, before CSRF/handlers read the body.
	api.Use(middleware.ConfigEnvelopeMiddleware())
	api.Use(middleware.CSRFMiddleware())

	api.GET("/openapi.json", ServeOpenAPISpec)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	clients := api.Group("/clients")
	NewClientController(clients)
	NewGroupController(clients)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Nodes API — multi-panel management
	nodes := api.Group("/nodes")
	a.nodeController = NewNodeController(nodes)

	// Hosts API — per-inbound override endpoints for subscription links
	hosts := api.Group("/hosts")
	a.hostController = NewHostController(hosts)

	// Settings + Xray config management live under the API surface too, so the
	// same API token drives them. Paths are /panel/api/setting/* and
	// /panel/api/xray/*.
	a.settingController = NewSettingController(api)
	a.xraySettingController = NewXraySettingController(api)

	// Subscription balancers — client-side balancers for the JSON sub output
	NewSubBalancerController(api)

	// Neon X: multi-admin management
	NewUsersController(api)

	// Extra routes
	api.POST("/backuptotgbot", a.BackuptoTgbot)
}

// BackuptoTgbot sends a backup of the panel data to Telegram bot admins.
func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
