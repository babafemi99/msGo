package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}
	err := consumer.setUp()
	if err != nil {
		return nil, err
	}
	return &Consumer{conn: conn}, nil
}

func (c *Consumer) setUp() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Consumer) Listen(topics []string) error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	q, err := declareRandomQueue(channel)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err := channel.QueueBind(q.Name, s, "logs_topic", false, nil)
		if err != nil {
			return err
		}
	}
	messages, err := channel.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}
	forever := make(chan bool)
	go func() {
		for d := range messages {
			var payload Payload
			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)
		}
	}()
	fmt.Printf("Waiting for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "event":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "auth":
	// authentication logic
	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}

}

func logEvent(entry Payload) error {
	jsonData, je := json.MarshalIndent(entry, "", "\t")
	if je != nil {
		return je
	}
	logSrvUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logSrvUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}

	res, err := client.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}
