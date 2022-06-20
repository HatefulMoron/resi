package main

import (
	"sync"
	"testing"
	"time"

	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
)

func TestBasicAssociation(t *testing.T) {
	assoc := NewAssociation(3)
	data := []byte{1, 2, 3, 4, 5}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
	}()
	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
	}()
	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
	}()

	wg.Wait()
}

func TestAssociationBadData(t *testing.T) {
	assoc := NewAssociation(3)
	data := []byte{1, 2, 3, 4, 5}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, false, ret)
	}()
	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, []byte{1, 2, 3})
		assert.Equal(t, false, ret)
	}()
	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, false, ret)
	}()

	wg.Wait()
}

func TestAssociationOne(t *testing.T) {
	assoc := NewAssociation(1)
	data := []byte{1, 2, 3, 4, 5}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
	}()

	wg.Wait()
}

func TestAssociationWait(t *testing.T) {
	assoc := NewAssociation(3)
	data := []byte{1, 2, 3, 4, 5}

	var wg sync.WaitGroup
	wg.Add(3)

	timesCh := make(chan time.Time, 3)

	go func() {
		defer wg.Done()

		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
		timesCh <- time.Now()

	}()
	go func() {
		defer wg.Done()

		time.Sleep(time.Duration(250) * time.Millisecond)

		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)

		assert.Equal(t, true, ret)
		timesCh <- time.Now()

	}()
	go func() {
		defer wg.Done()

		time.Sleep(time.Duration(250) * time.Millisecond)

		ret := assoc.Push(&protos.RedundancyTSN{
			Tsn: 0,
		}, data)
		assert.Equal(t, true, ret)
		timesCh <- time.Now()

	}()

	wg.Wait()
	close(timesCh)

	times := make([]time.Time, 0)
	for {
		t, ok := <-timesCh
		if !ok {
			break
		}
		times = append(times, t)
	}

	for i := 0; i < len(times)-1; i++ {
		d := times[i+1].Sub(times[i])
		assert.Equal(t, true, d < (time.Duration(10)*time.Millisecond))
	}
}
