package app

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/database"
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
	dsn := app.Config.DB.String()
	_, err := pgx.Connect(ctx, dsn)
	if err != nil {
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
