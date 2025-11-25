package pubsub

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	clientChan, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	args := make(amqp.Table)
	args["x-dead-letter-exchange"] = "peril_dlx"
	queue, err := clientChan.QueueDeclare(
		queueName,
		queueType == SimpleQueueDurable,
		queueType == SimpleQueueTransient,
		queueType == SimpleQueueTransient,
		false,
		args,
	)
	clientChan.Qos(10, 0, false)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	err = clientChan.QueueBind(queue.Name,
		key,
		exchange,
		false,
		nil)

	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return clientChan, queue, nil

}
