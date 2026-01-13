package kafka

import (
	"fmt"
	consumer2 "imageProcessor/internal/kafka/consumer"
	"imageProcessor/internal/storage/sqlite"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
)

func Consumer(log *slog.Logger, brokers []string, topic string, doneChannel <-chan struct{}, storage *sqlite.StorageSqlite) error {
	const op = "kafka.NewConsumer"
	// validate fetched brokers
	if len(brokers) == 0 {
		return fmt.Errorf("%s,%s", op, "brokers list is empty")
	}

	// init consumer
	config := sarama.NewConfig()
	config.Version = sarama.V3_9_1_0
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return fmt.Errorf("NewConsumer error; %s, %w", op, err)
	}
	defer consumer.Close()

	// get topics
	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("get broker partitions error; %s, %w", op, err)
	}

	var wg sync.WaitGroup

	for _, partition := range partitions {
		wg.Add(1)
		go func(p int32) {
			defer wg.Done()

			partitionConsumer, err := consumer.ConsumePartition(topic, p, sarama.OffsetNewest) // to consumer all tasks w/o misses
			if err != nil {
				log.Error("create partition consumer error", "op", op, "err", err)
			}
			defer partitionConsumer.Close()

			for {
				select {
				case _, ok := <-doneChannel:
					if !ok {
						log.Info(fmt.Sprintf("Partition %d is closed down", p), "op", op)
						return
					}
				case msg := <-partitionConsumer.Messages():
					log.Info("Get message", "partition", p)
					time.Sleep(15 * time.Second)
					err := consumer2.ConsumedHandler(msg.Value, storage, log)
					if err != nil {
						log.Error("Consumer handler failed;", "err", err)
					}
				case err := <-partitionConsumer.Errors():
					log.Error("getting message from partition error", "op", op, "err", err)
				}
			}
		}(partition)
	}

	wg.Wait()
	log.Info("All consumers stopped gracefully")
	return nil
}
