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
	return d.logs.AddZapEntry(entry, d.fields, fields)
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
