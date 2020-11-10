package view

import (
	"ayachan/controller"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	mainGroup := r.Group("")
	{
		mainGroup.GET("/DiffAnalysis", controller.DiffAnalysis)
		mainGroup.POST("/DiffAnalysis", controller.DiffAnalysisPost)
		mainGroup.GET("/ChartNotes", controller.ChartNotes)
		mainGroup.StaticFile("/songList", "./songList.json")
		mainGroup.GET("/calcData", controller.CalcData)
		mainGroup.GET("/calcAuthor", controller.CalcAuthor)
	}
	return r
}
