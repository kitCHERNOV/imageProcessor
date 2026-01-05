package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// internal/kafka/consumer.go

type Consumer interface {
	Consume(topic string, handler func([]byte)) error
	Close() error
}

type KafkaConsumer struct {
	consumer sarama.ConsumerGroup
	handler  func([]byte)
}

type consumerGroupHandler struct {
	handler func([]byte)
}

func NewConsumer(brokers []string, groupID string) (Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &KafkaConsumer{consumer: consumer}, nil
}

func (c *KafkaConsumer) Consume(topic string, handler func([]byte)) error {
	c.handler = handler

	ctx := context.Background()
	topics := []string{topic}
	cgHandler := &consumerGroupHandler{handler: handler}

	for {
		if err := c.consumer.Consume(ctx, topics, cgHandler); err != nil {
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return c.consumer.Close()
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.handler(msg.Value)
		session.MarkMessage(msg, "")
	}
	return nil
}

// ConsumerMessage represents a consumed Kafka message
type ConsumerMessage struct {
	Topic string
	Data  []byte
}

// ConsumeMessages is a simplified consumer that returns a channel of messages
func (c *KafkaConsumer) ConsumeMessages(topic string) (<-chan ConsumerMessage, error) {
	msgChan := make(chan ConsumerMessage, 100)

	go func() {
		defer close(msgChan)
		handler := func(data []byte) {
			msgChan <- ConsumerMessage{
				Topic: topic,
				Data:  data,
			}
		}
		c.handler = handler

		ctx := context.Background()
		topics := []string{topic}
		cgHandler := &consumerGroupHandler{handler: handler}

		for {
			if err := c.consumer.Consume(ctx, topics, cgHandler); err != nil {
				log.Printf("Error consuming: %v", err)
				return
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	return msgChan, nil
}

// UnmarshalMessage unmarshals a JSON message into the provided struct
func UnmarshalMessage(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
