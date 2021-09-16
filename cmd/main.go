package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/c-jamie/sql-manager-acc-auth/serverlib/api"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/db"
	"github.com/c-jamie/sql-manager-acc-auth/serverlib/log"
	"github.com/gin-gonic/gin"
)

var (
	apiVersion   string
	buildVersion string
)

func main() {
	mode, ok := os.LookupEnv("SQM_SER_MODE")
	if !ok {
		log.Fatal("unable to load SQM_SER_MODE")
	}
	if mode == "debug" {
		log.New("debug")
	} else {
		log.New("info")
	}
	dbHost, ok := os.LookupEnv("SQM_SER_DB_HOST")
	if !ok {
		log.Fatal("unable to load DB_HOST")
	}
	addr, ok := os.LookupEnv("SQM_SER_HTTP_PORT")
	if !ok {
		log.Fatal("unable to load SQM_SER_GIN_PORT")
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
	log.Info("db conn str is ", dbConnStr)

	cfg := api.Config{}

	if mode == "debug" {
		cfg.GinMode = gin.DebugMode
	} else {
		cfg.GinMode = gin.ReleaseMode
	}

	cfg.Version = "v1"
	cfg.APIVerion = apiVersion
	cfg.BuildVersion = buildVersion
	cfg.DB.ConnStr = dbConnStr
	cfg.DB.MaxIdelTime = 5
	cfg.DB.MaxOpenConns = 3

	db, err := db.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	app, err := api.NewApplication(db, &cfg)
	if err != nil {
		log.Fatal(err)
	}
	app.Migrations.DoMigrations("up")
	address := ":" + addr
	log.Info("listening on address: ", address)
	service := &http.Server{
		Addr: address, 
		Handler: app.Routes(), 
		ReadTimeout: 8 * time.Second, 
		WriteTimeout: 8 * time.Second,
	}
	log.Fatal(service.ListenAndServe())
}
