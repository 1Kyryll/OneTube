package queue

import (
	"context"
	"encoding/json"
	"fmt"

	ampq "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn  *ampq.Connection
	ch    *ampq.Channel
	queue string
}

func NewPublisher(url, queue string) (*Publisher, error) {
	conn, err := ampq.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to open channel: %w", err)
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("Failed to declare queue: %w", err)
	}

	return &Publisher{conn: conn, ch: ch, queue: queue}, nil
}

func (p *Publisher) PublishTranscode(ctx context.Context, job TranscodeJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx, "", p.queue, false, false, ampq.Publishing{
		ContentType:  "application/json",
		DeliveryMode: ampq.Persistent,
		Body:         body,
	})
}

func (p *Publisher) Close() {
	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
}

type Handler func(ctx context.Context, job TranscodeJob) error

type Consumer struct {
	conn  *ampq.Connection
	ch    *ampq.Channel
	queue string
}

func NewConsumer(url, queue string) (*Consumer, error) {
	conn, err := ampq.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("Failed to open channel: %w", err)
	}

	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("Failed to declare queue: %w", err)
	}

	if err := ch.Qos(1, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("Failed to set QoS: %w", err)
	}

	return &Consumer{conn: conn, ch: ch, queue: queue}, nil
}

func (c *Consumer) Run(ctx context.Context, h Handler) error {
	deliveries, err := c.ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("Consumer channel closed")
			}
			var job TranscodeJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				d.Nack(false, false)
				continue
			}
			if err := h(ctx, job); err != nil {
				d.Nack(false, false)
				continue
			}
			d.Ack(false)
		}
	}
}

func (c *Consumer) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
