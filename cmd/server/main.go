package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer connection.Close()
	fmt.Printf("Connection success!")
	newChan, err := connection.Channel()
	if err != nil {
		fmt.Print(err)
		return
	}
	gamelogic.PrintServerHelp()

	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.SimpleQueueDurable)
	if err != nil {
		fmt.Print(err)
		return
	}

	pubsub.SubscribeGob(
		connection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		"game_logs.*",
		pubsub.SimpleQueueDurable,
		handlerLogs,
	)

Repl:
	for true {
		cmd := gamelogic.GetInput()
		if len(cmd) < 1 {
			fmt.Println()
			continue
		}
		switch cmd[0] {
		case "pause":
			pubsub.PublishJSON(newChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			pubsub.PublishJSON(newChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			break Repl
		default:
			fmt.Println("No command like that")
		}
	}
	fmt.Print("Programm closing!")
}
