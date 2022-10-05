package events

import (
	"context"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Emitter struct {
	connection *amqp.Connection
}

func NewEventEmitter(connection *amqp.Connection) (*Emitter, error) {
	emitter := &Emitter{connection: connection}
	err := emitter.setup()
	if err != nil {
		return nil, nil
	}
	return emitter, nil
}

func (e *Emitter) setup() error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	return declareExchange(channel)
}

func (e *Emitter) Push(event, severity string) error {
	channel, err := e.connection.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	log.Println("PUSHING TO CHANNEL")
	err = channel.PublishWithContext(context.Background(), "logs_topic", severity, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(event),
	})
	if err != nil {
		return err
	}
	return nil
}
