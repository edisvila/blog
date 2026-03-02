package routes

import (
	"blog_backend/api"
	"blog_backend/handler"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/api/health", handler.HealthCheck)

	h := handler.New(db)
	strictHandler := api.NewStrictHandler(h, nil)

	// Public routes
	public := r.Group("/api")
	public.POST("/auth/login", func(c *gin.Context) { strictHandler.PostAuthLogin(c) })
	public.GET("/posts", func(c *gin.Context) { strictHandler.GetPosts(c) })
	public.GET("/posts/:slug", func(c *gin.Context) { strictHandler.GetPostsSlug(c, c.Param("slug")) })

	// Protected routes
	protected := r.Group("/api")
	protected.Use(handler.AuthMiddleware())
	protected.POST("/posts", func(c *gin.Context) { strictHandler.PostPosts(c) })
	protected.PUT("/posts/:slug", func(c *gin.Context) { strictHandler.PutPostsSlug(c, c.Param("slug")) })
	protected.DELETE("/posts/:slug", func(c *gin.Context) { strictHandler.DeletePostsSlug(c, c.Param("slug")) })

	return r
}
