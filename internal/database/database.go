package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/core/logger"
	"go.uber.org/zap"
)

var ErrDB = errors.New("internal database error")
var ErrConnectDB = errors.New("unable to connect to database")

type DB struct {
	Pool *pgxpool.Pool
}

func (db *DB) ConnectToDatabase(ctx context.Context, databaseURI string) error {
	_, err := pgx.Connect(ctx, databaseURI)
	if err != nil {
		var pgErr *pgconn.ConnectError
		if errors.As(err, &pgErr) {
			logger.L.Error("unable to connect to database", zap.String("DATABASE_URI", databaseURI))
			return ErrConnectDB
		}
		logger.L.Error("unable to set up database connection", zap.Error(err))
		return ErrDB
	}

	db.Pool, err = pgxpool.New(ctx, databaseURI)
	if err != nil {
		logger.L.Error("unable to set create DB connection pool", zap.Error(err))
		return ErrConnectDB
	}

	err = db.Pool.Ping(ctx)
	if err != nil {
		logger.L.Error("unable to ping DB", zap.Error(err))
		return ErrConnectDB
	}
	logger.L.Debug("Connected to database", zap.String("DATABASE_URI", databaseURI))
	return nil
}
