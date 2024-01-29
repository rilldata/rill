package logutil

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/logbuffer"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BufferedZapCore struct {
	fields []zapcore.Field
	logs   *logbuffer.Buffer
	enc    zapcore.Encoder
}

func NewBufferedZapCore(logs *logbuffer.Buffer) *BufferedZapCore {
	encCfg := zap.NewDevelopmentEncoderConfig()
	encCfg.NameKey = zapcore.OmitKey
	encCfg.LevelKey = zapcore.OmitKey
	encCfg.TimeKey = zapcore.OmitKey
	encCfg.MessageKey = zapcore.OmitKey
	encCfg.SkipLineEnding = true
	fieldsOnlyEncoder := zapcore.NewJSONEncoder(encCfg)

	return &BufferedZapCore{
		logs: logs,
		enc:  fieldsOnlyEncoder,
	}
}

func (d *BufferedZapCore) Enabled(level zapcore.Level) bool {
	return true
}

func (d *BufferedZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	fields = append(fields, d.fields...)
	// encode fields using zapcore.Encoder, send empty entry as we want to store the message separately
	fieldsBuf, err := d.enc.EncodeEntry(zapcore.Entry{}, fields)
	if err != nil {
		return err
	}
	defer fieldsBuf.Free()
	payload := fieldsBuf.String()

	return d.logs.AddEntry(zapLevelToPBLevel(entry.Level), entry.Time, entry.Message, payload)
}

func (d *BufferedZapCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *d
	clone.fields = make([]zapcore.Field, len(d.fields)+len(fields))
	copy(clone.fields, d.fields)
	copy(clone.fields[len(d.fields):], fields)
	return &clone
}

func (d *BufferedZapCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	checkedEntry = checkedEntry.AddCore(entry, d)
	return checkedEntry
}

func (d *BufferedZapCore) Sync() error {
	return nil
}

func zapLevelToPBLevel(level zapcore.Level) runtimev1.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return runtimev1.LogLevel_LOG_LEVEL_DEBUG
	case zapcore.InfoLevel:
		return runtimev1.LogLevel_LOG_LEVEL_INFO
	case zapcore.WarnLevel:
		return runtimev1.LogLevel_LOG_LEVEL_WARN
	case zapcore.ErrorLevel:
		return runtimev1.LogLevel_LOG_LEVEL_ERROR
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return runtimev1.LogLevel_LOG_LEVEL_FATAL
	default:
		return runtimev1.LogLevel_LOG_LEVEL_UNSPECIFIED
	}
}
