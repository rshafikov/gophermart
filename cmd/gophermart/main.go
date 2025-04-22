package main

import (
	"context"
	"github.com/rshafikov/gophermart/internal/app"
	"log"
)

func main() {
	app.InitConfig()

	Application := app.NewApplication(app.Config)
	err := Application.ConnectToDatabase(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
