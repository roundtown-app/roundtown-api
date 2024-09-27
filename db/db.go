package db

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type DB struct {
    *bun.DB
}

func NewDB(dsn string) (*DB, error) {
    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))

    bunDB := bun.NewDB(sqldb, pgdialect.New())

    // Optional: set database options
    bunDB.SetMaxOpenConns(10)
    bunDB.SetMaxIdleConns(5)
    bunDB.SetConnMaxLifetime(time.Hour)

    // Verify the connection
    if err := bunDB.Ping(); err != nil {
        return nil, err
    }

    return &DB{bunDB}, nil
}