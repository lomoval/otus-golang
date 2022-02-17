package rabbit

import (
	"context"
	"fmt"
	"time"

	"github.com/streadway/amqp"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Queue    string
}

type Message struct {
	ID      string
	Name    string
	Time    time.Time
	OwnerID string
}

type Provider struct {
	conn       *amqp.Connection
	queue      amqp.Queue
	channel    *amqp.Channel
	connString string
	queueName  string
}

func New(config Config) *Provider {
	return &Provider{
		connString: fmt.Sprintf(
			"amqp://%s:%s@%s:%d/",
			config.User,
			config.Password,
			config.Host,
			config.Port,
		),
		queueName: config.Queue,
	}
}

func (r *Provider) Connect() error {
	var err error
	r.conn, err = amqp.Dial(r.connString)
	if err != nil {
		return err
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return err
	}
	r.queue, err = r.channel.QueueDeclare(
		r.queueName,
		false,
		true,
		false,
		false,
		nil,
	)
	return err
}

func (r *Provider) Close() {
	r.conn.Close()
}

func (r *Provider) Publish(body []byte) error {
	return r.channel.Publish(
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
}

type MessageProcess = func(msg amqp.Delivery)

func (r Provider) Consume(ctx context.Context, process MessageProcess) error {
	msgs, err := r.channel.Consume(
		r.queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case m, ok := <-msgs:
			if ok {
				process(m)
			}
		}
	}
}
