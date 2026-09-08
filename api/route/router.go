package route

import (
	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/gin-gonic/gin"
)

func New(article controller.ArticleController) *gin.Engine {
	r := gin.Default()
	r.POST("/articles", article.Create)
	r.GET("/articles", article.List)
	return r
}
