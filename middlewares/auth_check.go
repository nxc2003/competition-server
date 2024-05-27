package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type authenticatedUser struct {
	Account     string
	Identity    string
	Role        string
	Permissions []string
}

// CheckPermission 检查用户是否具有所需的权限
func CheckPermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("authenticatedUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "暂无权限---user为空"})
			c.Abort()
			return
		}

		authUser := user.(AuthenticatedUser)
		for _, p := range authUser.Permissions {
			if p == permission {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "暂无权限"})
		c.Abort()
	}
}

var Strategy = map[string]gin.HandlerFunc{
	"/user/add":          CheckPermission("user:add"),
	"/user/delete":       CheckPermission("user:delete"),
	"/user/reset":        CheckPermission("user:update"),
	"/user/list":         CheckPermission("user:query"),
	"/race/add":          CheckPermission("race:add"),
	"/race/delete":       CheckPermission("race:delete"),
	"/race/list":         CheckPermission("race:query"),
	"/race/update":       CheckPermission("race:update"),
	"/record/add":        CheckPermission("record:add"),
	"/record/delete":     CheckPermission("record:delete"),
	"/record/list":       CheckPermission("record:query"),
	"/permission/list":   CheckPermission("permission:query"),
	"/permission/add":    CheckPermission("permission:add"),
	"/permission/delete": CheckPermission("permission:delete"),
	"/permission/update": CheckPermission("permission:update"),
	"/role/list":         CheckPermission("role:query"),
	"/role/add":          CheckPermission("role:add"),
	"/role/delete":       CheckPermission("role:delete"),
	"/role/update":       CheckPermission("role:update"),
	"/role/grant":        CheckPermission("role:update"),
}

func AuthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		for route, checker := range Strategy {
			if strings.HasPrefix(path, route) {
				checker(c)
				return
			}
		}
		c.Next()
	}
}
