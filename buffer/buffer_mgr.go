package buffer

import (
	"sync"
)

type BufferPoolMgr struct {
	defaultBufferSize int // if size<defaultBufferSize ,get from defaultPool ,if not get from biggerPool
	defaultPool       sync.Pool
	biggerPool        sync.Pool
}

func NewBufferPoolMgr(defaultBufferSize int) *BufferPoolMgr {
	return &BufferPoolMgr{
		defaultBufferSize: defaultBufferSize,
		defaultPool:       sync.Pool{},
		biggerPool:        sync.Pool{},
	}
}

func (pool *BufferPoolMgr) Put(data []byte) {
	if cap(data) > pool.defaultBufferSize {
		pool.biggerPool.Put(data)
	} else {
		pool.defaultPool.Put(data)
	}

}

func (pool *BufferPoolMgr) getBigger(size int) (data []byte) {
	bufferObj := pool.biggerPool.Get()
	if bufferObj == nil {
		data = make([]byte, size, size)
		return
	}
	bufferData := bufferObj.([]byte)
	if cap(bufferData) >= size {
		bufferData = bufferData[0:size]
		return bufferData
	}
	pool.Put(bufferData) // put it back ,because it is not big enough

	data = make([]byte, size, size)
	return

}
func (pool *BufferPoolMgr) Get(size int) (data []byte) {
	capSize := size
	if size < pool.defaultBufferSize {
		capSize = pool.defaultBufferSize
	}
	if capSize > pool.defaultBufferSize {
		return pool.getBigger(size)
	}
	return pool.getDefault(size)

}
func (pool *BufferPoolMgr) getDefault(size int) (data []byte) {
	bufferObj := pool.defaultPool.Get()
	if bufferObj == nil {
		data = make([]byte, size, pool.defaultBufferSize)
		return
	}
	bufferData := bufferObj.([]byte)
	if cap(bufferData) >= size {
		bufferData = bufferData[0:size]
		return bufferData
	}
	pool.Put(bufferData) // put it back ,because it is not big enough

	data = make([]byte, size, pool.defaultBufferSize)
	return

}
