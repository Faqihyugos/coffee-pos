package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NewRouter initialises a Gin engine with global middleware and all routes.
// appEnv controls the Gin mode: "production" uses ReleaseMode, everything
// else uses DebugMode.
func NewRouter(appEnv string) *gin.Engine {
	if appEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthCheck)
	}

	return r
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "server is running",
	})
}
