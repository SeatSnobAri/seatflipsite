package dbrepo

import (
	"database/sql"
	"github.com/SeatSnobAri/seatflip2.0/internal/config"
	"github.com/SeatSnobAri/seatflip2.0/internal/repository"
)

var app *config.AppConfig

type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}
type testDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo creates the repository
func NewPostgresRepo(Conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	app = a
	return &postgresDBRepo{
		App: a,
		DB:  Conn,
	}
}

// NewTestingRepo creates a repo with a dummy database for testing
//func NewTestingRepo(a *config.AppConfig) repository.DatabaseRepo {
//	return &testDBRepo{
//		App: a,
//	}
//}
