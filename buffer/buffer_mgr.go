package buffer

import (
	"sync"
)

type BufferPoolMgr struct {
	defaultBufferSize int       // if size<defaultBufferSize ,get from defaultPool ,if not get from size5Pool
	defaultPool       sync.Pool // cache cap(data)<=defaultBufferSize
	size5Pool         sync.Pool // cache defaultBufferSize<cap(data)<=defaultBufferSize*5 and
	size10Pool        sync.Pool // cache 10*defaultBufferSize<cap(data)
	anyBiggerPool     sync.Pool // 10*defaultBufferSize>cap(data)
}

func NewBufferPoolMgr(defaultBufferSize int) *BufferPoolMgr {
	return &BufferPoolMgr{
		defaultBufferSize: defaultBufferSize,
		defaultPool:       sync.Pool{},
		size5Pool:         sync.Pool{},
		anyBiggerPool:     sync.Pool{},
	}
}

func (pool *BufferPoolMgr) Put(data []byte) {
	var capSize int
	if data == nil {
		return
	}
	capSize = cap(data)
	if capSize == 0 {
		return
	}

	if capSize <= pool.defaultBufferSize {
		pool.defaultPool.Put(data)
	} else if capSize <= 5*pool.defaultBufferSize {
		pool.size5Pool.Put(data)
	} else if capSize <= 10*pool.defaultBufferSize {
		pool.size10Pool.Put(data)
	} else {
		pool.anyBiggerPool.Put(data)
	}

}

func (pool *BufferPoolMgr) Get(size int) (data []byte) {
	capSize := size
	if size < pool.defaultBufferSize {
		capSize = pool.defaultBufferSize
	}
	if capSize <= pool.defaultBufferSize {
		return pool.getDefault(size)
	} else if capSize <= 5*pool.defaultBufferSize {
		return pool.getFromSize5Pool(size)
	} else if capSize <= 10*pool.defaultBufferSize {
		return pool.getFromSize10Pool(size)
	} else {
		return pool.getFromAnyBiggerPool(size)
	}
}

func (pool *BufferPoolMgr) getByCreate(size int) (data []byte) {
	return make([]byte, size, size)
}
func (pool *BufferPoolMgr) getFromAnyBiggerPool(size int) (data []byte) {
	bufferObj := pool.anyBiggerPool.Get()
	if bufferObj == nil {
		return pool.getByCreate(size)
	}
	bufferData := bufferObj.([]byte)
	if cap(bufferData) >= size {
		bufferData = bufferData[0:size]
		return bufferData
	}
	pool.Put(bufferData) // put it back ,because it is not big enough

	return pool.getByCreate(size)

}
func (pool *BufferPoolMgr) getFromSize10Pool(size int) (data []byte) {
	bufferObj := pool.size10Pool.Get()
	if bufferObj == nil {
		return pool.getFromAnyBiggerPool(size)
	}
	bufferData := bufferObj.([]byte)
	if cap(bufferData) >= size {
		bufferData = bufferData[0:size]
		return bufferData
	}
	pool.Put(bufferData) // put it back ,because it is not big enough

	return pool.getFromAnyBiggerPool(size)

}
func (pool *BufferPoolMgr) getFromSize5Pool(size int) (data []byte) {
	bufferObj := pool.size5Pool.Get()
	if bufferObj == nil {
		return pool.getFromSize10Pool(size)
	}
	bufferData := bufferObj.([]byte)
	if cap(bufferData) >= size {
		bufferData = bufferData[0:size]
		return bufferData
	}
	pool.Put(bufferData) // put it back ,because it is not big enough

	return pool.getFromSize10Pool(size)
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
