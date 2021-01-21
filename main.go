package main

import (
	"ayachan/model"
	"ayachan/view"
	"log"
)

func main() {
	log.Println("Starting service...")
	model.InitConfig()
	model.InitTencentCos()
	err := model.MinusTableReader()
	if err != nil {
		log.Println(err)
	}
	err = model.InitDatabase()
	if err != nil {
		log.Println(err)
	}
	r := view.InitRouter()
	err = r.Run(model.RunAddr)
	if err != nil {
		log.Println(err)
	}

}
