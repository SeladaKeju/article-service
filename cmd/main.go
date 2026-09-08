package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/SeladaKeju/article-service.git/api/route"
	"github.com/SeladaKeju/article-service.git/bootstrap"
	"github.com/SeladaKeju/article-service.git/repository"
	"github.com/SeladaKeju/article-service.git/usecase"
	"github.com/gin-gonic/gin"
)

func router(db *sql.DB, timeout time.Duration) *gin.Engine {
	articleRepository := repository.NewArticleRepository(db)
	articleUsecase := usecase.NewArticleUsecase(articleRepository)
	return route.New(controller.NewArticleController(articleUsecase), timeout)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app, err := bootstrap.App(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if err := router(app.DB, app.Env.RequestTimeout).Run(app.Env.Address); err != nil {
		log.Fatal(err)
	}
}
