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
	query := config.DB.Model(&models.Roles{}).
		Preload("Permissions").
		Preload("Permissions.Permission")

	if label := c.Query("label"); label != "" {
		query = query.Where("label LIKE ?", "%"+label+"%")
	}
	if description := c.Query("description"); description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&roles)

	// 构建只包含必要字段的角色列表，并包含权限信息
	var roleDTOs []models.RoleDTO
	for _, role := range roles {
		var permissions []models.Permissions
		for _, rp := range role.Permissions {
			permissions = append(permissions, models.Permissions{
				ID:     rp.Permission.ID,
				Label:  rp.Permission.Label,
				Action: rp.Permission.Action,
				Type:   rp.Permission.Type,
			})
		}
		roleDTO := models.RoleDTO{
			ID:          role.ID,
			Label:       role.Label,
			Description: role.Description,
			Permissions: permissions,
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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误", "error": err.Error()})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 创建新角色
		role := models.Roles{Label: data.Label, Description: data.Description}
		if err := tx.Create(&role).Error; err != nil {
			return err
		}

		// 创建角色与权限的关联
		for _, permissionID := range data.Permissions {
			rolePermission := models.Rolepermission{
				RoleID:       role.ID,
				PermissionID: permissionID,
			}
			if err := tx.Create(&rolePermission).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "添加失败", "error": err.Error()})
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

// UpdateRole 处理更新角色的请求
func UpdateRole(c *gin.Context) {
	var data struct {
		ID          int    `json:"id"`
		Label       string `json:"label"`
		Description string `json:"description"`
		Permissions []int  `json:"permissions"`
	}

	// 绑定 JSON 数据到结构体
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	// 使用事务进行更新操作
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 更新角色的基本信息
		if err := tx.Model(&models.Roles{}).Where("id = ?", data.ID).Updates(models.Roles{Label: data.Label, Description: data.Description}).Error; err != nil {
			return err
		}

		// 删除旧的角色权限关联
		if err := tx.Where("role_id = ?", data.ID).Delete(&models.Rolepermission{}).Error; err != nil {
			return err
		}

		// 添加新的角色权限关联
		for _, permissionID := range data.Permissions {
			rolePermission := models.Rolepermission{
				RoleID:       data.ID,
				PermissionID: permissionID,
			}
			if err := tx.Create(&rolePermission).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}

// GrantRole 改变角色权限
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

	// 检查角色分配是否符合要求
	if (data.Type == "student" && data.RoleID == 4) || (data.Type == "teacher" && data.RoleID == 3) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "学生不能分配教师角色，教师不能分配学生角色"})
		return
	}

	// 更新用户表中的角色 ID
	if err := config.DB.Model(&models.User{}).Where("account = ?", data.Account).Update("role_id", data.RoleID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新用户表中的角色 ID 失败"})
		return
	}

	// 根据类型更新相应表中的角色 ID
	switch data.Type {
	case "student":
		if err := config.DB.Model(&models.Students{}).Where("sid = ?", data.Account).Update("role_id", data.RoleID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新学生表中的角色 ID 失败"})
			return
		}
	case "teacher":
		if err := config.DB.Model(&models.Teachers{}).Where("tid = ?", data.Account).Update("role_id", data.RoleID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "更新教师表中的角色 ID 失败"})
			return
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "未知的类型"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "操作成功"})
}
