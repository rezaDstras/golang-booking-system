package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

//DB holds the database connection poll
type DB struct {
	SQL sql.DB
}

var dbConn = &DB{}

const maxOpenDbConn = 10
const maxIdleDbConn = 5
const maxLifeTimeDbConn = 5 * time.Minute

//ConnectSQL Create database pool for Postgres
func ConnectSQL(dsn string) (*DB, error) {
	d, err := NewDatabase(dsn)
	if err != nil {
		panic(err)
	}

	d.SetMaxOpenConns(maxOpenDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxLifetime(maxLifeTimeDbConn)

	dbConn.SQL = *d

	err = testDB(d)
	if err != nil {
		return nil, err
	}

	return dbConn, nil

}

//testDB test the database connection
func testDB(d *sql.DB) error {
	err := d.Ping()
	if err != nil {
		return err
	}
	return nil
}

//NewDatabase create a new database connection
func NewDatabase(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	//test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil

}
