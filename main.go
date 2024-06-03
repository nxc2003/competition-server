package main

import (
	"competition-server/config"
	"competition-server/routes"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"time"
)

func main() {
	// 初始化数据库
	config.InitDB()

	// 创建Gin路由
	r := gin.Default()

	// 设置日志
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // 前端服务器地址
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 设置会话
	store := cookie.NewStore([]byte(config.CookieKey))
	r.Use(sessions.Sessions("mysession", store))

	// 请求速率限制
	r.Use(rateLimitMiddleware(5, time.Second))

	// 路由
	routes.SetupRouter(r)

	// 错误处理
	r.Use(errorHandler)

	// 启动服务器
	if err := r.Run(":3000"); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// 错误处理
func errorHandler(c *gin.Context) {
	c.Next()
	if len(c.Errors) > 0 {
		err := c.Errors[0].Err
		var status int
		var message string

		var validationErr *config.ValidationError
		if errors.As(err, &validationErr) {
			status = http.StatusBadRequest
			message = validationErr.Error()
		} else {
			status = http.StatusInternalServerError
			message = "内部服务器错误"
		}

		c.JSON(status, gin.H{"code": status, "msg": message})
	}
}

// 请求速率限制中间件
func rateLimitMiddleware(maxRequests int, duration time.Duration) gin.HandlerFunc {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	return func(c *gin.Context) {
		limit := make(chan struct{}, maxRequests)

		select {
		case limit <- struct{}{}:
			c.Next()
			<-limit
		default:
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "msg": "请求太频繁，歇会吧~"})
			c.Abort()
		}
	}
}
