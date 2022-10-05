package event

import amqp "github.com/rabbitmq/amqp091-go"

func declareExchange(ch *amqp.Channel) error {
	err := ch.ExchangeDeclare("logs_topic", "topic", true, false, false, false, nil)
	if err != nil {
		return err
	}
	return nil
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
}
