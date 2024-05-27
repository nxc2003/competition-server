package controllers

import (
	"competition-server/config"
	"competition-server/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mojocn/base64Captcha"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

var TokenKey = "token-ncu_university-competition_server"
var store = base64Captcha.DefaultMemStore

// Login handles user login and returns a JWT token
func Login(c *gin.Context) {
	var req struct {
		Account  string `json:"account"`
		Password string `json:"password"`
		Identity string `json:"identity"`
		Code     string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	sysCode, err := c.Cookie("captchaAnswer")
	//c.JSON(http.StatusBadRequest, gin.H{"code": sysCode, "msg": "验证码测试"})
	//return
	if err != nil || req.Code == "" || sysCode == "" || req.Code != sysCode {
		c.JSON(http.StatusBadRequest, gin.H{"code": 3, "msg": "验证码有误"})
		return
	}

	var user models.User
	if err := config.DB.Where("account = ? AND identity = ?", req.Account, req.Identity).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "用户不存在"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 2, "msg": "密码错误"})
		return
	}

	exp := time.Now().Add(7 * 24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"account":  req.Account,
		"identity": req.Identity,
		"exp":      exp.Unix(),
	})

	tokenString, err := token.SignedString([]byte(TokenKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成令牌失败"})
		return
	}

	c.SetCookie("uid", tokenString, int(exp.Unix()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "登陆成功"})
}

// GenerateCaptcha 生成验证码
func GenerateCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, store)
	_, b64s, answer, err := captcha.Generate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "生成验证码失败"})
		return
	}
	c.SetCookie("captchaAnswer", answer, 5*60, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "获取成功",
		"data": map[string]string{
			"answer":  answer,
			"picPath": b64s,
		},
	})
}
