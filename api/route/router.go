package route

import (
	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/gin-gonic/gin"
)

func New(health controller.HealthController, article controller.ArticleController) *gin.Engine {
	r := gin.Default()
	r.GET("/health", health.Get)
	r.POST("/articles", article.Create)
	return r
}
