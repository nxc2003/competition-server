package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Roles struct {
	ID          int    `gorm:"primaryKey"`
	Label       string `gorm:"unique"`
	Description string
	Students    []Students       `gorm:"foreignKey:RoleID;references:ID"`
	Teachers    []Teachers       `gorm:"foreignKey:RoleID;references:ID"`
	Permissions []Rolepermission `gorm:"foreignKey:RoleID;references:ID"`
}

type Permissions struct {
	ID     int    `gorm:"primaryKey"`
	Label  string `gorm:"size:255;unique"`
	Action string `gorm:"type:enum('add','delete','update','query','import','export')"`
	Type   string `gorm:"type:enum('user','role','race','record','permission')"`
}

// Rolepermission 定义角色与权限对应关系的结构体
type Rolepermission struct {
	PermissionID int         `gorm:"column:permissionid;primaryKey"`
	RoleID       int         `gorm:"column:roleid;primaryKey"`
	Permission   Permissions `gorm:"foreignKey:PermissionID;references:ID"`
	Role         Roles       `gorm:"foreignKey:RoleID;references:ID"`
}

type User struct {
	Account   string         `gorm:"column:account;primaryKey;type:varchar(255);not null"`
	Password  string         `gorm:"type:varchar(255);not null"`
	Identity  string         `gorm:"type:varchar(255);not null;check:identity IN ('student', 'teacher')"`
	RoleID    int            `gorm:"index"`
	CreatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Races struct {
	RaceID      int    `gorm:"column:raceid;primaryKey"`
	Title       string `gorm:"size:255"`
	Sponsor     string `gorm:"size:255"`
	Type        string `gorm:"size:255"`
	Level       int
	Location    string `gorm:"size:255"`
	Date        time.Time
	Description string    `gorm:"size:255"`
	Records     []Records `gorm:"foreignKey:RaceID;references:RaceID"`
	CreateTime  time.Time
	UpdateTime  time.Time
}

type Students struct {
	SID        string    `gorm:"column:sid;primaryKey"`
	Name       string    `gorm:"size:255;not null"`
	Password   string    `gorm:"size:255;not null"`
	Sex        int       `gorm:"not null"`
	Grade      int       `gorm:"not null"`
	Class      string    `gorm:"size:255;not null"`
	RoleID     int       `gorm:"index"`
	Records    []Records `gorm:"foreignKey:SID;references:SID"`
	CreateTime time.Time `gorm:"not null"`
	UpdateTime time.Time `gorm:"not null"`
}

type Teachers struct {
	TID         string    `gorm:"column:tid;primaryKey"`
	Name        string    `gorm:"size:255;not null"`
	Password    string    `gorm:"size:255;not null"`
	Rank        int       `gorm:"not null;default:0"`
	Description string    `gorm:"size:255"`
	CreateTime  time.Time `gorm:"not null"`
	UpdateTime  time.Time `gorm:"not null"`
	RoleID      int       `gorm:"index"`
}

type Records struct {
	RecordID    int      `gorm:"column:recordid;primaryKey"`
	Status      int      `gorm:"default:0"`
	Score       string   `gorm:"type:varchar(255)"`
	Description string   `gorm:"type:varchar(255)"`
	SID         string   `gorm:"type:varchar(255)"`
	TID         string   `gorm:"type:varchar(255)"`
	RaceID      int      `gorm:"index"`
	Student     Students `gorm:"foreignKey:SID;references:SID"`
	Teacher     Teachers `gorm:"foreignKey:TID;references:TID"`
	Race        Races    `gorm:"foreignKey:RaceID;references:RaceID"`
	CreateTime  time.Time
	UpdateTime  time.Time
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
