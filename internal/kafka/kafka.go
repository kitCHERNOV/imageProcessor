package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
)

type brokerManager struct {
	admin sarama.ClusterAdmin
}

func NewKafkaManager(brokers []string) (*brokerManager, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V3_0_2_0

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("creating kafka manger error")
	}

	return &brokerManager{admin: admin}, nil
}

func (bm *brokerManager) InitTopics(topics map[string]sarama.TopicDetail) error {
	for name, details := range topics {
		err := bm.admin.CreateTopic(name, &details, false)
		if err != nil {
			return fmt.Errorf("creating topic error: %w", err)
		}
	}
	return nil
}

func (bm *brokerManager) Close() error {
	return bm.admin.Close()
}
