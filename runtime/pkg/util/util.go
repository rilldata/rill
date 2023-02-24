// Package util provides general purpose utility functions
package util

// ReturnFirstErr returns first non nil error from a variable list of errors
func ReturnFirstErr(errs ...error) error {
	for _, r := range errs {
		if r != nil {
			return r
		}
	}
	return nil
}
