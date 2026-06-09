package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"unistep-platform/backend/internal/response"
)

// JWTAuth 校验 Authorization: Bearer <token>，并把用户信息写入 Gin 上下文
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "missing bearer token")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 只允许 HMAC 签名算法，避免算法降级风险
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Unauthorized(c, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Unauthorized(c, "invalid token claims")
			return
		}

		// Gin 上下文供后续 handler 或权限中间件读取
		c.Set("userId", claims["userId"])
		c.Set("username", claims["username"])
		c.Set("roles", claims["roles"])
		c.Next()
	}
}

// RequireRole 校验当前用户是否拥有指定角色
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles := strings.Split(c.GetString("roles"), ",")
		for _, item := range roles {
			if strings.TrimSpace(item) == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "permission denied"})
	}
}
