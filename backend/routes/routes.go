package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog_backend/health"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/health", health.CheckHandler)

	// API-Route Beispiel
	r.GET("/api/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})

	return r
}
