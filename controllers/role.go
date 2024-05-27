package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListRoles handles GET requests to list roles
func ListRoles(c *gin.Context) {
	var roles []models.Roles
	var count int64
	query := config.DB.Model(&models.Roles{}).Preload("Permissions")

	if label := c.Query("label"); label != "" {
		query = query.Where("label LIKE ?", "%"+label+"%")
	}
	if description := c.Query("description"); description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&roles)

	// 构建只包含三个字段的角色列表
	var roleDTOs []models.RoleDTO
	for _, role := range roles {
		roleDTO := models.RoleDTO{
			ID:          role.ID,
			Label:       role.Label,
			Description: role.Description,
		}
		roleDTOs = append(roleDTOs, roleDTO)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "查询成功",
		"count": count,
		"data":  roleDTOs,
	})
}

// AddRole handles POST requests to add a new role
func AddRole(c *gin.Context) {
	var data struct {
		Label       string `json:"label"`
		Description string `json:"description"`
		Permissions []int  `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		role := models.Roles{Label: data.Label, Description: data.Description}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

		var permissions []models.Permissions
		if err := tx.Where("id IN ?", data.Permissions).Find(&permissions).Error; err != nil {
			return err
		}

		return tx.Model(&role).Association("Permissions").Append(&permissions)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功"})
}

// DeleteRole handles DELETE requests to delete roles
func DeleteRole(c *gin.Context) {
	var data []int
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		for _, id := range data {
			var role models.Roles
			if err := tx.Preload("Students").Preload("Teachers").First(&role, id).Error; err != nil {
				return err
			}

			if len(role.Students) == 0 && len(role.Teachers) == 0 {
				if err := tx.Delete(&role).Error; err != nil {
					return err
				}
			} else {
				return fmt.Errorf("角色%d包含引用，不能删除", id)
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateRole handles POST requests to update a role
func UpdateRole(c *gin.Context) {
	var data struct {
		ID          int    `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
		Permissions []int  `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Roles{}).Where("id = ?", data.ID).Updates(models.Roles{Label: data.Label, Description: data.Description}).Error; err != nil {
			return err
		}

		var role models.Roles
		if err := tx.Preload("Permissions").First(&role, data.ID).Error; err != nil {
			return err
		}

		var permissions []models.Permissions
		if err := tx.Where("id IN ?", data.Permissions).Find(&permissions).Error; err != nil {
			return err
		}

		return tx.Model(&role).Association("Permissions").Replace(&permissions)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}

// GrantRole handles POST requests to grant a role to a user
func GrantRole(c *gin.Context) {
	var data struct {
		Type    string `json:"type"`
		Account string `json:"account"`
		RoleID  int    `json:"role_id"`
	}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	//	var user models.User
	userType := getUserType(data.Type)
	if err := config.DB.Model(userType).Where("account = ?", data.Account).Update("role_id", data.RoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "操作失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
}

func getUserType(userType string) interface{} {
	switch userType {
	case "student":
		return &models.Students{}
	case "teacher":
		return &models.Teachers{}
	default:
		return &models.User{}
	}
}
