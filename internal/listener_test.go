package internal

import (
	"context"
	"math/rand"
	"sync"
	"testing"

	protos "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
)

func vanillaTcp(l *ResiListener, t *testing.T) {
	expected := make([]byte, 48)
	rand.Read(expected)

	assert.Equal(t, "127.0.0.1:7000", l.TcpAddr().String())

	client, err := backClient("127.0.0.1:7000")
	assert.Equal(t, nil, err)

	resp, err := client.Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: expected,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, resp.Data)
}

func vanillaNKN(l *ResiListener, t *testing.T) {
	expected := make([]byte, 48)
	rand.Read(expected)

	client, err := backClient(l.NKNAddr().String())
	assert.Equal(t, nil, err)

	resp, err := client.Test(context.Background(), &protos.TestRequest{
		Ext:  nil,
		Data: expected,
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, resp.Data)
}

func TestListenerVanillaTCP(t *testing.T) {
	l, err := NewResiListener(SocketAddr{
		Net: "tcp",
		Str: "127.0.0.1:7000",
	})
	assert.Equal(t, nil, err)

	go l.Serve()
	vanillaTcp(l, t)

	l.Stop()
}

func TestListenerVanillaNKN(t *testing.T) {
	l, err := NewResiListener(SocketAddr{
		Net: "tcp",
		Str: "127.0.0.1:7000",
	})
	assert.Equal(t, nil, err)

	go l.Serve()
	vanillaTcp(l, t)

	l.Stop()
}

func TestListenerVanillaBoth(t *testing.T) {
	l, err := NewResiListener(SocketAddr{
		Net: "tcp",
		Str: "127.0.0.1:7000",
	})
	assert.Equal(t, nil, err)

	go l.Serve()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		vanillaTcp(l, t)
		wg.Done()
	}()
	go func() {
		vanillaNKN(l, t)
		wg.Done()
	}()

	wg.Wait()

	l.Stop()
}
