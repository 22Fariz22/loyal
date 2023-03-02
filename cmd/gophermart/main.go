package main

import (
	"github.com/22Fariz22/loyal/internal/app"
	"github.com/22Fariz22/loyal/internal/config"
)

func main() {
	cfg := config.NewConfig()

	app := app.NewApp(cfg)
	app.Run()
}

//  go run cmd/gophermart/main.go -d="postgres://postgres:55555@127.0.0.1:5432/gophermart" -a="localhost:8080"
