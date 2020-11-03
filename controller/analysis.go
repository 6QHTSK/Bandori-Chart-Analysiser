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
		fail(ctx, err)
		return
	}
	if diffStr == "" {
		diff = 3
	} else {
		diff, err = strconv.Atoi(diffStr)
		if err != nil {
			fail(ctx, err)
			return
		}
		if diff < 0 || diff > 4 {
			fail(ctx, fmt.Errorf("diff out of range"))
			return
		}
	}

	var detail model.Detail
	var chart model.Chart

	chart, empty := model.QueryChartBasic(chartID, diff)
	if empty {
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) not found, updating... ", chartID, diff))
		chart, detail, err = UpdateChart(chartID, diff)
		if err != nil {
			log.Println(err)
			fail(ctx, err)
			return
		}
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) updated successfully... ", chartID, diff))
	} else {
		detail, err = model.QueryChartDetail(chartID, diff)
	}
	if err != nil {
		log.Println(err)
	}
	chart.Notes = nil
	result := gin.H{
		"result": true,
		"basic":  chart,
		"detail": detail,
	}
	ctx.JSON(http.StatusOK, result)
}

func fail(ctx *gin.Context, err error) {
	result := gin.H{
		"result": false,
		"error":  err.Error(),
	}
	ctx.JSON(http.StatusOK, result)
}
