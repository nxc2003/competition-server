package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListPermissions 获取权限列表
func ListPermissions(c *gin.Context) {
	var permissions []models.Permissions
	var count int64
	query := config.DB.Model(&models.Permissions{})

	if label := c.Query("label"); label != "" {
		query = query.Where("label LIKE ?", "%"+label+"%")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&permissions)

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "查询成功",
		"count": count,
		"data":  permissions,
	})
}

// AddPermission handles POST requests to add a new permission
func AddPermission(c *gin.Context) {
	var data models.Permissions
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if exists := config.DB.Where("action = ? AND type = ?", data.Action, data.Type).First(&models.Permissions{}).RowsAffected; exists > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "权限已存在"})
		return
	}

	config.DB.Create(&data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功"})
}

// DeletePermission handles DELETE requests to delete permissions
func DeletePermission(c *gin.Context) {
	var data []int
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range data {
			var permission models.Permissions
			if err := tx.First(&permission, id).Error; err != nil {
				return err
			}
			// 使用原始SQL查询统计关联角色数
			var count int64
			if err := tx.Raw("SELECT COUNT(*) FROM rolepermissions WHERE permission_id = ?", permission.ID).Scan(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				if err := tx.Delete(&permission).Error; err != nil {
					return err
				}
			} else {
				return errors.New("权限被角色引用，不能删除")
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdatePermission handles POST requests to update a permission
func UpdatePermission(c *gin.Context) {
	var data models.Permissions
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if data.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if exists := config.DB.Where("action = ? AND type = ?", data.Action, data.Type).First(&models.Permissions{}).RowsAffected; exists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "权限不存在"})
		return
	}

	config.DB.Model(&models.Permissions{}).Where("id = ?", data.ID).Updates(data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}
