package models

import "time"

// Role 角色表
// 角色编码作为业务标识（admin/teacher/student_cadre/student），与现有 users.roles 字符串保持一致
type Role struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:32;not null" json:"code"`         // 角色编码：admin/teacher/student_cadre/student
	Name        string    `gorm:"size:64;not null" json:"name"`                     // 角色中文名
	DataScope   string    `gorm:"size:16;not null;default:self" json:"dataScope"`   // 数据作用域：all/college/team/self
	Description string    `gorm:"type:text" json:"description"`
	Builtin     bool      `gorm:"not null;default:false" json:"builtin"` // 内置角色不可删除
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Permission 权限点（菜单 / API / 按钮）
// Code 形如 menu:dashboard、menu:members、api:member:list
type Permission struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Code        string    `gorm:"uniqueIndex;size:64;not null" json:"code"` // 权限码
	Name        string    `gorm:"size:128;not null" json:"name"`            // 名称
	Module      string    `gorm:"size:32;not null;index" json:"module"`     // 所属模块：menu/member/activity/community/workstudy/admin/dashboard
	Type        string    `gorm:"size:16;not null;default:api" json:"type"` // 类型：menu/api/button
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// RolePermission 角色-权限关联表
type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"roleId"`
	PermissionID uint      `gorm:"primaryKey" json:"permissionId"`
	CreatedAt    time.Time `json:"createdAt"`
}

// TableName 显式指定表名，避免 GORM 默认复数规则误差
func (RolePermission) TableName() string { return "role_permissions" }
