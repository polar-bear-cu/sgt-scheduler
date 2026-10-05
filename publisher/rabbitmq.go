package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/polar-bear-cu/sgt-scheduler/usecases"
)

// Same as Queue Name in noti consumer
const QueueName = "email_notifications"

type reminderMessage struct {
	ReminderID     string `json:"reminder_id"`
	UserID         string `json:"user_id"`
	SubscriptionID string `json:"subscription_id"`
	To             string `json:"to"`
	Title          string `json:"title"`
	Content        string `json:"content"`
}

type RabbitMQPublisher struct {
	ch *amqp.Channel
}

func NewRabbitMQPublisher(conn *amqp.Connection) (*RabbitMQPublisher, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare(QueueName, true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &RabbitMQPublisher{ch: ch}, nil
}

func (p *RabbitMQPublisher) Publish(ctx context.Context, job usecases.ReminderJob) error {
	payload, err := json.Marshal(reminderMessage{
		ReminderID:     job.ReminderID,
		UserID:         job.UserID,
		SubscriptionID: job.SubscriptionID,
		To:             job.To,
		Title:          job.Title,
		Content:        job.Content,
	})
	if err != nil {
		return err
	}

	confirm, err := p.ch.PublishWithDeferredConfirmWithContext(ctx, "", QueueName, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    job.ReminderID,
		Timestamp:    time.Now(),
		Body:         payload,
	})
	if err != nil {
		return err
	}

	acked, err := confirm.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !acked {
		return fmt.Errorf("broker nacked message %s", job.ReminderID)
	}
	return nil
}
