package database

import (
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDB = errors.New("internal database error")
var ErrConnectDB = errors.New("unable to connect to database")

type DB struct {
	Pool *pgxpool.Pool
}
