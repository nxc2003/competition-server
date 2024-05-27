package controllers

import (
	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// ListRecords 处理 GET 请求以列出记录
func ListRecords(c *gin.Context) {
	var records []models.Records
	var count int64
	query := config.DB.Model(&models.Records{}).Preload("Student").Preload("Teacher").Preload("Race")

	if score := c.Query("score"); score != "" {
		query = query.Where("score LIKE ?", "%"+score+"%")
	}
	if title := c.Query("title"); title != "" {
		query = query.Where("race.title LIKE ?", "%"+title+"%")
	}
	if tname := c.Query("tname"); tname != "" {
		query = query.Where("teacher.name LIKE ?", "%"+tname+"%")
	}
	if sname := c.Query("sname"); sname != "" {
		query = query.Where("student.name LIKE ?", "%"+sname+"%")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "查询成功",
		"count": count,
		"data":  records,
	})
}

// AddRecord 处理 POST 请求以添加新记录
func AddRecord(c *gin.Context) {
	var data models.Records
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if msg := validateRecord(data); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": msg})
		return
	}

	config.DB.Create(&data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功"})
}

// DeleteRecord 处理 DELETE 请求以删除记录
func DeleteRecord(c *gin.Context) {
	var data []int
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	config.DB.Delete(&models.Records{}, data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateRecord 处理 PATCH 请求以更新记录
func UpdateRecord(c *gin.Context) {
	var data models.Records
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	var record models.Records
	if err := config.DB.First(&record, data.RecordID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "记录不存在"})
		return
	}

	var student models.Students
	if err := config.DB.Model(&record).Association("Student").Find(&student); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "查找学生信息失败"})
		return
	}

	if checkPermission(c, "record:update") {
		if err := config.DB.Model(&record).Updates(data).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "修改失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
	} else if student.SID == c.GetString("account") && c.GetString("identity") == "student" {
		if err := config.DB.Model(&record).Update("score", data.Score).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "修改失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "暂无权限"})
	}
}

// validateRecord 验证记录数据
func validateRecord(data models.Records) string {
	if data.RaceID == 0 || data.SID == "" {
		return "参数有误"
	}

	var count int64
	config.DB.Model(&models.Records{}).Where("race_id = ? AND sid = ?", data.RaceID, data.SID).Count(&count)
	if count > 0 {
		return "请勿重复报名"
	}

	if err := config.DB.First(&models.Races{}, data.RaceID).Error; err != nil {
		return "比赛不存在"
	}

	if err := config.DB.First(&models.Students{}, data.SID).Error; err != nil {
		return "学生信息不存在"
	}

	if data.TID != "" {
		if err := config.DB.First(&models.Teachers{}, data.TID).Error; err != nil {
			return "教师信息不存在"
		}
	}

	return ""
}

type AuthenticatedUser struct {
	Account     string
	Identity    string
	Role        models.Roles
	Permissions []string
}

// checkPermission 检查用户是否具有所需的权限
func checkPermission(c *gin.Context, permission string) bool {
	user, exists := c.Get("authenticatedUser")
	if !exists {
		return false
	}

	authUser := user.(AuthenticatedUser)
	for _, p := range authUser.Permissions {
		if p == permission {
			return true
		}
	}

	return false
}
