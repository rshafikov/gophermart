package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/gophermart/internal/app"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/repository"
	"github.com/rshafikov/gophermart/internal/router"
	"github.com/rshafikov/gophermart/internal/service"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	app.InitConfig()

	Application := app.NewApplication(app.Config)
	err := Application.ConnectToDatabase(context.TODO())
	if err != nil {
		logger.L.Fatal("database connect error", zap.Error(err))
	}

	jwtHanlder := security.NewJWTHandler()
	userRepository := repository.NewUserRepository(Application.DB.Pool)
	userService := service.NewUserService(userRepository)
	mainRouter := router.NewRouter(userService, jwtHanlder)
	r := chi.NewRouter()
	r.Mount("/", mainRouter.Routes())

	err = http.ListenAndServe(app.Config.RunAddress.String(), r)
	if err != nil {
		log.Fatal(err)
	}
}
