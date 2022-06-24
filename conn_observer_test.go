package main

import (
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
)

func TestConnObserver(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:6000")
	defer listener.Close()

	expected := make([]byte, 48)
	rand.Read(expected)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		fd, _ := listener.Accept()

		w := func(c *ConnObserver, data []byte) []byte {
			wg.Done()

			return append(data, 16)
		}

		r := func(c *ConnObserver, data []byte) ([]byte, error) {
			assert.Equal(t, expected, data)
			c.Write(data)

			wg.Done()
			return []byte{}, nil
		}

		_ = NewConnObserver(fd, 1024, w, r, nil)
	}()

	r := func(c *ConnObserver, data []byte) ([]byte, error) {
		assert.Equal(t, len(expected)+1, len(data))
		assert.Equal(t, expected, data[:len(expected)])
		assert.Equal(t, uint8(16), data[len(data)-1])

		wg.Done()
		return []byte{}, nil
	}

	fd, _ := net.Dial("tcp", "localhost:6000")
	conn := NewConnObserver(fd, 1024, nil, r, nil)

	conn.Write(expected)

	wg.Wait()
}

func TestConnObserverBuffer(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:6000")
	defer listener.Close()

	expected := make([]byte, 16)
	rand.Read(expected)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		fd, _ := listener.Accept()

		r := func(c *ConnObserver, data []byte) ([]byte, error) {
			if len(data) < 16 {
				return data, nil
			}

			l := len(data)
			if l > len(expected) {
				l = len(expected)
			}

			assert.Equal(t, expected[:l], data[:l])

			wg.Done()
			return []byte{}, nil
		}

		_ = NewConnObserver(fd, 1024, nil, r, nil)
	}()

	fd, _ := net.Dial("tcp", "localhost:6000")

	for _, b := range expected {
		n, err := fd.Write([]byte{b})
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, n)
	}

	wg.Done()
	wg.Wait()
}

func TestConnObserverProtobuf(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:6000")
	defer listener.Close()

	pktCh := make(chan []byte)

	go func() {
		fd, _ := listener.Accept()

		r := func(c *ConnObserver, data []byte) ([]byte, error) {
			var pkt protos.Packet
			buffer := proto.NewBuffer(data)

			for {
				if len(buffer.Unread()) == 0 {
					break
				}

				err := buffer.DecodeMessage(&pkt)
				if err != nil {
					break
				}

				if pkt.Data != nil {
					pktCh <- pkt.Data
					pkt.Reset()
				}
			}

			return buffer.Unread(), nil
		}

		_ = NewConnObserver(fd, 1024, nil, r, nil)
	}()

	fd, _ := net.Dial("tcp", "localhost:6000")

	for i := 0; i < 100; i++ {
		expected := make([]byte, 128)
		rand.Read(expected)

		buffer := proto.NewBuffer([]byte{})
		err := buffer.EncodeMessage(&protos.Packet{
			Sn:   5,
			Data: expected,
		})
		assert.Equal(t, nil, err)

		ptr := 0
		for {
			if ptr == len(buffer.Bytes()) {
				break
			}

			n, err := fd.Write(buffer.Bytes()[ptr:])
			assert.Equal(t, nil, err)

			ptr += n
		}

		r := <-pktCh
		assert.Equal(t, expected, r)
	}

	time.Sleep(100 * time.Millisecond)
}

func TestConnObserverProtobufHalf(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:6000")
	defer listener.Close()

	pktCh := make(chan []byte)

	go func() {
		fd, _ := listener.Accept()

		r := func(c *ConnObserver, data []byte) ([]byte, error) {
			var pkt protos.Packet
			buffer := proto.NewBuffer(data)

			orig := make([]byte, len(data))
			copy(orig, data)

			nRead := 0

			for {
				if len(buffer.Unread()) == 0 {
					break
				}

				before := len(buffer.Unread())

				err := buffer.DecodeMessage(&pkt)
				if err != nil {
					break
				}

				nRead += before - len(buffer.Unread())

				if pkt.Data != nil {
					pktCh <- pkt.Data
					pkt.Reset()
				}
			}

			return orig[nRead:], nil
		}

		_ = NewConnObserver(fd, 1024, nil, r, nil)
	}()

	fd, _ := net.Dial("tcp", "localhost:6000")

	expected := make([]byte, 128)
	rand.Read(expected)

	buffer := proto.NewBuffer([]byte{})
	err := buffer.EncodeMessage(&protos.Packet{
		Sn:   5,
		Data: expected,
	})
	assert.Equal(t, nil, err)

	n, err := fd.Write(buffer.Bytes()[:50])
	assert.Equal(t, nil, err)
	assert.Equal(t, 50, n)

	time.Sleep(100 * time.Millisecond)

	n, err = fd.Write(buffer.Bytes()[50:])
	assert.Equal(t, nil, err)
	assert.Equal(t, len(buffer.Bytes())-50, n)

	r := <-pktCh
	assert.Equal(t, expected, r)

	time.Sleep(100 * time.Millisecond)
}
