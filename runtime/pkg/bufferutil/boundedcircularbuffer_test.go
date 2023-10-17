package bufferutil

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoundedCircularBuffer(t *testing.T) {
	cb := NewBoundedCircularBuffer[int](3, 10)
	cb.Push(Item[int]{1, 4})
	cb.Push(Item[int]{2, 4})
	i := 1
	cb.Iterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i++
	}, cb.Count())
	cb.Push(Item[int]{3, 4})
	i = 2
	cb.Iterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i++
	}, cb.Count())
	item, err := cb.Pop()
	assert.NoError(t, err)
	assert.Equal(t, 2, item.Value)
}
