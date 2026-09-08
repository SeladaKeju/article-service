package route

import (
	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/SeladaKeju/article-service.git/api/middleware"
	"github.com/gin-gonic/gin"
	"time"
)

func New(article controller.ArticleController, timeout time.Duration) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Timeout(timeout))
	r.POST("/articles", article.Create)
	r.GET("/articles", article.List)
	return r
}
