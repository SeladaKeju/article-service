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

func router(db *sql.DB) *gin.Engine {
	articleRepository := repository.NewArticleRepository(db)
	articleUsecase := usecase.NewArticleUsecase(articleRepository)
	return route.New(controller.HealthController{}, controller.NewArticleController(articleUsecase))
}

func main() {
	config, err := bootstrap.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := bootstrap.OpenDatabase(ctx, config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := router(db).Run(config.Address); err != nil {
		log.Fatal(err)
	}
}
