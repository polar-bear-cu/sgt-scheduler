package publisher

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Same as Queue Name in noti consumer
const QueueName = "email_notifications"

type emailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type RabbitMQPublisher struct {
	ch *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection) (*RabbitMQPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{ch: ch}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, to, subject, body string) error {
	payload, err := json.Marshal(emailMessage{To: to, Subject: subject, Body: body})
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, "", QueueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        payload,
	})
}
