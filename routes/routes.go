package routes

import (
	"competition-server/controllers"
	"competition-server/middlewares"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) *gin.Engine {

	// 身份验证路由
	auth := r.Group("/auth")
	{
		//获取验证码
		auth.GET("/code", controllers.GenerateCaptcha)
		//登录
		auth.POST("/login", controllers.Login)
	}
	// 应用登录检查和权限检查中间件
	r.Use(middlewares.LoginCheckMiddleware())
	r.Use(middlewares.AuthCheckMiddleware())
	//获取用户数据--初始化+权限
	r.GET("/get_user", controllers.InitUser)
	// 权限相关路由
	permission := r.Group("/permission")
	{
		permission.GET("/list", controllers.ListPermissions)
		permission.POST("/add", controllers.AddPermission)
		permission.DELETE("/delete", controllers.DeletePermission)
		permission.PUT("/update", controllers.UpdatePermission)
	}

	// 用户相关路由
	users := r.Group("/user")
	{
		// 返回学生/老师信息
		users.GET("/list", controllers.ListUsers)
		users.PUT("/update", controllers.UpdateUser)
		users.PATCH("/password", controllers.UpdatePassword)
		users.PUT("/reset", controllers.ResetPassword)
		users.POST("/add", controllers.AddUsers)
		users.POST("/import", controllers.AddImport)
		users.DELETE("/delete", controllers.DeleteUsers)
	}

	// 角色相关路由 -- 即超级管理员 管理员 学生等权限的管理
	role := r.Group("/role")
	{
		role.GET("/list", controllers.ListRoles)
		role.POST("/add", controllers.AddRole)
		role.POST("/update", controllers.UpdateRole)
		role.DELETE("/delete", controllers.DeleteRole)
		role.POST("/grant", controllers.GrantRole)
	}
	// 比赛相关路由
	race := r.Group("/race")
	{
		race.GET("/list", controllers.ListRaces)
		race.POST("/add", controllers.AddRace)
		race.DELETE("/delete", controllers.DeleteRace)
		race.PUT("/update", controllers.UpdateRace)
	}

	// 记录相关路由
	record := r.Group("/record")
	{
		record.POST("/add", controllers.AddRecord)
		record.DELETE("/delete", controllers.DeleteRecord)
		record.PATCH("/update", controllers.UpdateRecord)
		record.GET("/list", controllers.ListRecords)
	}

	// 文件上传下载管理
	file := r.Group("/file")
	{
		file.GET("/get_upload_token", controllers.GetUploadToken)
		file.GET("/get_file_url", controllers.GetFileUrl)
		file.POST("/refresh_file_url", controllers.RefreshFileUrl)
		file.GET("/get_file_info", controllers.GetFileInfo)
		file.POST("/delete_file", controllers.DeleteFile)
	}

	//// 使用 SetUser 中间件
	//r.Use(middlewares.SetUser())

	return r
}
