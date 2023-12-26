package rabbitmq

import (
	"context"
	"github.com/Verce11o/yata-tweets/config"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"time"
)

type TweetPublisher struct {
	AmqpConn *amqp.Connection
	log      *zap.SugaredLogger
	trace    trace.Tracer
	cfg      config.RabbitMQ
}

func NewTweetPublisher(amqpConn *amqp.Connection, log *zap.SugaredLogger, trace trace.Tracer, cfg config.RabbitMQ) *TweetPublisher {
	return &TweetPublisher{AmqpConn: amqpConn, log: log, trace: trace, cfg: cfg}
}

func (c *TweetPublisher) createChannel(exchangeName string, queueName string, bindingKey string) *amqp.Channel {

	ch, err := c.AmqpConn.Channel()

	if err != nil {
		panic(err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	queue, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		false,
		nil,
	)

	if err != nil {
		panic(err)
	}

	return ch

}

func (c *TweetPublisher) Publish(ctx context.Context, message []byte) error {

	ch := c.createChannel(c.cfg.ExchangeName, c.cfg.QueueName, c.cfg.BindingKey)

	if err := ch.PublishWithContext(
		ctx,
		c.cfg.ExchangeName,
		c.cfg.BindingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "text/plain",
			DeliveryMode: amqp.Persistent,
			MessageId:    uuid.New().String(),
			Timestamp:    time.Now(),
			Body:         message,
		}); err != nil {

		return err
	}

	return nil

}
