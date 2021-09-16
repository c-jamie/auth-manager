package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/api"
	_ "github.com/lib/pq"
)
// New creates a database connection pool
func New(cfg api.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.ConnStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}