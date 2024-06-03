package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Roles struct {
	ID          int              `gorm:"primaryKey" json:"id"`
	Label       string           `gorm:"unique" json:"label"`
	Description string           `json:"description"`
	Students    []Students       `gorm:"foreignKey:RoleID;references:ID"`
	Teachers    []Teachers       `gorm:"foreignKey:RoleID;references:ID"`
	Permissions []Rolepermission `gorm:"foreignKey:RoleID;references:ID"`
}

type Permissions struct {
	ID     int    `gorm:"primaryKey" json:"id"`
	Label  string `gorm:"size:255;unique" json:"label"`
	Action string `gorm:"type:enum('add','delete','update','query','import','export')" json:"action"`
	Type   string `gorm:"type:enum('user','role','race','record','permission')" json:"type"`
}

// Rolepermission 定义角色与权限对应关系的结构体
type Rolepermission struct {
	PermissionID int         `gorm:"primaryKey" json:"permission_id"`
	RoleID       int         `gorm:"primaryKey" json:"role_id"`
	Permission   Permissions `gorm:"foreignKey:PermissionID;references:ID" json:"permission"`
	Role         Roles       `gorm:"foreignKey:RoleID;references:ID" json:"role"`
}

type User struct {
	Account   string         `gorm:"column:account;primaryKey;type:varchar(255);not null" json:"account"`
	Password  string         `gorm:"type:varchar(255);not null" json:"password"`
	Identity  string         `gorm:"type:varchar(255);not null;check:identity IN ('student', 'teacher')" json:"identity"`
	RoleID    int            `gorm:"index" json:"role_id"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedat"`
}

// Races 改动新添加开始/截止日期
type Races struct {
	RaceID      int       `gorm:"column:race_id;primaryKey" json:"race_id"`
	Title       string    `gorm:"size:255" json:"title"`
	Sponsor     string    `gorm:"size:255" json:"sponsor"`
	Type        string    `gorm:"size:255" json:"type"`
	Level       int       `json:"level"`
	Location    string    `gorm:"size:255" json:"location"`
	Startdate   time.Time `json:"startdate" json:"startdate"`
	Enddate     time.Time `json:"enddate" json:"enddate"`
	Description string    `gorm:"size:255" json:"description"`
	Records     []Records `gorm:"foreignKey:RaceID;references:RaceID" json:"records"`
	CreateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

//type Races struct {
//	RaceID      int       `gorm:"column:race_id;primaryKey" json:"race_id"`
//	Title       string    `gorm:"size:255" json:"title"`
//	Sponsor     string    `gorm:"size:255" json:"sponsor"`
//	Type        string    `gorm:"size:255" json:"type"`
//	Level       int       `json:"level"`
//	Location    string    `gorm:"size:255" json:"location"`
//	Date        time.Time `json:"date"`
//	Description string    `gorm:"size:255" json:"description"`
//	Records     []Records `gorm:"foreignKey:RaceID;references:RaceID" json:"records"`
//	CreateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
//	UpdateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
//}

type Students struct {
	SID        string    `gorm:"column:sid;primaryKey" json:"sid"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	Password   string    `gorm:"size:255;not null" json:"password"`
	Sex        *int      `gorm:"not null" json:"sex"` // 因为0代表女生故设为指针类型
	Grade      int       `gorm:"not null" json:"grade"`
	Class      string    `gorm:"size:255;not null" json:"class"`
	RoleID     int       `gorm:"index" json:"role_id"`
	Records    []Records `gorm:"foreignKey:SID;references:SID" json:"records"`
	CreateTime time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
}

type Teachers struct {
	TID         string    `gorm:"column:tid;primaryKey" json:"tid"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Password    string    `gorm:"size:255;not null" json:"password"`
	Rank        int       `gorm:"not null;default:0" json:"rank"`
	Description string    `gorm:"size:255" json:"description"`
	CreateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"create_time"`
	UpdateTime  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"update_time"`
	RoleID      int       `gorm:"index" json:"role_id"`
}

// Records 数据库表的结构体定义
type Records struct {
	RecordID    int       `gorm:"column:record_id" json:"record_id"`
	Status      int       `gorm:"column:status;default:0" json:"status"`
	Score       string    `gorm:"column:score;type:varchar(255)" json:"score"`
	Description string    `gorm:"column:description;type:varchar(255)" json:"description"`
	SID         string    `gorm:"column:sid;type:varchar(255)" json:"sid"`
	TID         string    `gorm:"column:tid;type:varchar(255);default:null" json:"tid"`
	RaceID      int       `gorm:"column:race_id;index" json:"race_id"`
	CreateTime  time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time" json:"update_time"`
	Student     Students  `gorm:"foreignKey:SID;references:SID" json:"student"`
	Teacher     Teachers  `gorm:"foreignKey:TID;references:TID" json:"teacher"`
	Race        Races     `gorm:"foreignKey:RaceID;references:RaceID" json:"race"`
}

func (s *Students) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	s.Password = string(hashedPassword)
	return nil
}

func (t *Teachers) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	t.Password = string(hashedPassword)
	return nil
}
