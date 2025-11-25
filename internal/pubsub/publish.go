package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonBytes})
	return nil
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var gobBuffer bytes.Buffer
	encoder := gob.NewEncoder(&gobBuffer)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        gobBuffer.Bytes()})
	return nil
}

func PublishGameLog(ch *amqp.Channel, logMsg string, username string) error {
	log := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     logMsg,
		Username:    username,
	}
	err := PublishGob(ch, string(routing.ExchangePerilTopic), string(routing.GameLogSlug)+"."+username, log)
	if err != nil {
		return err
	}
	return nil
}
