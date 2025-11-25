package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	ampq "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJson[T any](
	conn *ampq.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			var message T
			err := json.Unmarshal(d.Body, &message)
			if err != nil {
				fmt.Printf("Error unmarshaling message:%v", err)
				continue
			}
			ack := handler(message)
			switch ack {
			case Ack:
				d.Ack(false)
				fmt.Print("acknowledged")
			case NackRequeue:
				d.Nack(false, true)
				fmt.Print("nack req")
			case NackDiscard:
				d.Nack(false, false)
				fmt.Print("nack discard")
			}
		}
	}()
	return nil
}

func SubscribeGob[T any](
	conn *ampq.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) Acktype,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveries, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			var message T
			decoder := gob.NewDecoder(bytes.NewBuffer(d.Body))
			err := decoder.Decode(&message)
			if err != nil {
				fmt.Printf("Error unmarshaling message:%v", err)
				continue
			}
			ack := handler(message)
			switch ack {
			case Ack:
				d.Ack(false)
				fmt.Print("acknowledged")
			case NackRequeue:
				d.Nack(false, true)
				fmt.Print("nack req")
			case NackDiscard:
				d.Nack(false, false)
				fmt.Print("nack discard")
			}
		}
	}()
	return nil
}
