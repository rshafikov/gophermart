package app

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"github.com/rshafikov/gophermart/internal/database"
	"go.uber.org/zap"
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

func (app *Application) ConnectToDatabase(ctx context.Context) {
	dsn := app.Config.DB.URI
	_, err := pgx.Connect(ctx, dsn)
	if err != nil {
		var pgErr *pgconn.ConnectError
		if errors.As(err, &pgErr) {
			logger.L.Fatal("unable to connect to database", zap.String("DB_URI", dsn))

		}
		logger.L.Fatal("unable to set up database connection", zap.Error(err))
	}

	app.DB.Pool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		logger.L.Fatal("unable to set create DB connection pool", zap.Error(err))
	}

	err = app.DB.Pool.Ping(ctx)
	if err != nil {
		logger.L.Fatal("unable to ping DB", zap.Error(err))
	}
	logger.L.Info("Connected to database", zap.String("DATABASE_URI", app.Config.DB.URI))

}

func (app *Application) MigrateDatabase(ctx context.Context) {
	m, err := migrate.New("file://migrations", app.Config.DB.URI)
	if err != nil {
		logger.L.Fatal(err.Error())
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.L.Debug("database migrations is up to date")
			return
		}
		logger.L.Fatal(err.Error())
	}
	v, _, err := m.Version()
	if err == nil {
		logger.L.Debug("database has been migrated", zap.Uint("version", v))
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
