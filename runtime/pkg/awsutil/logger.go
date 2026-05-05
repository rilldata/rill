package awsutil

import (
	"fmt"

	"github.com/aws/smithy-go/logging"
	"go.uber.org/zap"
)

type zapLogger struct {
	logger *zap.Logger
}

// NewAWSLogger returns a logging.Logger that routes AWS SDK log output through the given Zap logger.
func NewAWSLogger(logger *zap.Logger) logging.Logger {
	return &zapLogger{logger: logger.Named("aws")}
}

func (l *zapLogger) Logf(classification logging.Classification, format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	switch classification {
	case logging.Warn:
		l.logger.Warn(msg)
	case logging.Debug:
		l.logger.Debug(msg)
	default:
		l.logger.Info(msg)
	}
}
