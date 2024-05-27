package controllers

import (
	"competition-server/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetUploadToken 获取上传令牌
func GetUploadToken(c *gin.Context) {
	name := c.Query("name")
	token := utils.GetToken(name)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// GetFileUrl 获取文件下载链接
func GetFileUrl(c *gin.Context) {
	filename := c.Query("filename")
	url := utils.GetFileUrl(filename)
	c.JSON(http.StatusOK, gin.H{
		"url": url,
	})
}

// RefreshFileUrl 刷新文件 CDN 缓存
func RefreshFileUrl(c *gin.Context) {
	name := c.Query("name")
	err := utils.RefreshUrl(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "刷新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "刷新成功"})
}

// GetFileInfo 获取文件信息
func GetFileInfo(c *gin.Context) {
	name := c.Query("name")
	info, err := utils.GetFileInfo(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "获取文件信息失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"info": info,
	})
}

// DeleteFile 删除文件
func DeleteFile(c *gin.Context) {
	var names []string
	if err := c.BindJSON(&names); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	err := utils.DeleteFile(names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}
