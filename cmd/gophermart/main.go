package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/rshafikov/gophermart/internal/app"
	"github.com/rshafikov/gophermart/internal/client"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/repository"
	"github.com/rshafikov/gophermart/internal/router"
	"github.com/rshafikov/gophermart/internal/service"
)

func main() {
	app.InitConfig()

	Application := app.NewApplication(app.Config)
	Application.ConnectToDatabase(context.Background())
	Application.MigrateDatabase(context.Background())

	jwt := security.NewJWTHandler()
	accrualClient := client.NewAccrualClient(Application.Config.AccrualAddress.String())

	userRepository := repository.NewUserRepository(Application.DB.Pool)
	orderRepository := repository.NewOrderRepository(Application.DB.Pool)
	balanceRepository := repository.NewBalanceRepository(Application.DB.Pool)
	txRepository := repository.NewTxRepository(Application.DB.Pool)

	userService := service.NewUserService(userRepository)
	orderService := service.NewOrderService(orderRepository)
	balanceService := service.NewBalanceService(balanceRepository, txRepository)

	mainRouter := router.NewRouter(userService, orderService, balanceService, jwt, accrualClient)
	r := chi.NewRouter()
	r.Mount("/", mainRouter.Routes())

	Application.RunServer(r)
}
