package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: "OK", Message: "success", Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Body{Code: "CREATED", Message: "created", Data: data})
}

func Fail(c *gin.Context, code string, message string) {
	c.JSON(http.StatusBadRequest, Body{Code: code, Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, Body{Code: "UNAUTHORIZED", Message: message})
}

func Forbidden(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, Body{Code: "FORBIDDEN", Message: message})
}