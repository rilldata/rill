package activity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Kafka producer config
const (
	lingerMs         = 200
	compressionCodec = "lz4"
)

// Retry config
const (
	retryN                       = 3
	retryWait                    = lingerMs * time.Millisecond
	metadataTimeout              = 5 * time.Second
	deliveryFailuresReportPeriod = 5 * time.Minute
)

// OTel metrics for Kafka delivery
var (
	meter                  = otel.Meter("github.com/rilldata/rill/runtime/pkg/activity")
	deliverySuccessCounter = must(meter.Int64Counter("kafka_delivery_success"))
	deliveryFailureCounter = must(meter.Int64Counter("kafka_delivery_failure"))
)

type kafkaSink struct {
	producer *kafka.Producer
	topic    string
	logger   *zap.Logger
	logChan  chan kafka.LogEvent
	closedCh chan struct{}
}

// NewKafkaSink returns a sink that sends events to a Kafka topic.
func NewKafkaSink(brokers, topic string, logger *zap.Logger) (Sink, error) {
	logChan := make(chan kafka.LogEvent, 100)
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      brokers,
		"go.logs.channel.enable": true,
		"go.logs.channel":        logChan,
		"linger.ms":              lingerMs,         // Configure waiting time before sending out a batch of messages
		"compression.codec":      compressionCodec, // Specify the compression type to be used for messages
	})
	if err != nil {
		return nil, err
	}

	go forwardKafkaLogEventToLogger(logChan, logger)

	// Check connectivity and fail fast if Kafka cluster is unreachable
	// If the topic doesn't exist, the request doesn't fail but returns no metadata
	// The topic might be auto-created on a first message
	_, err = producer.GetMetadata(&topic, false, int(metadataTimeout.Milliseconds()))
	if err != nil {
		return nil, err
	}

	sink := &kafkaSink{
		producer: producer,
		topic:    topic,
		logger:   logger,
		logChan:  logChan,
		closedCh: make(chan struct{}),
	}

	go sink.processProducerEvents()

	return sink, nil
}

func (s *kafkaSink) Emit(event Event) error {
	message, err := event.Marshal()
	if err != nil {
		return err
	}

	sendMessageFn := func() error {
		return s.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
			Value:          message,
		}, nil)
	}

	retryOnErrFn := func(err error) bool {
		kafkaErr := kafka.Error{}
		// Producer queue is full, wait for messages to be delivered then try again.
		return errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrQueueFull
	}

	err = retry(retryN, retryWait, sendMessageFn, retryOnErrFn)
	if err != nil {
		return err
	}

	return nil
}

func (s *kafkaSink) Close() {
	s.producer.Flush(10000)
	s.producer.Close()
	close(s.closedCh)
	close(s.logChan)
}

func (s *kafkaSink) processProducerEvents() {
	var deliveryFailureCount int
	var lastDeliveryError error

	ticker := time.NewTicker(deliveryFailuresReportPeriod)
	defer ticker.Stop()

	for {
		select {
		case e, ok := <-s.producer.Events():
			if !ok {
				return
			}
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					deliveryFailureCounter.Add(context.Background(), 1,
						metric.WithAttributes(attribute.String("error", ev.TopicPartition.Error.Error())))
					deliveryFailureCount++
					lastDeliveryError = ev.TopicPartition.Error
				} else {
					deliverySuccessCounter.Add(context.Background(), 1)
				}
			case kafka.Error:
				// This error might be a duplicate of what is logged by forwardKafkaLogEventToLogger
				// Use warn level to focus on non-delivered events only as broker disconnects might be false-positive
				s.logger.Warn("Kafka sink: producer error", zap.String("error", ev.Error()))
			default:
				// Ignore any other events
			}

		case <-ticker.C:
			if deliveryFailureCount > 0 {
				s.logger.Error(
					fmt.Sprintf("Kafka sink: delivery failures in the last observed period: %d. "+
						"Check preceding log events to investigate the issue", deliveryFailureCount),
					zap.Error(lastDeliveryError))
				deliveryFailureCount = 0
				lastDeliveryError = nil
			}

		case <-s.closedCh:
			return
		}
	}
}

func forwardKafkaLogEventToLogger(logChan chan kafka.LogEvent, logger *zap.Logger) {
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

// Log syslog level, lower is more critical.
// See: https://en.wikipedia.org/wiki/Syslog#Severity_level
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

func retry(maxRetries int, delay time.Duration, fn func() error, retryOnErrFn func(err error) bool) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil // success
		} else if retryOnErrFn(err) {
			time.Sleep(delay) // retry
		} else {
			break // failure
		}
	}
	return err
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
