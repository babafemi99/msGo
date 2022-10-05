package main

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"listener/event"
	"log"
	"math"
	"os"
	"time"
)

func main() {
	// connect to rabbit mq
	connection, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer connection.Close()
	// listen to queue
	log.Println("Listening and consuming rabbitMQ messages...")
	//create consumer
	consumer, err := event.NewConsumer(connection)
	if err != nil {
		log.Panicln("error:", err)
	}
	//watch the queue and consume events
	err = consumer.Listen([]string{"log.INFO", "log.ERROR", "log.WARNING"})
	if err != nil {
		log.Println("error:", err)
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
