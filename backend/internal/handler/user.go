package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/authz"
	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// UserHandler 用户信息相关操作
type UserHandler struct {
	DB *gorm.DB
}

// NewUserHandler 创建用户信息处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

// GetProfile 获取当前登录用户信息（含 RBAC 授权信息）
// GET /api/v1/users/me
// 返回字段：基础用户信息 + roles + permissions + menus + dataScope
// 前端登录成功后调用以构建动态路由与菜单
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDValue, _ := c.Get("userId")

	// 从数据库加载用户基础信息（fallback 用上下文中的 roles 字符串）
	roles := c.GetString("roles")
	var user models.User
	if userIDValue != nil {
		// JWT 中的 userId 是 float64（JSON 数字默认类型），需要兼容
		var uid uint
		switch v := userIDValue.(type) {
		case float64:
			uid = uint(v)
		case uint:
			uid = v
		case int:
			uid = uint(v)
		}
		if uid > 0 {
			if err := h.DB.First(&user, uid).Error; err == nil {
				roles = user.Roles
			}
		}
	}

	// 聚合授权信息
	az := authz.Resolve(h.DB, roles)

	response.OK(c, gin.H{
		"userId":      user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"realName":    user.RealName,
		"college":     user.College,
		"className":   user.ClassName,
		"roles":       az.Roles,
		"permissions": az.Permissions,
		"menus":       az.Menus,
		"dataScope":   az.DataScope,
	})
}
