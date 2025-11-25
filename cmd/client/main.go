package main

import (
	"fmt"
	"strconv"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connectionString := "amqp://guest:guest@localhost:5672/"
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer connection.Close()
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Print(err)
		return
	}
	_, _, err = pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient)
	if err != nil {
		fmt.Print(err)
		return
	}
	gamestate := gamelogic.NewGameState(username)
	err = pubsub.SubscribeJson(
		connection,
		routing.ExchangePerilDirect,
		fmt.Sprintf("pause.%s",
			username),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gamestate),
	)
	if err != nil {
		fmt.Println(err)
	}

	warChan, _, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.SimpleQueueDurable,
	)
	if err != nil {
		fmt.Println("failed to bind to war:", err)
	}

	moveChan, _, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		fmt.Println("failed to bind to move:", err)
	}

	err = pubsub.SubscribeJson(
		connection,
		routing.ExchangePerilTopic,
		fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username),
		fmt.Sprintf("%s.*", routing.ArmyMovesPrefix),
		pubsub.SimpleQueueTransient,
		handlerMove(gamestate, warChan),
	)
	if err != nil {
		fmt.Println("failed to sub to move:", err)
	}

	logChan, _, err := pubsub.DeclareAndBind(
		connection,
		routing.ExchangePerilTopic,
		"game_logs",
		fmt.Sprintf("%s.%s", routing.GameLogSlug, username),
		pubsub.SimpleQueueDurable,
	)

	err = pubsub.SubscribeJson(
		connection,
		routing.ExchangePerilTopic,
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.SimpleQueueDurable,
		handlerWar(gamestate, logChan),
	)
	if err != nil {
		fmt.Println("failed to sub to move:", err)
	}

Repl:
	for true {
		cmd := gamelogic.GetInput()
		switch cmd[0] {
		case "spawn":
			err = gamestate.CommandSpawn(cmd)
			fmt.Print(cmd[1:])
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			move, err := gamestate.CommandMove(cmd)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(moveChan, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username), move)
			if err != nil {
				fmt.Println("failed to publish:", err)
				continue
			}
			fmt.Println("Published a move!")

		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			if len(cmd) < 2 {
				fmt.Println("Spam needs 1 argument")
				continue
			}
			n, err := strconv.Atoi(cmd[1])
			if err != nil {
				fmt.Println("Wrong arg:", err)
				continue
			}
			for range n {
				maliciosLog := gamelogic.GetMaliciousLog()
				err = pubsub.PublishGameLog(
					logChan,
					maliciosLog,
					username,
				)
				if err != nil {
					fmt.Println("cant publish malicios log")
					break
				}
			}
		case "quit":
			break Repl
		default:
			fmt.Println("Command not recognised")
		}
	}
	fmt.Print("Client closed")
}
