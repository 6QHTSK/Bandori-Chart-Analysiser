package view

import (
	"ayachan/controller"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))
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
