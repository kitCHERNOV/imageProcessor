package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

// internal/kafka/producer.go

type Producer interface {
	SendMessage(topic string, message interface{}) error
	Close() error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewProducer(logbrokers []string) (Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: producer}, nil
}

func (p *KafkaProducer) SendMessage(topic string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}

func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}
