package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUser 处理函数用于获取用户信息
func GetUser(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未经授权"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "成功", "data": user})
}

//// SetUser 中间件用于设置模拟用户信息
//func SetUser() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		// 模拟系统初始账号
//		user := map[string]interface{}{
//			"permissions": []string{"read", "write"},
//			"role":        "student",
//			"account":     "admin",
//			"password":    "123", // 简化的密码，实际情况应加密处理
//			"identity":    "student",
//		}
//		c.Set("user", user)
//		c.Next()
//	}
//}
