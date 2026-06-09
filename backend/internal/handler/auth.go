package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/authz"
	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// AuthHandler 认证相关操作
type AuthHandler struct {
	DB        *gorm.DB // 数据库连接
	JWTSecret string   // JWT 签名密钥
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(db *gorm.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{DB: db, JWTSecret: jwtSecret}
}

// ---------- 请求/响应结构体 ----------

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`  // 用户名，3-64 字符
	Password string `json:"password" binding:"required,min=6,max=128"` // 密码，6-128 字符
	Email    string `json:"email"`                                     // 邮箱，可选
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"` // 用户名
	Password string `json:"password" binding:"required"` // 密码
}

// ---------- 处理器方法 ----------

// Register 用户注册
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	// 检查用户名是否已被占用
	var count int64
	h.DB.Model(&models.User{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		response.Fail(c, "USER_EXISTS", "用户名已存在")
		return
	}

	// bcrypt 加密密码，确保数据库中不存明文
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Fail(c, "INTERNAL_ERROR", "密码加密失败")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Roles:    "student", // 新注册用户默认角色为 student
	}

	if err := h.DB.Create(&user).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建用户失败")
		return
	}

	response.Created(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// Login 用户登录
// POST /api/v1/auth/login
// 验证用户名密码，返回 JWT Token 和用户基本信息
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	// 根据用户名查找用户
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Fail(c, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	// 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		response.Fail(c, "INVALID_CREDENTIALS", "用户名或密码错误")
		return
	}

	// 生成 JWT，有效期 24 小时
	claims := jwt.MapClaims{
		"userId":   user.ID,
		"username": user.Username,
		"roles":    user.Roles,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		response.Fail(c, "INTERNAL_ERROR", "Token 生成失败")
		return
	}

	// 聚合 RBAC 授权信息：角色、权限、菜单、数据作用域
	az := authz.Resolve(h.DB, user.Roles)

	response.OK(c, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":          user.ID,
			"username":    user.Username,
			"email":       user.Email,
			"realName":    user.RealName,
			"college":     user.College,
			"className":   user.ClassName,
			"roles":       az.Roles,
			"permissions": az.Permissions,
			"menus":       az.Menus,
			"dataScope":   az.DataScope,
		},
	})
}
