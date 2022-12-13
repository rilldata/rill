package queries

import "fmt"

func quoteName(name string) string {
	return fmt.Sprintf("%q", name)
}
