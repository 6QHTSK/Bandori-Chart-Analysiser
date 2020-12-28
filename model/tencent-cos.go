package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io/ioutil"
	"net/http"
)

var diffStr = []string{"easy", "normal", "hard", "expert", "special"}
var bucket *cos.BaseURL
var cosClient *cos.Client

func InitTencentCos() {
	bucket = &cos.BaseURL{BucketURL: cosUrl}
	cosClient = cos.NewClient(bucket, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})
}

func UploadString(name string, string []byte) (success bool) {
	f := bytes.NewReader(string)
	_, err := cosClient.Object.Put(context.Background(), name, f, nil)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func UploadChart(chart Chart) (success bool) {
	name := fmt.Sprintf("chart/%d.%s.json", chart.ID, diffStr[chart.Diff])
	chartStr, _ := json.Marshal(chart.Notes)
	return UploadString(name, chartStr)
}

func UploadSonolusChart(chartId int, SonolusScript []byte) (success bool) {
	name := fmt.Sprintf("Sonolus/%d/level.json", chartId)
	return UploadString(name, SonolusScript)
}

func DownloadChart(chartID int, diff int) (chart Chart, success bool) {
	name := fmt.Sprintf("%d.%s.json", chartID, diffStr[diff])
	resp, err := cosClient.Object.Get(context.Background(), name, nil)
	if cos.IsNotFoundError(err) {
		return chart, false
	}
	byteChart, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(byteChart, &chart.Notes)
	_ = resp.Body.Close()
	return chart, true
}
