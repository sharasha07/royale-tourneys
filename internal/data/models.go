package data

import "github.com/jackc/pgx/v5/pgxpool"

type DBModel struct {
	pool *pgxpool.Pool
}

func NewDBModel(pool *pgxpool.Pool) DBModel {
	return DBModel{
		pool: pool,
	}
}
