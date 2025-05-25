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
)

func main() {
	app.InitConfig()

	Application := app.NewApplication(app.Config)
	if err := Application.ConnectToDatabase(context.Background()); err != nil {
		logger.L.Fatal("database connect error", zap.Error(err))
	}

	jwtHanlder := security.NewJWTHandler()
	userRepository := repository.NewUserRepository(Application.DB.Pool)
	orderRepository := repository.NewOrderRepository(Application.DB.Pool)
	userService := service.NewUserService(userRepository)
	orderService := service.NewOrderService(orderRepository)
	mainRouter := router.NewRouter(userService, orderService, jwtHanlder)
	r := chi.NewRouter()
	r.Mount("/", mainRouter.Routes())

	Application.RunServer(r)
}
