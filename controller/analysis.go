package controller

import (
	"ayachan/model"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
)

func DiffAnalysis(ctx *gin.Context) {
	var chartID, diff int
	var diffStr, StrID string
	StrID = ctx.Query("id")
	diffStr = ctx.Query("diff")
	chartID, err := strconv.Atoi(StrID)
	if err != nil {
		//fail()
	}
	if diffStr == "" {
		diff = 3
	} else {
		diff, err = strconv.Atoi(diffStr)
		if err != nil {
			//fail()
		}
	}

	var detail model.Detail
	var chart model.Chart
	detail, empty := model.QueryChartDetail(chartID, diff)
	if empty {
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) not found, updating... ", chartID, diff))
		chart, detail, err = UpdateChart(chartID, diff)
		if err != nil {
			log.Println(err)
		}
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) updated successfully... ", chartID, diff))
	}
	result := gin.H{
		"result":  true,
		"basic": chart,
		"detail": detail,
	}
	ctx.JSON(http.StatusOK, result)
}
