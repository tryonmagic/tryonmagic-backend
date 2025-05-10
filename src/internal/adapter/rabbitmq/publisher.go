package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mbbipul/prodigyai-backend/src/config"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *RabbitMQConnection
}

func NewPublisher(conn *RabbitMQConnection) *Publisher {
	return &Publisher{conn: conn}
}

func (p *Publisher) DeclareQueue(queueName string) error {
	p.conn.CheckConnection()
	_, err := p.conn.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (p *Publisher) DeclareExchange(exchangeName, exchangeType string) error {
	p.conn.CheckConnection()
	return p.conn.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (p *Publisher) BindQueue(exchangeName, queueName, routingKey string) error {
	p.conn.CheckConnection()
	return p.conn.channel.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
}

type Headers struct {
	JobID    string `json:"jobId"`
	Priority int    `json:"priority"`
	TaskType string `json:"taskType,omitempty"`
}

func (p *Publisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}, headers Headers) error {
	p.conn.CheckConnection()

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	log.Printf("Publishing message to exchange: %s, routing key: %s, body: %s", exchange, routingKey, string(bodyBytes))

	err = p.conn.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Headers: amqp091.Table{
				"job_id":    headers.JobID,
				"priority":  headers.Priority,
				"task_type": headers.TaskType,
			},
			Body: bodyBytes,
		},
	)

	if err != nil {
		log.Println("Failed to publish message:", err)
	}
	return err
}

type BasePublisher interface {
	Publish(ctx context.Context, message any, headers Headers) error
}

type basePublisher struct {
	Publisher    *Publisher
	exchangeName string
	exchangeType string
	queueName    string
	routingKey   string
}

func (bp *basePublisher) Setup() error {
	err := bp.Publisher.DeclareExchange(bp.exchangeName, "direct")
	if err != nil {
		log.Printf("Failed to declare exchange: %v", err)
		return err
	}

	err = bp.Publisher.DeclareQueue(bp.queueName)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return err
	}

	err = bp.Publisher.BindQueue(bp.exchangeName, bp.queueName, bp.routingKey)
	if err != nil {
		log.Printf("Failed to bind queue: %v", err)
		return err
	}
	return nil
}
func (bp *basePublisher) Publish(ctx context.Context, message any, headers Headers) error {
	err := bp.Setup()
	if err != nil {
		log.Printf("Failed to setup publisher: %v", err)
		return err
	}
	return bp.Publisher.Publish(ctx, bp.exchangeName, bp.routingKey, message, headers)
}

func RemoveBackgroundTaskPublisher() (BasePublisher, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Printf("Failed to connect to Message broker: %v", err)
		return nil, err
	}
	publisher := basePublisher{
		Publisher:    NewPublisher(conn),
		exchangeName: config.AITaskExchangeName,
		exchangeType: "direct",
		queueName:    config.RemoveBackgroundTaskQueueName,
		routingKey:   config.RemoveBackgroundTaskRoutingKey,
	}
	return &publisher, nil
}

func ReplaceBackgroundTaskPublisher() (BasePublisher, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Printf("Failed to connect to Message broker: %v", err)
		return nil, err
	}
	publisher := basePublisher{
		Publisher:    NewPublisher(conn),
		exchangeName: config.AITaskExchangeName,
		exchangeType: "direct",
		queueName:    config.ReplaceBackgroundTaskQueueName,
		routingKey:   config.ReplaceBackgroundTaskRoutingKey,
	}
	return &publisher, nil
}

func GenerateAIModelTaskPublisher() (BasePublisher, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Printf("Failed to connect to Message broker: %v", err)
		return nil, err
	}
	publisher := basePublisher{
		Publisher:    NewPublisher(conn),
		exchangeName: config.AITaskExchangeName,
		exchangeType: "direct",
		queueName:    config.GenerateAIModelTaskQueueName,
		routingKey:   config.GenerateAIModelTaskRoutingKey,
	}
	return &publisher, nil
}

func UpscaleTaskPublisher() (BasePublisher, error) {
	conn, err := GetConnection()
	if err != nil {
		log.Printf("Failed to connect to Message broker: %v", err)
		return nil, err
	}
	publisher := basePublisher{
		Publisher:    NewPublisher(conn),
		exchangeName: config.AITaskExchangeName,
		exchangeType: "direct",
		queueName:    config.UpscaleTaskQueueName,
		routingKey:   config.UpscaleTaskRoutingKey,
	}
	return &publisher, nil
}
