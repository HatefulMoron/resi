package main

import (
	"bytes"
	"sync"

	protos "github.com/hatefulmoron/resi/protos"
)

type Chunk struct {
	n    int
	tsn  uint64
	data []byte
	lock sync.Mutex
	wg   sync.WaitGroup
	good bool
}

type Association struct {
	n        int
	messages map[uint64]*Chunk
	lock     sync.Mutex
}

func NewChunk(n int, tsn uint64, data []byte) *Chunk {

	chunk := &Chunk{
		n:    1,
		tsn:  tsn,
		data: data,
		good: true,
	}

	chunk.wg.Add(n)
	chunk.lock.Lock()

	return chunk
}

func NewAssociation(n int) *Association {
	return &Association{
		n:        n,
		messages: make(map[uint64]*Chunk),
	}
}

// Blocks until resolution
func (a *Association) Push(tsn *protos.RedundancyTSN, data []byte) bool {
	var chunk *Chunk

	a.lock.Lock()

	chunk, ok := a.messages[tsn.Tsn]
	if !ok {
		chunk = NewChunk(a.n, tsn.Tsn, data)
		a.messages[tsn.Tsn] = chunk

		chunk.wg.Done()
		a.lock.Unlock()

		goto wait
	}

	a.lock.Unlock()
	{
		if bytes.Compare(data, chunk.data) != 0 {
			chunk.good = false
			chunk.wg.Done()
			return false
		}

		// ok
		chunk.wg.Done()
	}

wait:
	chunk.wg.Wait()
	return chunk.good
}
