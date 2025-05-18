package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshafikov/gophermart/internal/database/queries"
	"github.com/rshafikov/gophermart/internal/models"
	"log"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{Pool: pool}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	exec, err := r.Pool.Exec(ctx, queries.CreateUser, user.Login, user.Password)
	if err != nil {
		log.Println("unable to CREATE user:", err)
		return err
	}
	log.Println("rows affected: ", exec.RowsAffected())
	return nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User

	q := r.Pool.QueryRow(ctx, queries.GetUserByLogin, login)
	err := q.Scan(&user.ID, &user.Login, &user.Password, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("there is no user with login '%s'", login)
			return nil, err
		}
		log.Println("unable to GET user, unknown error:", err)
		return nil, err
	}
	return &user, nil
}
