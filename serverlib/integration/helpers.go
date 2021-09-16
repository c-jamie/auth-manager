
package main

import (
	"bytes"
	"fmt"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/api"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/db"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/data"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/migrations"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/internal/mocks"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/log"
	"net/http"
	"net/http/httptest"
	"os"
)


func DoRequest(app *api.Application, json []byte, url string, token string, method string) (*bytes.Buffer, int) {
	testRouter := app.Routes()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(json))
	if token != ""{
		req.Header.Set("Authorization", "Bearer "+token)
	}
	testRouter.ServeHTTP(w, req)
	return w.Body, w.Code
}

func setup(mockAuth bool) *api.Application {
	log.New("debug")
	dbHost, ok := os.LookupEnv("SQM_SER_DB_HOST")
	if !ok {
		log.Fatal("unable to load DB_HOST")
	}
	dbPort, ok := os.LookupEnv("SQM_SER_DB_PORT")
	if !ok {
		log.Fatal("unable to load DB_HOST")
	}
	dbName, ok := os.LookupEnv("SQM_SER_DB_NAME")
	if !ok {
		log.Fatal("unable to load DB_NAME")
	}
	dbUser, ok := os.LookupEnv("SQM_SER_DB_USER")
	if !ok {
		log.Fatal("unable to load DB_USER")
	}
	dbPW, ok := os.LookupEnv("SQM_SER_DB_PW")
	if !ok {
		log.Fatal("unable to load DB_PW")
	}
	if !ok {
		log.Fatal("unable to load DB_PW")
	}
	ssl := "disable"
	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", dbUser, dbPW, dbHost, dbPort, dbName, ssl)
	log.Info("user ", dbUser)
	log.Info("pw ", dbPW)
	log.Info("db conn str is ", dbConnStr)

	cfg := api.Config{}
	cfg.Version = "v1"
	cfg.DB.ConnStr = dbConnStr
	cfg.DB.MaxIdelTime = 5
	cfg.DB.MaxOpenConns = 3

	db, err := db.New(cfg)

	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}
	var middleware api.Middleware
	if mockAuth {
		middleware = &mocks.MockMiddleware{}
	} else {
		middleware = api.NewMiddleware("/", db)
	}

	app := api.Application{
		Config:     &cfg,
		Models:     data.NewModels(db),
		Middleware: middleware,
		Migrations: migrations.Migrations{DB: db},
	}
	app.Migrations.DoMigrations("up")
	app.Migrations.DoMigrations("down")
	app.Migrations.DoMigrations("up")

	return &app
}
