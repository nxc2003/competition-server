package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
)

// ListRaces 关键字查询比赛
func ListRaces(c *gin.Context) {
	var races []models.Races
	var count int64
	query := config.DB.Model(&models.Races{})

	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if sponsor := c.Query("sponsor"); sponsor != "" {
		query = query.Where("sponsor LIKE ?", "%"+sponsor+"%")
	}
	if location := c.Query("location"); location != "" {
		query = query.Where("location LIKE ?", "%"+location+"%")
	}
	if raceType := c.Query("type"); raceType != "" {
		query = query.Where("type = ?", raceType)
	}
	if level := c.Query("level"); level != "" {
		levelInt, err := strconv.Atoi(level)
		if err == nil {
			query = query.Where("level = ?", levelInt)
		}
	}
	//后续优化data 根据截止日期进行查询
	if date := c.Query("date"); date != "" {
		dates := strings.Split(date, "~")
		if len(dates) == 2 {
			query = query.Where("enddate BETWEEN ? AND ?", dates[0], dates[1])
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "1"))
	query.Count(&count).Limit(limit).Offset(limit * (offset - 1)).Find(&races).Order("create_time DESC")

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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误", "error": err.Error()})
		return
	}

	// 设置创建和更新时间
	now := time.Now()
	data.CreateTime = now
	data.UpdateTime = now

	// 插入新记录
	if err := config.DB.Create(&data).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "数据库错误", "error": err.Error()})
		return
	}

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
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误---数据读取失败"})
		return
	}

	if data.RaceID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数有误---RaceID为0"})
		return
	}

	config.DB.Model(&models.Races{}).Where("race_id = ?", data.RaceID).Updates(data)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "修改成功"})
}
