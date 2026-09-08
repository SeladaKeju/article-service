package controller

import "github.com/gin-gonic/gin"

type HealthController struct{}

func (HealthController) Get(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
