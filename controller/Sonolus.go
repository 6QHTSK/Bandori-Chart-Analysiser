package controller

import (
	"ayachan/model"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const reCaptchaServer = "https://recaptcha.net/recaptcha/api/siteverify?secret=%s&response=%s"

func BDFan2Sonolus(chart []Note) (SonolusScript []byte, err error) {
	notesWithTime, _, _, _, _, _ := calcTime(chart)
	var SonolusEntities = []model.SonolusEntity{{Archetype: 0}, {Archetype: 1}}
	var pos = map[string]float32{"A": 0.0, "B": 0.0}
	var lastnotetime float32
	for i, note := range notesWithTime {
		time := note.Time
		lane := float32(note.Lane - 4)
		var SonolusEntity model.SonolusEntity
		if note.Pos != "" {
			if note.Start {
				SonolusEntity.Archetype = 3 // 绿条起点.
				SonolusEntity.Data.Index = 1
				SonolusEntity.Data.Values = []float32{time, lane}
				if pos[note.Pos] != 0.0 {
					return nil, fmt.Errorf("Unexpected start note!")
				}
				pos[note.Pos] = float32(i + 2)
			} else if note.End {
				if pos[note.Pos] == 0.0 {
					return nil, fmt.Errorf("Unexpected end note!")
				}
				SonolusEntity.Data.Index = 0
				SonolusEntity.Data.Values = []float32{pos[note.Pos], time, lane}
				if note.Flick {
					SonolusEntity.Archetype = 7 // 绿条尾粉
				} else {
					SonolusEntity.Archetype = 6 // 绿条尾
				}
				pos[note.Pos] = 0.0
			} else {
				if pos[note.Pos] == 0.0 {
					return nil, fmt.Errorf("Unexpected tick note!")
				}
				SonolusEntity.Data.Index = 0
				SonolusEntity.Archetype = 5 // 绿条节点
				SonolusEntity.Data.Values = []float32{pos[note.Pos], time, lane}
				pos[note.Pos] = float32(i + 2)
			}
		} else {
			SonolusEntity.Data.Index = 1
			SonolusEntity.Data.Values = []float32{time, lane}
			if note.Flick {
				SonolusEntity.Archetype = 4 // 粉键
			} else {
				SonolusEntity.Archetype = 2 // 蓝键
			}
		}
		if lastnotetime == time {
			SonolusEntity.Data.Values = append(SonolusEntity.Data.Values, 1.0)
		}
		SonolusEntities = append(SonolusEntities, SonolusEntity)
	}
	SonolusScriptCore, _ := json.Marshal(SonolusEntities)
	SonolusScriptFront, _ := ioutil.ReadFile("basic_front.txt")
	SonolusScriptBack, _ := ioutil.ReadFile("basic_back.txt")
	SonolusScript = append(SonolusScript, SonolusScriptFront...)
	SonolusScript = append(SonolusScript, SonolusScriptCore...)
	SonolusScript = append(SonolusScript, SonolusScriptBack...)
	return SonolusScript, nil
}

func SonolusScriptPost(ctx *gin.Context) {
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
	result, err := BDFan2Sonolus(chart.Data)
	if err == nil {
		ctx.String(http.StatusOK, string(result))
	} else {
		log.Println(err)
		fail(ctx, err)
	}
}

func UploadSonolusTest(ctx *gin.Context) {
	type uploadInfo struct {
		MusicUrl          string `json:"bgm"`
		Title             string `json:"title"`
		Notes             []Note `json:"notes"`
		RecaptchaResponse string `json:"g-recaptcha-response"`
	}
	type RecaptchaResponse struct {
		Success bool `json:"success"`
	}

	info := uploadInfo{}
	err := ctx.BindJSON(&info)
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	// 检查请求合法性：Recaptcha
	response := RecaptchaResponse{}
	raw, err := requestForJson(fmt.Sprintf(reCaptchaServer, model.ReCaptchaKey, info.RecaptchaResponse))
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	_ = json.Unmarshal(raw, &response)
	if !response.Success {
		fail(ctx, fmt.Errorf("Recaptcha Not Pass"))
		return
	}

	Script, err := BDFan2Sonolus(info.Notes) // 将发来的Bestdori Fanmade谱面转换为Sonolus脚本
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}

	// 生成随机ID
	rand.Seed(time.Now().Unix())
	var id int
	id = rand.Intn(1000)
	for model.CheckSonolusList(id) {
		id = rand.Intn(1000)
	}
	var listItem = model.SonolusList{
		ID:      id,
		Title:   info.Title,
		Artists: "UploadAt: " + time.Now().Format("2006/01/02 15:04"),
		Author:  "ID: " + strconv.Itoa(id),
		Cover:   "https://assets.ayachan.fun/pic/black.png",
		BGM:     info.MusicUrl,
		Difficulties: model.SonolusDifficulty{
			Name:   strconv.Itoa(id),
			Rating: "20",
		},
		Upload: time.Now().Unix(),
	}
	model.UpdateSonolusList(listItem)
	model.UploadSonolusChart(id, Script)
	ctx.JSON(http.StatusOK, gin.H{
		"result": true,
		"id":     id,
	})
}

func GetSonolusList(ctx *gin.Context) {
	list, err := model.GetSonolusList()
	if err != nil {
		log.Println(err)
		fail(ctx, err)
		return
	}
	//listByte,_ := json.Marshal(list)
	ctx.JSON(http.StatusOK, gin.H{
		"list":      list,
		"pageCount": "1",
	})
}
