package main

import (
	"log"
	"main/server"
	"main/service"
)

func main() {
	svc := service.NewService("data.csv")
	err := svc.Init()
	if err != nil {
		log.Fatal(err)
	}

	srv := server.NewServer(svc)
	log.Fatal(srv.Run())
}
