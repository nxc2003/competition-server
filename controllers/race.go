package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
)

// ListRaces handles GET requests to list races
func ListRaces(c *gin.Context) {
	var races []models.Races
	var count int64
	query := config.DB.Model(&models.Races{})

	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if location := c.Query("location"); location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if sponsor := c.Query("sponsor"); sponsor != "" {
		query = query.Where("sponsor LIKE ?", "%"+sponsor+"%")
	}
	if date := c.Query("date"); date != "" {
		dates := strings.Split(date, "~")
		if len(dates) == 2 {
			query = query.Where("date BETWEEN ? AND ?", dates[0], dates[1])
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&races)

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "查询成功",
		"count": count,
		"data":  races,
	})
}

// AddRace handles POST requests to add a new race
func AddRace(c *gin.Context) {
	var data models.Races
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	config.DB.Create(&data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功"})
}

// DeleteRace handles DELETE requests to delete races
func DeleteRace(c *gin.Context) {
	var data []int
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	config.DB.Delete(&models.Races{}, data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// UpdateRace handles PUT requests to update a race
func UpdateRace(c *gin.Context) {
	var data models.Races
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	if data.RaceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误"})
		return
	}

	config.DB.Model(&models.Races{}).Where("race_id = ?", data.RaceID).Updates(data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}
