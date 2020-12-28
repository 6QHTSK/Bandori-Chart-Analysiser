package controller

import (
	"ayachan/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	officialURL      = "https://player.banground.fun/api/bestdori/official/info/%s/zh"
	officialURL2     = "https://player.banground.fun/api/bestdori/official/info/%s/en"
	officialChartURL = "https://bestdori.com/api/songs/chart/%s.%s.json"
	fanChartURL      = "https://bestdori.com/api/post/details?id="
)

func UpdateChart(chartID, diff int) (chart model.Chart, detail model.Detail, err error) {
	if chartID <= 1100 {
		chart, err = GetOfficialChart(chartID, diff)
		if chart.Level == 0 {
			chart, err = GetFanMadeChart(chartID, diff)
		}
	} else {
		chart, err = GetFanMadeChart(chartID, diff)
	}
	if chart.Level == 0 {
		if err == nil {
			err = fmt.Errorf("Not A Chart")
		}
		return chart, detail, err
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
	detail = getChartDetail(chartID, chart.Diff, notes)
	err = model.UpdateDetail(chartID, chart.Diff, detail)
	return chart, detail, err
}

// Get chart data from official website: player.banground.fun
func GetOfficialChart(chartID, diff int) (chart model.Chart, err error) {
	// Request basic info of a particular chart
	strID := strconv.Itoa(chartID)
	strDiff := strconv.Itoa(diff)
	basicURL := officialURL
	if chartID >= 1000 {
		basicURL = officialURL2
	}
	raw, err := requestForJson(fmt.Sprintf(basicURL, strID))
	if err != nil {
		return chart, err
	}
	var res model.OfficialBasic
	err = json.Unmarshal(raw, &res)

	// Get notes data of a chart:
	diffName := [5]string{"easy", "normal", "hard", "expert", "special"}
	raw, err = requestForJson(fmt.Sprintf(officialChartURL, strID, diffName[diff]))
	if err != nil {
		return chart, err
	}
	var notes []model.Note
	err = json.Unmarshal(raw, &notes)
	var fanNotes []model.Note
	fanNotes, err = model.BD2BDFan(notes)
	if err != nil {
		return chart, err
	}

	// generate a complete chart for return
	chart = model.Chart{
		Notes:    fanNotes,
		Level:    res.Data.Difficulty[strDiff].Level,
		AuthorID: 0, //getAuthorID(),
		Artist:   res.Data.Band,
		Title:    res.Data.Name,
		ID:       chartID,
		Diff:     diff,
	}
	return chart, nil
}

// Get chart data from fan website: bestdori.reikohaku.fun
func GetFanMadeChart(chartID, diff int) (chart model.Chart, err error) {
	strID := strconv.Itoa(chartID)
	raw, err := requestForJson(fanChartURL + strID)
	if err != nil {
		return chart, err
	}
	var res model.FanBasic
	err = json.Unmarshal(raw, &res)
	var AuthorInfo model.Author
	AuthorInfo, err = getAuthor(res.Post.Author.Username, res.Post.Author.Nickname)
	chart = model.Chart{
		Notes:    res.Post.Chart,
		Level:    res.Post.Level,
		AuthorID: AuthorInfo.AuthorID,
		Artist:   res.Post.Artist,
		Title:    res.Post.Title,
		ID:       chartID,
		Diff:     res.Post.Diff,
	}
	return chart, nil
}

func getAuthor(username string, nickname string) (AuthorInfo model.Author, err error) {
	var empty bool
	AuthorInfo, empty = model.QueryAuthor(username)
	if empty || AuthorInfo.NickName != nickname {
		err = model.UpdateAuthor(username, nickname)
		AuthorInfo, empty = model.QueryAuthor(username)
	}
	return AuthorInfo, err
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
