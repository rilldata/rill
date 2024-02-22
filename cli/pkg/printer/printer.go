package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func (p *Printer) Print(i ...interface{}) {
	fmt.Fprint(p.humanOut(), i...)
}

func (p *Printer) Println(i ...interface{}) {
	fmt.Fprintln(p.humanOut(), i...)
}

func (p *Printer) Printf(format string, i ...interface{}) {
	fmt.Fprintf(p.humanOut(), format, i...)
}

func (p *Printer) PrintBold(str string) {
	p.Print(ColorBold.Sprint(str))
}

func (p *Printer) PrintlnSuccess(str string) {
	p.Println(ColorGreenBold.Sprint(str))
}

func (p *Printer) PrintlnWarn(str string) {
	p.Println(ColorYellowBold.Sprint(str))
}

func (p *Printer) PrintlnError(str string) {
	p.Println(ColorRedBold.Sprint(str))
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
