package activity

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// KafkaSink sinks events to a Kafka cluster.
type KafkaSink struct {
	producer *kafka.Producer
	topic    string
	logger   zap.Logger
	logChan  chan kafka.LogEvent
}

func NewKafkaSink(brokers, topic string, logger zap.Logger) (*KafkaSink, error) {
	logChan := make(chan kafka.LogEvent, 100)
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      brokers,
		"go.logs.channel.enable": true,
		"go.logs.channel":        logChan,
	})
	if err != nil {
		return nil, err
	}

	go forwardKafkaLogEventToLogger(logChan, logger)

	return &KafkaSink{
		producer: producer,
		topic:    topic,
		logger:   logger,
		logChan:  logChan,
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
	close(s.logChan)
	return nil
}

func forwardKafkaLogEventToLogger(logChan chan kafka.LogEvent, logger zap.Logger) {
	for logEvent := range logChan {
		zapLevel := kafkaLogLevelToZapLevel(logEvent.Level)
		if logger.Core().Enabled(zapLevel) {
			fields := []zapcore.Field{
				zap.String("kafka.producer.client.name", logEvent.Name),
				zap.String("kafka.producer.tag", logEvent.Tag),
			}
			logger.Log(zapLevel, logEvent.Message, fields...)
		}
	}
}

//Level	Description
//OFF	Turns off logging.
//FATAL	Severe errors that cause premature termination.
//ERROR	Other runtime errors or unexpected conditions.
//WARN	Runtime situations that are undesirable or unexpected, but not necessarily wrong.
//INFO	Runtime events of interest at startup and shutdown.
//DEBUG	Detailed diagnostic information about events.
//TRACE	Detailed diagnostic information about everything.
func kafkaLogLevelToZapLevel(level int) zapcore.Level {
	switch level {
	case 0, 1, 2:
		return zap.FatalLevel
	case 3:
		return zap.ErrorLevel
	case 4:
		return zap.WarnLevel
	case 5, 6:
		return zap.InfoLevel
	case 7:
		return zap.DebugLevel
	default:
		return zap.DebugLevel // Default to debug for unrecognized levels
	}
}
