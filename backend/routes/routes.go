package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"blog_backend/api"
	"blog_backend/handler"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/api/health", handler.HealthCheck)

	// API
	h := handler.New(db)
	apiGroup := r.Group("/api")
	api.RegisterHandlersWithOptions(
		apiGroup,
		api.NewStrictHandler(h, nil),
		api.GinServerOptions{},
	)
	return r
}
