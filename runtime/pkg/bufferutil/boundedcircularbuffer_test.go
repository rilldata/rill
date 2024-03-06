package bufferutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	i = 3
	cb.ReverseIterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i--
	}, cb.Count())
}

func TestBoundedCircularBufferWithLimits(t *testing.T) {
	cb := NewBoundedCircularBuffer[int](3, 10)
	cb.Push(Item[int]{1, 4})
	cb.Push(Item[int]{2, 4})
	i := 2
	cb.Iterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i++
	}, 1)
	cb.Push(Item[int]{3, 4})
	i = 3
	cb.ReverseIterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i--
	}, 1)
	item, err := cb.Pop()
	assert.NoError(t, err)
	assert.Equal(t, 2, item.Value)
	cb.Push(Item[int]{4, 4})
	i = 3
	cb.Iterate(func(item Item[int]) {
		assert.Equal(t, i, item.Value)
		i++
	}, cb.Count())
	item, err = cb.Pop()
	assert.NoError(t, err)
	assert.Equal(t, 3, item.Value)
	item, err = cb.Pop()
	assert.NoError(t, err)
	assert.Equal(t, 4, item.Value)
}

func TestBoundedCircularBuffer_ReverseIterateUntil(t *testing.T) {
	cb := NewBoundedCircularBuffer[int](10, 40)
	cb.Push(Item[int]{1, 4})
	cb.Push(Item[int]{2, 4})
	cb.Push(Item[int]{3, 4})
	cb.Push(Item[int]{4, 4})
	cb.Push(Item[int]{5, 4})
	cb.Push(Item[int]{6, 4})
	cb.Push(Item[int]{7, 4})
	cb.Push(Item[int]{8, 4})
	cb.Push(Item[int]{9, 4})
	cb.Push(Item[int]{10, 4})
	lastItem := -1
	numItems := 0
	cb.ReverseIterateUntil(func(item Item[int]) bool {
		numItems++
		lastItem = item.Value
		return item.Value > 8
	})
	assert.Equal(t, 8, lastItem)
	assert.Equal(t, 3, numItems)
}
