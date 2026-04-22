package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	wbfrabbitmq "github.com/wb-go/wbf/rabbitmq"
)

const (
	dlxExchangeArg   = "x-dead-letter-exchange"
	dlxRoutingKeyArg = "x-dead-letter-routing-key"
)

type retryQueueConfig struct {
	name string
	ttl  int32
}

func DeclareTopology(client *wbfrabbitmq.RabbitClient) error {
	err := client.DeclareExchange(
		mainExchange,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("declare main exchange: %w", err)
	}

	err = client.DeclareQueue(
		mainQueue,
		mainExchange,
		mainQueue,
		true,
		false,
		false,
		amqp.Table{
			dlxExchangeArg:   "",
			dlxRoutingKeyArg: retryQueue5s,
		},
	)

	if err != nil {
		return fmt.Errorf("declare main queue: %w", err)
	}

	ch, err := client.GetChannel()
	if err != nil {
		return fmt.Errorf("get channel for retry queues: %w", err)
	}
	defer func() {
		_ = ch.Close()
	}()

	retryQueues := []retryQueueConfig{
		{retryQueue5s, 5000},
		{retryQueue30s, 30000},
		{retryQueue2m, 120000},
	}

	for _, rq := range retryQueues {
		_, err = ch.QueueDeclare(
			rq.name,
			true,
			false,
			false,
			false,
			amqp.Table{
				amqp.QueueMessageTTLArg: rq.ttl,
				dlxExchangeArg:          mainExchange,
				dlxRoutingKeyArg:        mainQueue,
			},
		)
		if err != nil {
			return fmt.Errorf("declare retry queue %s: %w", rq.name, err)
		}
	}

	return nil
}
