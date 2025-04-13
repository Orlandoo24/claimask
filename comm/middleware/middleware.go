package middleware

import (
	"claimask/comm/initialize"

	"github.com/gin-gonic/gin"
)

// Placeholder file for middleware package.

// InitMiddleware 初始化全局中间件
func InitMiddleware(server *initialize.Server) {
	// 在这里添加全局中间件
	// 例如：server.Use(gin.Logger())
}

// AuthMiddleware bk auth middleware
func AuthMiddleware(actions ...string) gin.HandlerFunc {
	return nil
}
