package controllers

import (
	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

// 获取当前用户类型
func getUserType(userType string) interface{} {
	switch userType {
	case "student":
		return &models.Students{}
	case "teacher":
		return &models.Teachers{}
	default:
		return nil
	}
}

// InitUser 初始化信息
func InitUser(c *gin.Context) {
	user, exists := c.Get("authenticatedUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "用户未认证"})
		return
	}
	authUser := user.(models.AuthenticatedUser)
	account := authUser.Account
	identity := authUser.Identity
	var userDetails map[string]interface{}
	if identity == "student" {
		var student models.Students
		if err := config.DB.Where("sid = ?", account).First(&student).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "学生信息未找到"})
			return
		}
		// 5-31 23:00 修改返回信息account为sid/tid
		userDetails = map[string]interface{}{
			"sid":     student.SID,
			"name":    student.Name,
			"sex":     student.Sex,
			"grade":   student.Grade,
			"class":   student.Class,
			"role_id": student.RoleID,
		}
	} else if identity == "teacher" {
		var teacher models.Teachers
		if err := config.DB.Where("tid = ?", account).First(&teacher).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "教师信息未找到"})
			return
		}
		userDetails = map[string]interface{}{
			"tid":         teacher.TID,
			"name":        teacher.Name,
			"rank":        teacher.Rank,
			"description": teacher.Description,
			"role_id":     teacher.RoleID,
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的身份类型"})
		return
	}

	// 将 authUser 的信息合并到 userDetails 中
	userDetails["identity"] = authUser.Identity
	userDetails["role"] = authUser.Role
	userDetails["permissions"] = authUser.Permissions

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "获取成功", "data": userDetails})
}

// ListUsers 用于学生/教师用户查询
func ListUsers(c *gin.Context) {
	type QueryParams struct {
		Type   string `form:"type"`
		Offset int    `form:"offset"`
		Limit  int    `form:"limit"`
		Name   string `form:"name"`
		Class  string `form:"class"`
		Rank   *int   `form:"rank"` // 使用指针类型来区分零值和未提供的值
		SID    string `form:"sid"`
		Sex    *int   `form:"sex"` // 使用指针类型来区分零值和未提供的值
		Grade  int    `form:"grade"`
		TID    string `form:"tid"`
	}

	var queryParams QueryParams
	if err := c.ShouldBindQuery(&queryParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数有误",
		})
		return
	}

	if queryParams.Type == "student" {
		var students []models.Students
		var count int64
		query := config.DB.Model(&models.Students{})

		if queryParams.Name != "" {
			query = query.Where("name LIKE ?", "%"+queryParams.Name+"%")
		}
		if queryParams.Class != "" {
			query = query.Where("class LIKE ?", "%"+queryParams.Class+"%")
		}
		if queryParams.SID != "" {
			query = query.Where("sid = ?", queryParams.SID)
		}
		if queryParams.Sex != nil {
			query = query.Where("sex = ?", *queryParams.Sex)
		}
		if queryParams.Grade != 0 {
			query = query.Where("grade = ?", queryParams.Grade)
		}

		limit := queryParams.Limit
		offset := queryParams.Offset
		if limit <= 0 {
			limit = 10
		}
		if offset <= 0 {
			offset = 1
		}
		query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Order("create_time DESC").Find(&students)

		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"msg":   "查询成功",
			"count": count,
			"data":  students,
		})
	} else if queryParams.Type == "teacher" {
		var teachers []models.Teachers
		var count int64
		query := config.DB.Model(&models.Teachers{})

		if queryParams.Name != "" {
			query = query.Where("name LIKE ?", "%"+queryParams.Name+"%")
		}
		if queryParams.Rank != nil {
			query = query.Where("`rank` = ?", *queryParams.Rank) // 使用反引号转义 rank 字段
		}
		if queryParams.TID != "" {
			query = query.Where("tid = ?", queryParams.TID)
		}

		limit := queryParams.Limit
		offset := queryParams.Offset
		if limit <= 0 {
			limit = 10
		}
		if offset <= 0 {
			offset = 1
		}
		query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Order("create_time DESC").Find(&teachers)

		c.JSON(http.StatusOK, gin.H{
			"code":  200,
			"msg":   "查询成功",
			"count": count,
			"data":  teachers,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的用户类型",
		})
	}
}

// UpdateUser 更行用户信息
func UpdateUser(c *gin.Context) {
	var requestData struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if requestData.Type == "student" {
		var studentData models.Students
		if err := mapstructure.Decode(requestData.Data, &studentData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "学生数据解析失败"})
			return
		}

		if err := config.DB.Model(&models.Students{}).Where("sid = ?", studentData.SID).Updates(studentData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "学生更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "学生信息修改成功"})
	} else if requestData.Type == "teacher" {
		var teacherData models.Teachers
		if err := mapstructure.Decode(requestData.Data, &teacherData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "教师数据解析失败"})
			return
		}

		if err := config.DB.Model(&models.Teachers{}).Where("tid = ?", teacherData.TID).Updates(teacherData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "教师更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "教师信息修改成功"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "未知的用户类型"})
	}
}

// UpdatePassword 处理更新密码请求
func UpdatePassword(c *gin.Context) {
	var req struct {
		Identity string `json:"identity"`
		OldVal   string `json:"oldVal"`
		NewVal   string `json:"newVal"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	authenticatedUser, exists := c.Get("authenticatedUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未授权"})
		return
	}
	user := authenticatedUser.(models.AuthenticatedUser)

	// 校验旧密码并更新密码
	var userAccount models.User
	if err := config.DB.Where("account = ?", user.Account).First(&userAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "用户不存在"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userAccount.Password), []byte(req.OldVal)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "旧密码错误"})
		return
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewVal), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	userAccount.Password = string(hashedNewPassword)

	if err := config.DB.Save(&userAccount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新用户密码失败"})
		return
	}

	// 更新学生或教师密码
	if req.Identity == "student" {
		var student models.Students
		if err := config.DB.Where("sid = ?", user.Account).First(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "学生不存在"})
			return
		}

		student.Password = string(hashedNewPassword)

		if err := config.DB.Save(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新学生密码失败"})
			return
		}
	} else if req.Identity == "teacher" {
		var teacher models.Teachers
		if err := config.DB.Where("tid = ?", user.Account).First(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "教师不存在"})
			return
		}

		teacher.Password = string(hashedNewPassword)

		if err := config.DB.Save(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新教师密码失败"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "身份不合法"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "密码更新成功"})
}

// ResetPassword 处理密码重置请求
func ResetPassword(c *gin.Context) {
	var req struct {
		Type    string `json:"type"`
		Account string `json:"account"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	if req.Type == "student" {
		var student models.Students
		if err := config.DB.Where("sid = ?", req.Account).First(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "学生账户不存在"})
			return
		}
		student.Password = string(hashedPassword)
		if err := config.DB.Save(&student).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新学生密码失败"})
			return
		}
	} else if req.Type == "teacher" {
		var teacher models.Teachers
		if err := config.DB.Where("tid = ?", req.Account).First(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "教师账户不存在"})
			return
		}
		teacher.Password = string(hashedPassword)
		if err := config.DB.Save(&teacher).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新教师密码失败"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "身份不合法"})
		return
	}

	// 更新 User 表中的密码
	var user models.User
	if err := config.DB.Where("account = ?", req.Account).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "用户账户不存在"})
		return
	}
	user.Password = string(hashedPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新用户密码失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "密码重置成功"})
}

// AddUsers 添加用户
func AddUsers(c *gin.Context) {
	var requestData struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	// 处理初始密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "密码加密失败"})
		return
	}

	if requestData.Type == "student" {
		var studentData models.Students
		if err := mapstructure.Decode(requestData.Data, &studentData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "学生数据解析失败"})
			return
		}

		studentData.Password = string(hashedPassword)
		studentData.RoleID = 3 // 设置学生的默认 role_id 为 3

		// 添加学生到数据库
		if err := config.DB.Create(&studentData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "学生创建失败"})
			return
		}

		// 添加用户到User表
		user := models.User{
			Account:  studentData.SID,
			Password: studentData.Password,
			Identity: "student",
			RoleID:   3, // 对应的用户表角色ID
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "用户创建失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "学生创建成功"})
	} else if requestData.Type == "teacher" {
		var teacherData models.Teachers
		if err := mapstructure.Decode(requestData.Data, &teacherData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "教师数据解析失败"})
			return
		}

		teacherData.Password = string(hashedPassword)
		teacherData.RoleID = 4 // 设置教师的默认 role_id 为 4

		// 添加教师到数据库
		if err := config.DB.Create(&teacherData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "教师创建失败"})
			return
		}

		// 添加用户到User表
		user := models.User{
			Account:  teacherData.TID,
			Password: teacherData.Password,
			Identity: "teacher",
			RoleID:   4, // 对应的用户表角色ID
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "用户创建失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "教师创建成功"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "未知的用户类型"})
	}
}

// AddImport 批量导入学生/教师数据
func AddImport(c *gin.Context) {
	var requestData struct {
		Type string                   `json:"type"`
		Data []map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误", "error": err.Error()})
		return
	}

	initialPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	switch requestData.Type {
	case "student":
		for _, studentData := range requestData.Data {
			sid, _ := studentData["sid"].(string)
			name, _ := studentData["name"].(string)
			sex, _ := studentData["sex"].(int)
			grade, _ := studentData["grade"].(int)
			class, _ := studentData["class"].(string)

			newStudent := models.Students{
				SID:        sid,
				Name:       name,
				Password:   string(initialPassword),
				Sex:        &sex,
				Grade:      grade,
				Class:      class,
				RoleID:     3,
				CreateTime: time.Now(),
				UpdateTime: time.Now(),
			}
			config.DB.Create(&newStudent)

			newUser := models.User{
				Account:  sid,
				Password: string(initialPassword),
				Identity: "student",
				RoleID:   3,
			}
			config.DB.Create(&newUser)
		}
	case "teacher":
		for _, teacherData := range requestData.Data {
			tid, _ := teacherData["tid"].(string)
			name, _ := teacherData["name"].(string)
			rank, _ := teacherData["rank"].(int)
			description, _ := teacherData["description"].(string)

			newTeacher := models.Teachers{
				TID:         tid,
				Name:        name,
				Password:    string(initialPassword),
				Rank:        rank,
				Description: description,
				RoleID:      4,
				CreateTime:  time.Now(),
				UpdateTime:  time.Now(),
			}
			config.DB.Create(&newTeacher)

			newUser := models.User{
				Account:  tid,
				Password: string(initialPassword),
				Identity: "teacher",
				RoleID:   4,
			}
			config.DB.Create(&newUser)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "未知的类型"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "导入成功"})
}

// DeleteUsers 批量删除学生/教师并且清除账户
func DeleteUsers(c *gin.Context) {
	var requestData struct {
		Type string `json:"type"`
		Data struct {
			IDs []string `json:"ids"`
		} `json:"data"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误", "error": err.Error()})
		return
	}

	switch requestData.Type {
	case "student":
		for _, sid := range requestData.Data.IDs {
			// 删除 users 表中的记录
			config.DB.Where("account = ?", sid).Delete(&models.User{})
			// 删除 students 表中的记录
			config.DB.Where("sid = ?", sid).Delete(&models.Students{})
		}
	case "teacher":
		for _, tid := range requestData.Data.IDs {
			// 删除 users 表中的记录
			config.DB.Where("account = ?", tid).Delete(&models.User{})
			// 删除 teachers 表中的记录
			config.DB.Where("tid = ?", tid).Delete(&models.Teachers{})
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "未知的类型"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}
