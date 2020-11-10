package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

const (
	Author_URL = "http://localhost:20009/calcAuthor?author="
	Data_URL   = "http://localhost:20009/calcData"
)

func CalcAuthor(ctx *gin.Context) {
	var authorStr string
	authorStr = ctx.Query("author")
	raw, err := requestForJson(fmt.Sprint(Author_URL, authorStr))
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	var temp map[string]interface{}
	err = json.Unmarshal(raw, &temp)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, temp)
}
func CalcData(ctx *gin.Context) {
	raw, err := requestForJson(Data_URL)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	var temp map[string]interface{}
	err = json.Unmarshal(raw, &temp)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, temp)
}
