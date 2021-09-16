package api

import (
	"database/sql"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/migrations"
)
// Application represents our Application model
type Application struct {
	Config     *Config
	Models     data.Models
	Middleware Middleware
	Migrations migrations.Migrations
}
// Config represents our Application configuration
type Config struct {
	Port         int
	Env          string
	Version      string
	BuildVersion string
	APIVerion    string
	GinMode      string
	DB           struct {
		ConnStr      string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdelTime  int
	}
}
// NewApplication creates a new Application
func NewApplication(db *sql.DB, cfg *Config) (*Application, error) {
	app := Application{
		Config:     cfg,
		Models:     data.NewModels(db),
		Middleware: NewMiddleware("/", db),
		Migrations: migrations.Migrations{DB: db},
	}

	return &app, nil
}
