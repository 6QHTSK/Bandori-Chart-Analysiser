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
	var diffStr, StrID, speedStr string
	StrID = ctx.Query("id")
	diffStr = ctx.Query("diff")
	speedStr = ctx.Query("speed")
	speed, err := strconv.ParseFloat(speedStr, 32)
	if err != nil {
		speed = 1.0
	}
	chartID, err = strconv.Atoi(StrID)
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
	if speed == -1.0 {
		result := gin.H{
			"result": true,
			"basic":  chart,
			"detail": detail,
			"author": author,
		}
		ctx.JSON(http.StatusOK, result)
		return
	}
	multipleSpeed(&detail, float32(speed))
	result := gin.H{
		"result": true,
		"basic":  chart,
		"detail": detail,
		"diff":   getDiff(detail),
		"author": author,
	}
	ctx.JSON(http.StatusOK, result)
}

func DiffAnalysisPost(ctx *gin.Context) {
	type inputChart struct {
		Data  []Note  `json:"data"`
		Diff  int     `json:"diff"`
		Speed float32 `json:"speed"`
	}
	chart := inputChart{}
	err := ctx.BindJSON(&chart)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	detail := getChartDetail(-1, chart.Diff, chart.Data)
	multipleSpeed(&detail, chart.Speed)
	result := gin.H{
		"result": true,
		"detail": detail,
		"diff":   getDiff(detail),
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
	var status bool
	chart, status = model.DownloadChart(chartID, diff)
	if !status {
		if chartID <= 1100 {
			chart, err = GetOfficialChart(chartID, diff)
			if chart.Level == 0 {
				chart, err = GetFanMadeChart(chartID)
			}
		} else {
			chart, err = GetFanMadeChart(chartID)
		}
		if err == nil {
			model.UploadChart(chart)
		}
	}
	result := gin.H{
		"result": true,
		"data":   chart.Notes,
		"err":    err,
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
