package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gocarina/gocsv"
	"github.com/lensesio/tableprinter"
)

type Format int

const (
	FormatUnspecified Format = iota
	FormatHuman
	FormatJSON
	FormatCSV
)

func (f Format) String() string {
	switch f {
	case FormatHuman:
		return "human"
	case FormatJSON:
		return "json"
	case FormatCSV:
		return "csv"
	}
	return "unknown format"
}

func (f *Format) Type() string {
	return "string"
}

func (f *Format) Set(s string) error {
	var v Format
	switch s {
	case "human":
		v = FormatHuman
	case "json":
		v = FormatJSON
	case "csv":
		v = FormatCSV
	default:
		return fmt.Errorf("failed to parse Format: %q. Valid values: %+v", s, []string{"human", "json", "csv"})
	}
	*f = v
	return nil
}

var (
	ColorBold       = color.New(color.Bold)
	ColorGreenBold  = color.New(color.FgGreen).Add(color.Bold)
	ColorYellowBold = color.New(color.FgYellow).Add(color.Bold)
	ColorRedBold    = color.New(color.FgRed).Add(color.Bold)
)

// Printer is a helper for printing output in a specific format.
// It differentiates between human-readable and machine-readable output.
// Regular log messages are always produced as human-readable output.
// Human-readable output is discarded when the format is not FormatHuman.
type Printer struct {
	Format           Format
	humanOutOverride io.Writer
	dataOutOverride  io.Writer
}

func NewPrinter(format Format) *Printer {
	return &Printer{
		Format: format,
	}
}

func (p *Printer) PrintData(v interface{}) {
	out := p.dataOut()
	switch p.Format {
	case FormatHuman:
		var b strings.Builder
		tableprinter.Default.RowCharLimit = 120
		tableprinter.Print(&b, v)
		fmt.Fprint(out, b.String())
	case FormatJSON:
		buf, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(fmt.Errorf("failed to marshal JSON: %w", err))
		}
		fmt.Fprintln(out, string(buf))
	case FormatCSV:
		buf, err := gocsv.MarshalString(v)
		if err != nil {
			panic(fmt.Errorf("failed to marshal CSV: %w", err))
		}
		fmt.Fprint(out, buf)
	default:
		panic(fmt.Errorf("unexpected print format <%v>", p.Format))
	}
}

func (p *Printer) PrintDataWithTitle(v interface{}, title string) {
	if p.Format == FormatHuman {
		p.Printf("  %s\n", strings.ToUpper(title))
	}
	p.PrintData(v)
}

func (p *Printer) Print(i ...interface{}) {
	fmt.Fprint(p.humanOut(), i...)
}

func (p *Printer) Println(i ...interface{}) {
	fmt.Fprintln(p.humanOut(), i...)
}

func (p *Printer) Printf(format string, i ...interface{}) {
	fmt.Fprintf(p.humanOut(), format, i...)
}

func (p *Printer) PrintfBold(str string, i ...interface{}) {
	p.Print(ColorBold.Sprintf(str, i...))
}

func (p *Printer) PrintfSuccess(str string, i ...interface{}) {
	p.Print(ColorGreenBold.Sprintf(str, i...))
}

func (p *Printer) PrintfWarn(str string, i ...interface{}) {
	p.Print(ColorYellowBold.Sprintf(str, i...))
}

func (p *Printer) PrintfError(str string, i ...interface{}) {
	p.Print(ColorRedBold.Sprintf(str, i...))
}

func (p *Printer) OverrideHumanOutput(out io.Writer) {
	p.humanOutOverride = out
}

func (p *Printer) OverrideDataOutput(out io.Writer) {
	p.dataOutOverride = out
}

func (p *Printer) humanOut() io.Writer {
	if p.humanOutOverride != nil {
		return p.humanOutOverride
	}

	if p.Format == FormatHuman {
		return color.Output
	}

	return io.Discard
}

func (p *Printer) dataOut() io.Writer {
	if p.dataOutOverride != nil {
		return p.dataOutOverride
	}
	return os.Stdout
}

// FormatBytes converts bytes to human readable format
func (p *Printer) FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KiB", float64(bytes)/KB)
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MiB", float64(bytes)/MB)
	} else if bytes < TB {
		return fmt.Sprintf("%.1f GiB", float64(bytes)/GB)
	}
	return fmt.Sprintf("%.1f TiB", float64(bytes)/TB)
}

// FormatNumber formats a number with appropriate suffix (K, M, B, etc.)
func (p *Printer) FormatNumber(num int64) string {
	if num < 1000 {
		return fmt.Sprintf("%d", num)
	} else if num < 1000000 {
		return fmt.Sprintf("%.1fK", float64(num)/1000)
	} else if num < 1000000000 {
		return fmt.Sprintf("%.1fM", float64(num)/1000000)
	}
	return fmt.Sprintf("%.1fB", float64(num)/1000000000)
}

// FormatValue formats a value for display, avoiding scientific notation for numbers
func (p *Printer) FormatValue(val interface{}) string {
	if val == nil {
		return "null"
	}

	switch v := val.(type) {
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		if v == float32(int32(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
