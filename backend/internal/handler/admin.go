package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// AdminHandler 管理员后台相关接口
// 仅 admin 角色可访问；路由层通过 middleware.RequireRole("admin") 控制
type AdminHandler struct {
	DB *gorm.DB
}

// NewAdminHandler 创建管理员处理器
func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{DB: db}
}

// ListUsers 用户列表
// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Order("id ASC").Find(&users).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "加载用户失败")
		return
	}
	// 不直接返回 model（带 json:"-" 的密码已隐藏），手动整理字段保持前端契约稳定
	items := make([]gin.H, 0, len(users))
	for _, u := range users {
		items = append(items, gin.H{
			"id":        u.ID,
			"username":  u.Username,
			"email":     u.Email,
			"realName":  u.RealName,
			"college":   u.College,
			"className": u.ClassName,
			"roles":     u.Roles,
			"status":    u.Status,
			"createdAt": u.CreatedAt,
		})
	}
	response.OK(c, items)
}

// ListRoles 角色列表
// GET /api/v1/admin/roles
func (h *AdminHandler) ListRoles(c *gin.Context) {
	var roles []models.Role
	if err := h.DB.Order("id ASC").Find(&roles).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "加载角色失败")
		return
	}
	response.OK(c, roles)
}

// ListPermissions 权限列表
// GET /api/v1/admin/permissions
func (h *AdminHandler) ListPermissions(c *gin.Context) {
	var permissions []models.Permission
	if err := h.DB.Order("module, id").Find(&permissions).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "加载权限失败")
		return
	}
	response.OK(c, permissions)
}
