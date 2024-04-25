package formatter

import (
	"fmt"
)

type nonFormatter struct{}

func newNonFormatter() *nonFormatter {
	return &nonFormatter{}
}

func (f *nonFormatter) StringFormat(x any) (string, error) {
	return fmt.Sprintf("%v", x), nil
}
