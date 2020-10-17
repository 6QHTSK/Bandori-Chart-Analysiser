package view

import (
	"ayachan/controller"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine{
	r := gin.Default()
	mainGroup := r.Group("")
	{
		mainGroup.GET("/DiffAnalysis",controller.DiffAnalysis)
		mainGroup.POST("/DiffAnalysis",controller.DiffAnalysis)
	}
	return r
}
