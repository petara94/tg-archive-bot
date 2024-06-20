package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Dial(url string) (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), url)
}
