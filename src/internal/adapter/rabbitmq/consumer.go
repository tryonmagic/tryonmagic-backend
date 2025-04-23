package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn *RabbitMQConnection
}

func NewConsumer(conn *RabbitMQConnection) *Consumer {
	return &Consumer{conn: conn}
}

func (c *Consumer) DeclareQueue(queueName string) error {
	c.conn.CheckConnection()
	_, err := c.conn.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (c *Consumer) DeclareExchange(exchangeName, exchangeType string) error {
	c.conn.CheckConnection()
	return c.conn.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (c *Consumer) BindQueue(exchangeName, queueName, routingKey string) error {
	c.conn.CheckConnection()
	return c.conn.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}

func (c *Consumer) Consume(queueName string, handler func(ctx context.Context, msg amqp091.Delivery)) error {
	c.conn.CheckConnection()

	msgs, err := c.conn.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			handler(context.Background(), d)
		}
	}()

	return nil
}

func DecodeMessage[T any](body []byte) (T, error) {
	var decoded T
	err := json.Unmarshal(body, &decoded)
	return decoded, err
}
