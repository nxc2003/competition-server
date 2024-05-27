package models

// RoleDTO 定义一个新的结构体用于只包含三个字段
type RoleDTO struct {
	ID          int    `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}
