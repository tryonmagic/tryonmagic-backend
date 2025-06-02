package rabbitmq

import (
	"log"
	"sync"

	"github.com/mbbipul/tryonmagic-backend/src/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConnection struct {
	url        string
	connection *amqp.Connection
	channel    *amqp.Channel
	mu         sync.Mutex
}

func NewRabbitMQConnection(url string) (*RabbitMQConnection, error) {
	conn := &RabbitMQConnection{
		url: url,
	}
	err := conn.connect()
	return conn, err
}

func (r *RabbitMQConnection) connect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var err error
	r.connection, err = amqp.Dial(r.url)
	if err != nil {
		log.Printf("Failed to connect to RabbitMQ %v", err)
		return err
	}
	r.channel, err = r.connection.Channel()
	if err != nil {
		log.Printf("Failed to open a channel %v", err)
	}
	return nil
}

func (r *RabbitMQConnection) CheckConnection() {
	if r.connection == nil || r.connection.IsClosed() {
		log.Println("RabbitMQ connection is closed, reconnecting...")
		r.connect()
	}
}

func (r *RabbitMQConnection) Close() {
	r.channel.Close()
	r.connection.Close()
}

var (
	instance *RabbitMQConnection
	once     sync.Once
)

// GetConnection returns the singleton instance of the RabbitMQ connection
func GetConnection() (*RabbitMQConnection, error) {
	var err error
	once.Do(func() {
		instance, err = NewRabbitMQConnection(config.RabbitMQConnectionUrl)
	})
	return instance, err
}
