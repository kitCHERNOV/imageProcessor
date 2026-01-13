package kafka

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

// internal/kafka/producer.go

type Producer interface {
	SendMessage(message interface{}) error
	Close() error
}

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
	logger   *slog.Logger
}

func NewProducer(brokers []string, topic string, log *slog.Logger) (Producer, error) {
	log.Info("producer is creating...")
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{producer: producer, topic: topic, logger: log}, nil
}

func (p *KafkaProducer) SendMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil { // if message was not sent
		p.logger.Error("message was not sent", slog.String("topic", p.topic), "error", err)
		return err
	}

	p.logger.Debug("message sent successfully",
		slog.String("topic", p.topic),
		slog.Int64("partition", int64(partition)),
		slog.Int64("offset", offset),
	)

	return nil
}

func (p *KafkaProducer) Close() error {
	p.logger.Info("producer is stopping...")
	return p.producer.Close()
}
