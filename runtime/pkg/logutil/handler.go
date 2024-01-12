package logutil

import (
	"github.com/rilldata/rill/runtime/pkg/logbuffer"
	"go.uber.org/zap/zapcore"
)

type BufferedZapCore struct {
	fields []zapcore.Field
	logs   *logbuffer.Buffer
}

func NewBufferedZapCore(logs *logbuffer.Buffer) *BufferedZapCore {
	return &BufferedZapCore{
		logs: logs,
	}
}

func (d *BufferedZapCore) Enabled(level zapcore.Level) bool {
	return true
}

func (d *BufferedZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	fields = append(d.fields, fields...)
	return d.logs.AddZapEntry(entry, fields)
}

func (d *BufferedZapCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *d
	clone.fields = append(clone.fields, fields...)
	return &clone
}

func (d *BufferedZapCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	checkedEntry.AddCore(entry, d)
	return checkedEntry
}

func (d *BufferedZapCore) Sync() error {
	return nil
}
