package buffer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuffer(t *testing.T) {
	pool := NewBufferPoolMgr(4)
	data := pool.Get(3)
	assert.Equal(t, len(data), 3)
	assert.Equal(t, cap(data), 4)

	data = pool.Get(3)
	assert.Equal(t, len(data), 3)
	assert.Equal(t, cap(data), 4)

	data = pool.Get(5)
	assert.Equal(t, len(data), 5)
	assert.Equal(t, cap(data), 5)

	data = pool.Get(5)
	assert.Equal(t, len(data), 5)
	assert.Equal(t, cap(data), 5)

	pool.Put(data)

}
