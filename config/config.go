package config

import (
	"competition-server/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"time"
)

var CookieKey = "your-cookie-key"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

var DB *gorm.DB

// InitDB 数据库初始化
func InitDB() {
	dsn := "root:123456@tcp(localhost:3306)/COMPETITION?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败")
	}

	// 测试数据库连接
	sqlDB, err := DB.DB()
	if err != nil {
		panic("获取数据库实例失败")
	}
	if err = sqlDB.Ping(); err != nil {
		panic("数据库连接测试失败")
	}

	log.Println("数据库连接成功")
	// 同步用户信息
	if err := SyncUsers(DB); err != nil {
		log.Fatalf("同步用户信息失败: %v", err)
	}
	log.Println("用户信息同步成功")
}

func SyncUsers(db *gorm.DB) error {
	// 获取所有学生信息
	var students []models.Students
	if err := db.Find(&students).Error; err != nil {
		return err
	}

	// 获取所有教师信息
	var teachers []models.Teachers
	if err := db.Find(&teachers).Error; err != nil {
		return err
	}

	// 创建用户切片
	var users []models.User

	// 遍历学生信息
	for _, student := range students {
		roleID := 3 // 默认角色ID为3
		if student.SID == "admin" {
			roleID = 1
		}
		users = append(users, models.User{
			Account:   student.SID,
			Password:  student.Password,
			Identity:  "student",
			RoleID:    roleID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// 遍历教师信息
	for _, teacher := range teachers {
		if teacher.TID == "" {
			log.Printf("教师账号为空: %v", teacher)
			continue
		}
		users = append(users, models.User{
			Account:   teacher.TID,
			Password:  teacher.Password,
			Identity:  "teacher",
			RoleID:    4, // 教师角色ID默认设置为4
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	// 批量插入用户信息
	if err := db.Create(&users).Error; err != nil {
		return err
	}

	return nil
}
