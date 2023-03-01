package main

import (
	"github.com/22Fariz22/loyal/internal/app"
	"github.com/22Fariz22/loyal/internal/config"
	"log"
	"os"
)

func main() {
	cfg := config.NewConfig()

	log.Println("cfg from main: ", cfg)

	log.Println("len(os.Args) ", len(os.Args))
	for _, arg := range os.Args[1:] {
		log.Println("arg: ", arg)
	}
	app := app.NewApp(cfg)

	app.Run()
}

//  go run cmd/gophermart/main.go -d="postgres://postgres:55555@127.0.0.1:5432/gophermart" -a="localhost:8080"
