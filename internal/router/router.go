package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rshafikov/gophermart/internal/core/security"
	"github.com/rshafikov/gophermart/internal/handlers"
	"github.com/rshafikov/gophermart/internal/middlewares"
	"github.com/rshafikov/gophermart/internal/service"
	"net/http"
)

type Router struct {
	UserService  *service.UserService
	OrderService *service.OrderService
	JWT          security.JWTHandler
}

func NewRouter(u *service.UserService, o *service.OrderService, jwt security.JWTHandler) *Router {
	return &Router{UserService: u, OrderService: o, JWT: jwt}
}

func (mr *Router) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middlewares.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5, "application/json", "text/plain"))

	userHandler := handlers.NewUserHandler(mr.UserService, mr.JWT)
	orderHandler := handlers.NewOrderHandler(mr.OrderService)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
			r.Group(func(r chi.Router) {
				r.Use(middlewares.Authenticater(mr.JWT, mr.UserService))
				r.Post("/orders", orderHandler.CreateOrder)
				r.Get("/orders", orderHandler.GetOrders)
				r.Get("/balance", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				r.Post("/balance/withdraw", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				r.Get("/withdrawals", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			})
		})
	})

	return r
}
