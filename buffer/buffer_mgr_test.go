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

	data = pool.Get(4*5 + 1)
	assert.Equal(t, len(data), 4*5+1)
	assert.True(t, cap(data) >= 4*5+1)

	data = pool.Get(4*5 + 2)
	assert.Equal(t, len(data), 4*5+2)
	assert.True(t, cap(data) >= 4*5+2)
	pool.Put(data)

	data = pool.Get(4*5 + 1)
	assert.Equal(t, len(data), 4*5+1)
	assert.True(t, cap(data) >= 4*5+1)

	data = pool.Get(4*10 + 1)
	assert.Equal(t, len(data), 4*10+1)
	assert.True(t, cap(data) >= 4*10+1)

}
