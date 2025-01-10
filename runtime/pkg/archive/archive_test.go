package archive

import (
	"fmt"
	"os"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
)

func TestEstimateTarSize(t *testing.T) {
	tempFiles := []drivers.DirEntry{}
	tmpRoot := "."

	// Create a temporary files for testing
	for i := 0; i < 10; i++ {
		file, err := os.CreateTemp(tmpRoot, "rill-test-")
		if err != nil {
			t.Fatalf("Failed to create temporary file: %v", err)
		}
		defer os.Remove(file.Name())

		_, err = file.WriteString(fmt.Sprintf("test file %d", i))
		if err != nil {
			t.Fatalf("Failed to write to temporary file: %v", err)
		}

		tempFiles = append(tempFiles, drivers.DirEntry{
			IsDir: false,
			Path:  file.Name(),
		})

		err = file.Close()
		if err != nil {
			t.Fatalf("Failed to close temporary file: %v", err)
		}
	}

	size, err := EstimateTarSize(tempFiles, tmpRoot)
	if err != nil {
		t.Fatalf("Failed to estimate tar size: %v", err)
	}

	if size <= 0 {
		t.Fatalf("Invalid tar size: %d", size)
	}

	t.Logf("Estimated tar size: %d", size)
}
