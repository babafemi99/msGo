package main

import (
	data "auth/models"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"log"
	"net/http"
	"os"
	"time"
	// imports
)

var count int64

type Config struct {
	DB    *sql.DB
	Model data.Models
}

func main() {
	log.Println("starting auth service")
	// connect to DB

	conn := connectToDB()
	if conn == nil {
		log.Panic("Cannot connect to postgres")
	}

	// set up configs
	app := Config{
		DB:    conn,
		Model: data.New(conn),
	}

	srv := &http.Server{
		Addr:    ":80",
		Handler: app.Routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	fmt.Println("DSN is ", dsn)
	for {
		conn, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres service not yet available")
			count++
		} else {
			log.Println("Connected to postgres service")
			return conn
		}

		if count > 10 {
			log.Println(err)
			return nil
		}
		log.Printf("retrying after 3 seconds ... %v out of 10", count)
		time.Sleep(3 * time.Second)
		continue
	}
}
