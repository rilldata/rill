package activity

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// KafkaSink sinks events to a Kafka cluster.
type KafkaSink struct {
	producer *kafka.Producer
	topic    string
}

func NewKafkaSink(brokers, topic string) (*KafkaSink, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}

	return &KafkaSink{
		producer: producer,
		topic:    topic,
	}, nil
}

// Sink doesn't wait till all events are delivered to Kafka
func (s *KafkaSink) Sink(_ context.Context, events []Event) error {
	for _, event := range events {
		message, err := event.Marshal()
		if err != nil {
			return err
		}

		err = s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
			Value:          message,
		}, nil)

		if err != nil {
			return err
		}
	}
	return nil
}

func (s *KafkaSink) Close() error {
	s.producer.Flush(100)
	s.producer.Close()
	return nil
}
