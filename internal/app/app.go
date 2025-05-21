package app

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/database"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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

func (app *Application) ConnectToDatabase(ctx context.Context) error {
	dsn := app.Config.DB.URI
	_, err := pgx.Connect(ctx, dsn)
	if err != nil {
		var pgErr *pgconn.ConnectError
		if errors.As(err, &pgErr) {
			logger.L.Debug("unable to connect to database", zap.String("DB_URI", dsn))
			return database.ErrConnectDB
		}
		return err
	}

	app.DB.Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}

	err = app.DB.Pool.Ping(ctx)
	if err != nil {
		return err
	}
	log.Println("Connected to database:", dsn)
	return nil
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
