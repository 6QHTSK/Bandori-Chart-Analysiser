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
		detail, err = model.QueryChartDetail(chartID, diff, chart.AuthorID)
	}
	if err != nil {
		log.Println(err)
	}
	chart.Notes = nil
	author, _ := model.QueryAuthorID(chart.AuthorID)
	result := gin.H{
		"result": true,
		"basic":  chart,
		"detail": detail,
		"diffs":  getDiff(detail),
		"author": author,
	}
	ctx.JSON(http.StatusOK, result)
}

func DiffAnalysisPost(ctx *gin.Context) {
	type inputChart struct {
		Data []Note `json:"data"`
	}
	chart := inputChart{}
	err := ctx.BindJSON(&chart)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	detail := getChartDetail(-1, -1, chart.Data)
	result := gin.H{
		"result": true,
		"detail": detail,
	}
	ctx.JSON(http.StatusOK, result)

}

func ChartNotes(ctx *gin.Context) {
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

	var chart model.Chart

	chart, empty := model.QueryChartBasic(chartID, diff)
	if empty {
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) not found, updating... ", chartID, diff))
		chart, _, err = UpdateChart(chartID, diff)
		if err != nil {
			log.Println(err)
			fail(ctx, err)
			return
		}
		log.Println(fmt.Sprintf("chart (id-%d, diff-%d) updated successfully... ", chartID, diff))
	}
	result := gin.H{
		"result": true,
		"data":   chart.Notes,
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
