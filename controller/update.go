package controller

import (
	"ayachan/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
)

const (
	officialURL      = "https://player.banground.fun/api/bestdori/official/info/%s/zh"
	officialChartURL = "https://bestdori.com/api/songs/chart/%s.%s.json"
	fanChartURL      = "https://bestdori.reikohaku.fun/api/post/details?id="
)

func UpdateChart(chartID, diff int) (chart model.Chart, detail model.Detail, err error) {
	if chartID <= 500 || chartID == 1000 || chartID == 1001 {
		chart, err = getOfficialChart(chartID, diff)
	} else {
		chart, err = getFanMadeChart(chartID, diff)
	}
	if err != nil {
		return chart, detail, err
	}
	err = model.UpdateBasic(chartID, diff, chart)
	if err != nil {
		return chart, detail, err
	}
	var notes []Note
	for _, note := range chart.Notes {
		notes = append(notes, Note{Note: note})
	}
	detail = getChartDetail(chartID, diff, notes)
	err = model.UpdateDetail(chartID, diff, detail)
	return chart, detail, err
}

// Get chart data from official website: player.banground.fun
func getOfficialChart(chartID, diff int) (chart model.Chart, err error) {
	// Request basic info of a particular chart
	strID := strconv.Itoa(chartID)
	strDiff := strconv.Itoa(diff)
	raw, err := requestForJson(fmt.Sprintf(officialURL, strID))
	if err != nil {
		return chart, err
	}
	var res model.OfficialBasic
	err = json.Unmarshal(raw, &res)

	// Get notes data of a chart:
	reg := regexp.MustCompile(fmt.Sprintf("\"(\\d)\":{\"level\":%d,\"notes\"", diff))
	match := reg.FindSubmatch(raw)
	diffType, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return chart, fmt.Errorf("diff not found")
	}
	diffName := [5]string{"easy", "normal", "hard", "expert", "special"}
	raw, err = requestForJson(fmt.Sprintf(officialChartURL, strID, diffName[diffType]))
	if err != nil {
		return chart, err
	}
	var notes []model.Note
	err = json.Unmarshal(raw, &notes)

	// generate a complete chart for return
	chart = model.Chart{
		Notes:    notes,
		Level:    res.Data.Difficulty[strDiff].Level,
		AuthorID: getAuthorID(),
		Artist:   res.Data.Band,
		Title:    res.Data.Name,
		ID:       chartID,
		Diff:     diff,
	}
	return chart, nil
}

// Get chart data from fan website: bestdori.reikohaku.fun
func getFanMadeChart(chartID, diff int) (chart model.Chart, err error) {
	strID := strconv.Itoa(chartID)
	raw, err := requestForJson(fanChartURL + strID)
	if err != nil {
		return chart, err
	}
	var res model.FanBasic
	err = json.Unmarshal(raw, &res)
	chart = model.Chart{
		Notes:    res.Post.Chart,
		Level:    res.Post.Level,
		AuthorID: getAuthorID(),
		Artist:   res.Post.Artist,
		Title:    res.Post.Title,
		ID:       chartID,
		Diff:     res.Post.Diff,
	}
	return chart, nil
}

func getAuthorID() (ID int) {

	return 0
}

// requestForJson accepts a url and returns raw bytes of json result
func requestForJson(url string) (res []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return res, err
	}
	req.Header.Set("Content-Type", "application/json;charset=utf8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	res, err = ioutil.ReadAll(resp.Body)
	return res, err
}
