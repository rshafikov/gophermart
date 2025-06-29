package app

import (
	"context"
	"errors"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/database"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const dbConnectionTimeout = 10 * time.Second

type Application struct {
	Config defaultConfig
	DB     *database.DB
}

func NewApplication(cfg defaultConfig) *Application {
	return &Application{
		Config: cfg,
		DB:     &database.DB{},
	}
}

func (app *Application) SetupDB() {
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectionTimeout)
	defer cancel()
	err := app.DB.ConnectToDatabase(ctx, app.Config.DB.URI)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.L.Fatal("db connection timeout", zap.Error(err))
		}
		if errors.Is(err, database.ErrDB) {
			logger.L.Fatal("unable to connect to db", zap.Error(err))
		} else if errors.Is(err, database.ErrConnectDB) {
			logger.L.Fatal("db connection error", zap.Error(err))
		}
	}
}

func (app *Application) RunServer(router http.Handler) {
	server := http.Server{Addr: app.Config.RunAddress.String(), Handler: router}
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sigReceived := <-sig
		logger.L.Debug("received shutdown signal", zap.String("signal", sigReceived.String()))
		shutdownCtx, shutdownCancelCtx := context.WithTimeout(serverCtx, 5*time.Second)
		defer shutdownCancelCtx()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.L.Fatal("graceful shutdown timed out...forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.L.Fatal("shutdowning error", zap.Error(err))
		}
		serverStopCtx()
		logger.L.Debug("graceful shutdown completed")
	}()

	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.L.Fatal("listening error", zap.Error(err))
	}

	<-serverCtx.Done()
}
