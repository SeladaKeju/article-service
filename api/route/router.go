package route

import (
	"net/http"
	"time"

	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/SeladaKeju/article-service.git/api/middleware"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// New registers HTTP routes and middleware for the article API.
func New(article controller.ArticleController, timeout time.Duration) *gin.Engine {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/swagger" || c.Request.URL.Path == "/swagger/" {
			c.Redirect(http.StatusFound, "/swagger/index.html")
			c.Abort()
			return
		}
		c.Next()
	})
	r.Use(middleware.Timeout(timeout))
	r.POST("/articles", article.Create)
	r.GET("/articles", article.List)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return r
}
