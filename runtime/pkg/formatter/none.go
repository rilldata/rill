package formatter

import (
	"fmt"
)

type nonFormatter struct{}

func newNonFormatter() *nonFormatter {
	return &nonFormatter{}
}

func (f *nonFormatter) stringFormat(x any) (string, error) {
	return fmt.Sprintf("%v", x), nil
}
