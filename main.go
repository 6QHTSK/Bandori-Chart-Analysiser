package main

import (
	"ayachan/model"
	"ayachan/view"
	"log"
)

func main() {
	log.Println("Starting service...")
	err := model.InitDatabase()
	if err != nil {
		log.Println(err)
	}
	r := view.InitRouter()
	err = r.Run("0.0.0.0:17555")
	if err != nil {
		log.Println(err)
	}

}
