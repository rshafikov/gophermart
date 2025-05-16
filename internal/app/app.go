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
