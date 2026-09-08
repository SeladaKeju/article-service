package main

import (
	"context"
	"log"
	"time"

	"github.com/SeladaKeju/article-service.git/api/controller"
	"github.com/SeladaKeju/article-service.git/api/route"
	"github.com/SeladaKeju/article-service.git/bootstrap"
	"github.com/gin-gonic/gin"
)

func router() *gin.Engine {
	return route.New(controller.HealthController{})
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

	if err := router().Run(config.Address); err != nil {
		log.Fatal(err)
	}
}
