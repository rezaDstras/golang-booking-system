package dbrepo

import (
	"database/sql"

	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/repository"
)

type postgressDBRepo struct {
	App *config.AppConfig
	//postgres db connection
	DB *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &postgressDBRepo{
		App: a,
		DB:  conn,
	}

}
