package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const webPort = ":8080" // default is 80

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	// connect to rabbit mq
	connection, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer connection.Close()

	app := Config{
		Rabbit: connection,
	}

	log.Printf("Starting broker service on port %s", webPort)
	srv := http.Server{Addr: webPort, Handler: app.routes()}
	err = srv.ListenAndServe()
	if err != nil {
		log.Panicln(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			log.Println("RabbitMQ not available yet...")
			counts++
		} else {
			connection = c
			log.Println("connected to rabbit mq")
			break
		}

		if counts > 5 {
			log.Println(err)
			return nil, err
		}

		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off ...")
		time.Sleep(backOff)
	}
	return connection, nil
}
