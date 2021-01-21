package controller

import (
	"ayachan/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func CalcAuthor(ctx *gin.Context) {
	var authorStr string
	authorStr = ctx.Query("author")
	AuthorUrl := fmt.Sprint(model.PythonAPI, "calcAuthor?author=", authorStr)
	ctx.Redirect(http.StatusMovedPermanently, AuthorUrl)
}
func CalcData(ctx *gin.Context) {
	DataUrl := fmt.Sprint(model.PythonAPI, "calcData")
	ctx.Redirect(http.StatusMovedPermanently, DataUrl)
}
func SongList(ctx *gin.Context) {
	SongListUrl := fmt.Sprint(model.PythonAPI, "songList")
	ctx.Redirect(http.StatusMovedPermanently, SongListUrl)
}

func BestdoriSearch(ctx *gin.Context) {
	var page string
	searchStr := ctx.Param("search")
	page = ctx.Query("page")
	BestdoriSearchUrl := fmt.Sprint(model.PythonAPI, "bestdoriSearch/", searchStr, "?page=", page)
	ctx.Redirect(http.StatusMovedPermanently, BestdoriSearchUrl)
}
