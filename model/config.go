package model

import (
	"encoding/csv"
	"fmt"
	"github.com/Unknwon/goconfig"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var MongoURL string
var RunAddr string
var PythonAPI string
var secretId string
var secretKey string
var cosUrl *url.URL
var MinusTableUrl string
var ReCaptchaKey string
var Debug string
var MinusTable [121][2]float64

func InitConfig() {
	cfg, err := goconfig.LoadConfigFile("conf.ini")
	if err != nil {
		f, err := os.Create("conf.ini")
		//判断是否出错
		if err != nil {
			fmt.Println("Failed to open config file: ", err)
			return
		}
		defer f.Close()
		cfg, _ = goconfig.LoadConfigFile("conf.ini")
	}
	MongoURL = initStringConfig(cfg, "mongoURL", "mongodb://localhost:27017/")
	RunAddr = initStringConfig(cfg, "RunAddr", "0.0.0.0:20008")
	PythonAPI = initStringConfig(cfg, "PythonAPI", "http://localhost:20009/")
	secretId = initStringConfig(cfg, "secretId", "")
	secretKey = initStringConfig(cfg, "secretKey", "")
	ReCaptchaKey = initStringConfig(cfg, "reCaptchaKey", "")
	cosUrl, _ = url.Parse(initStringConfig(cfg, "cosUrl", "0.0.0.0"))
	Debug = initStringConfig(cfg, "debug", "debug")
	MinusTableUrl = initStringConfig(cfg, "minusTableUrl", "minusTable.csv")
	err = goconfig.SaveConfigFile(cfg, "conf.ini")
	if err != nil {
		fmt.Println("Failed to Save config file: ", err)
		return
	}
}

func initStringConfig(cfg *goconfig.ConfigFile, key string, defaultValue string) (value string) {
	var err error
	value, err = cfg.GetValue("", key)
	if err != nil {
		_ = cfg.SetValue("", key, defaultValue)
		return defaultValue
	}
	return value
}

func MinusTableReader() (err error) {
	file, err := os.Open(MinusTableUrl)
	if err != nil {
		return err
	}
	reader := csv.NewReader(file)
	stringTable, err := reader.ReadAll()
	if err != nil {
		return err
	}
	for key, lines := range stringTable {
		MinusTable[key][0], err = strconv.ParseFloat(strings.Trim(lines[1], " "), 64)
		if err != nil {
			return err
		}
		MinusTable[key][1], err = strconv.ParseFloat(strings.Trim(lines[2], " "), 64)
		if err != nil {
			return err
		}
	}
	return nil
}
