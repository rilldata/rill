package printer

import (
	"encoding/json"
	"errors"
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

type Printer struct {
	Format      Format
	humanOut    io.Writer
	resourceOut io.Writer
}

func NewPrinter(format Format) *Printer {
	return &Printer{
		Format: format,
	}
}

// SetHumanOutput sets the output for human readable messages.
func (p *Printer) SetHumanOutput(out io.Writer) {
	p.humanOut = out
}

// SetResourceOutput sets the output for pringing resources via PrintResource.
func (p *Printer) SetResourceOutput(out io.Writer) {
	p.resourceOut = out
}

func (p *Printer) PrintResource(v interface{}) error {
	if p.Format == FormatUnspecified {
		return errors.New("printer.Format is not set")
	}

	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	switch p.Format {
	case FormatHuman:
		var b strings.Builder
		tableprinter.Print(&b, v)
		fmt.Fprint(out, b.String())
		return nil
	case FormatJSON:
		return p.PrintJSON(v)
	case FormatCSV:
		return p.PrintCSV(v)
	}
	return fmt.Errorf("unknown printer.Format: %T", p.Format)
}

func (p *Printer) PrintCSV(v interface{}) error {
	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	buf, err := gocsv.MarshalString(v)
	if err != nil {
		return fmt.Errorf("failed to marshal CSV: %w", err)
	}

	fmt.Fprint(out, buf)
	return nil
}

func (p *Printer) PrintJSON(v interface{}) error {
	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(out, string(buf))
	return nil
}

func (p *Printer) Printf(format string, i ...interface{}) {
	fmt.Fprintf(p.out(), format, i...)
}

func (p *Printer) Println(i ...interface{}) {
	fmt.Fprintln(p.out(), i...)
}

func (p *Printer) Print(i ...interface{}) {
	fmt.Fprint(p.out(), i...)
}

func (p *Printer) PrintlnSuccess(str string) {
	p.Println(BoldGreen(str))
}

func (p *Printer) PrintlnWarn(str string) {
	p.Println(BoldYellow(str))
}

func (p *Printer) PrintlnError(str string) {
	p.Println(BoldRed(str))
}

func (p *Printer) PrintlnInfo(str string) {
	p.Println(BoldWhite(str))
}

func (p *Printer) PrintBold(str string) {
	p.Print(Bold(str))
}

// BoldGreen returns a string formatted with green and bold.
func BoldGreen(msg interface{}) string {
	return color.New(color.FgGreen).Add(color.Bold).Sprint(msg)
}

// BoldYellow returns a string formatted with yellow and bold.
func BoldYellow(msg interface{}) string {
	return color.New(color.FgYellow).Add(color.Bold).Sprint(msg)
}

// BoldRed returns a string formatted with red and bold.
func BoldRed(msg interface{}) string {
	return color.New(color.FgRed).Add(color.Bold).Sprint(msg)
}

// BoldWhite returns a string formatted with white and bold.
func BoldWhite(msg interface{}) string {
	return color.New(color.FgWhite).Add(color.Bold).Sprint(msg)
}

// Bold returns a string formatted with bold.
func Bold(msg interface{}) string {
	return color.New(color.Bold).Sprint(msg)
}

func (p *Printer) out() io.Writer {
	if p.humanOut != nil {
		return p.humanOut
	}

	if p.Format == FormatHuman {
		return color.Output
	}

	return io.Discard
}
