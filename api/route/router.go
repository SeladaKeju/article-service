package route

import (
	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/gin-gonic/gin"
)

func New(health controller.HealthController) *gin.Engine {
	r := gin.Default()
	r.GET("/health", health.Get)
	return r
}
