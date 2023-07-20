package publisher

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
)

// KafkaSink sinks events to a Kafka cluster.
type KafkaSink struct {
	producer *kafka.Producer
	topic    string
	logger   *zap.Logger
}

func NewKafkaSink(brokers, topic string, logger *zap.Logger) (*KafkaSink, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}

	return &KafkaSink{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

// Sink doesn't wait till all events are delivered to Kafka
func (s *KafkaSink) Sink(events []Event) {
	for _, event := range events {
		message, err := event.Marshal()
		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not serialize event: %v", event), zap.Error(err))
			continue
		}

		err = s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
			Value:          message,
		}, nil)

		if err != nil {
			s.logger.Debug(fmt.Sprintf("could not produce event to Kafka: %v", event), zap.Error(err))
		}
	}
}

func (s *KafkaSink) Close() error {
	s.producer.Flush(100)
	s.producer.Close()
	return nil
}
