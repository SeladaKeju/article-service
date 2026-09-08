package route

import (
	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/SeladaKeju/article-service.git/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"time"
)

func New(article controller.ArticleController, timeout time.Duration) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Timeout(timeout))
	r.POST("/articles", article.Create)
	r.GET("/articles", article.List)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
