package models

import "time"

// User 用户模型，映射 users 表
// 密码字段通过 json:"-" 在序列化时自动隐藏，避免泄露
type User struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Username  string `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Password  string `gorm:"size:255;not null" json:"-"` // JSON 响应中隐藏密码
	Email     string `gorm:"size:128" json:"email"`
	Roles     string `gorm:"size:255;default:student" json:"roles"` // 逗号分隔，如 "admin,teacher,student"
	// Sprint6 / RBAC：组织信息，用于数据权限作用域
	RealName  string    `gorm:"size:64" json:"realName"`                       // 真实姓名
	College   string    `gorm:"size:64;index" json:"college"`                  // 所属学院/部门
	ClassName string    `gorm:"size:64;index" json:"className"`                // 班级（学生）
	Status    string    `gorm:"size:16;not null;default:active" json:"status"` // active/disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
