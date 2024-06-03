package models

// RoleDTO 定义一个新的结构体用于只包含三个字段
type RoleDTO struct {
	ID          int           `json:"id"`
	Label       string        `json:"label"`
	Description string        `json:"description"`
	Permissions []Permissions `json:"permissions"`
}

type AuthenticatedUser struct {
	Account     string   // 账号
	Identity    string   // 身份
	Role        Roles    // 角色
	Permissions []string // 权限
}

type UserData struct {
	Account     string `json:"sid" gorm:"column:sid"`
	Name        string `json:"name"`
	Sex         *int   `json:"sex"`
	Grade       int    `json:"grade"`
	Class       string `json:"class"`
	Rank        int    `json:"rank"`
	Description string `json:"description"`
}
