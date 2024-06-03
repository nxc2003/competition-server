package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// TokenKey 是用于签名和验证 JWT 令牌的密钥
var TokenKey = "token-ncu_university-competition_server"

// LoginCheckMiddleware 是一个中间件函数，用于检查用户的登录状态和权限
func LoginCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Cookie 中获取令牌
		token, err := c.Cookie("uid")
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "拒绝访问"})
			c.Abort()
			return
		}

		// 验证令牌并解析负载
		payload, err := verifyToken(token)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "拒绝访问"})
			c.Abort()
			return
		}

		// 检查令牌是否过期
		exp, ok := payload["exp"].(float64)
		if !ok || time.Now().After(time.Unix(int64(exp), 0)) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "请重新登录"})
			c.Abort()
			return
		}

		// 从数据库中获取用户信息
		var user models.User
		if err := config.DB.Where("account = ?", payload["account"]).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "无法找到用户"})
			c.Abort()
			return
		}

		//// 测试代码：返回查询出的用户信息 -- 获取成功
		//c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "用户查询成功", "data": user})
		//c.Abort()
		//return

		// 获取用户的角色信息
		var role models.Roles
		if err := config.DB.Where("id = ?", user.RoleID).First(&role).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "无法找到角色"})
			c.Abort()
			return
		}

		//// 测试代码：返回查询出的角色信息--获取成功
		//c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "角色查询成功", "data": role})
		//c.Abort()
		//return

		// 获取用户的权限信息
		var rolePermissions []models.Rolepermission
		if err := config.DB.Where("role_id = ?", role.ID).Find(&rolePermissions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "无法找到权限"})
			c.Abort()
			return
		}
		//// 测试代码：返回查询出的角色权限信息--获取成功
		//c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "角色权限查询成功", "data": rolePermissions})
		//c.Abort()
		//return
		var permissions []models.Permissions
		for _, rp := range rolePermissions {
			var permission models.Permissions
			if err := config.DB.Where("id = ?", rp.PermissionID).First(&permission).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "无法找到权限"})
				c.Abort()
				return
			}
			permissions = append(permissions, permission)
		}

		// 将权限信息转换为字符串数组
		var userPermissions []string
		for _, p := range permissions {
			userPermissions = append(userPermissions, p.Type+":"+p.Action)
		}

		//// 测试代码：返回用户信息和权限
		//c.JSON(http.StatusOK, gin.H{
		//	"code":        200,
		//	"msg":         "用户信息",
		//	"user":        user,
		//	"role":        role,
		//	"permissions": userPermissions,
		//})
		//return

		// 将用户信息和权限添加到 Gin 的上下文中
		c.Set("authenticatedUser", models.AuthenticatedUser{
			Account:     user.Account,
			Identity:    user.Identity,
			Role:        role,
			Permissions: userPermissions,
		})
		c.Next()
	}
}

// verifyToken 验证和解析 JWT 令牌
func verifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非法的签名方法: %v", token.Header["alg"])
		}
		return []byte(TokenKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("无效的令牌")
}
