// Copyright (C) 2016 The Protocol Authors.

package protocol

import (
	"fmt"
	"sync"
)

// Global pool to get buffers from. Requires Blocksizes to be initialised,
// therefore it is initialized in the same init() as BlockSizes
var BufferPool bufferPool

type bufferPool struct {
	pools []sync.Pool
}

func newBufferPool() bufferPool {
	return bufferPool{make([]sync.Pool, len(BlockSizes))}
}

func (p *bufferPool) Get(size int) []byte {
	// Too big, isn't pooled
	if size > MaxBlockSize {
		return make([]byte, size)
	}

	// Try the fitting and all bigger pools
	var bs []byte
	bkt := getBucketForSize(size)
	for j := bkt; j < len(BlockSizes); j++ {
		if intf := p.pools[j].Get(); intf != nil {
			bs = *intf.(*[]byte)
			return bs[:size]
		}
	}

	// All pools are empty, must allocate. For very small slices where we
	// didn't have a block to reuse, just allocate a small slice instead of
	// a large one. We won't be able to reuse it, but avoid some overhead.
	if size < MinBlockSize/64 {
		return make([]byte, size)
	}
	return make([]byte, BlockSizes[bkt])[:size]
}

// Put makes the given byte slice available again in the global pool.
// You must only Put() slices that were returned by Get() or Upgrade().
func (p *bufferPool) Put(bs []byte) {
	// Don't buffer slices outside of our pool range
	if cap(bs) > MaxBlockSize || cap(bs) < MinBlockSize {
		return
	}

	bkt := putBucketForSize(cap(bs))
	p.pools[bkt].Put(&bs)
}

// Upgrade grows the buffer to the requested size, while attempting to reuse
// it if possible.
func (p *bufferPool) Upgrade(bs []byte, size int) []byte {
	if cap(bs) >= size {
		// Reslicing is enough, lets go!
		return bs[:size]
	}

	// It was too small. But it pack into the pool and try to get another
	// buffer.
	p.Put(bs)
	return p.Get(size)
}

// getBucketForSize returns the bucket where we should get a slice of a
// certain size. Each bucket is guaranteed to hold slices that are precisely
// the block size for that bucket, so if the block size is larger than our
// size we are good.
func getBucketForSize(size int) int {
	for i, blockSize := range BlockSizes {
		if size <= blockSize {
			return i
		}
	}

	panic(fmt.Sprintf("bug: tried to get impossible block size %d", size))
}

// putBucketForSize returns the bucket where we should put a slice of a
// certain size. Each bucket is guaranteed to hold slices that are precisely
// the block size for that bucket, so we just find the matching one.
func putBucketForSize(cap int) int {
	for i, blockSize := range BlockSizes {
		if cap == blockSize {
			return i
		}
	}

	panic(fmt.Sprintf("bug: tried to put impossible block size %d", cap))
}
