package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_driver_Open(t *testing.T) {
	d := driver{}
	got, err := d.Open("https://github.com/anshulk-13/test-rill-2?github_installation_id=34698894", nil)

	assert.True(t, got != nil)
	assert.NoError(t, err)
}
