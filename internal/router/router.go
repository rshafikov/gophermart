package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/gophermart/internal/client"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/handlers"
	"github.com/rshafikov/gophermart/internal/middlewares"
	"github.com/rshafikov/gophermart/internal/service"
	"github.com/rshafikov/gophermart/internal/workerpool"
)

type Router struct {
	UserService    *service.UserService
	OrderService   *service.OrderService
	BalanceService *service.BalanceService
	JWT            security.JWTHandler
	AccrualClient  client.Client
}

func NewRouter(
	u *service.UserService,
	o *service.OrderService,
	b *service.BalanceService,
	jwt security.JWTHandler,
	c client.Client,
) *Router {
	return &Router{
		UserService:    u,
		OrderService:   o,
		BalanceService: b,
		JWT:            jwt,
		AccrualClient:  c}
}

func (mr *Router) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/plain"))
	r.Use(middleware.SetHeader("Content-Type", "application/json; encoding=utf-8"))

	wp := workerpool.NewWorkerPool(2, mr.AccrualClient, mr.OrderService, mr.BalanceService)
	wp.Start()

	userHandler := handlers.NewUserHandler(mr.UserService, mr.JWT)
	orderHandler := handlers.NewOrderHandler(mr.OrderService, wp)
	balanceHandler := handlers.NewBalanceHandler(mr.BalanceService)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Authenticater(mr.JWT, mr.UserService))
				r.Post("/orders", orderHandler.CreateOrder)
				r.Get("/orders", orderHandler.GetOrders)
				r.Get("/balance", balanceHandler.GetUserBalance)
				r.Post("/balance/withdraw", balanceHandler.WithdrawFromBalance)
				r.Get("/withdrawals", balanceHandler.GetUserWithdrawals)
			})
		})
	})

	return r
}
