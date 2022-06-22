package main

import (
	"bytes"
	"context"
	"log"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	helloworld "github.com/hatefulmoron/resi/protos"
	resi "github.com/hatefulmoron/resi/protos"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestSocketSingle(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:5000")
	var wg sync.WaitGroup
	wg.Add(1)

	addr := listener.Addr().String()
	expected := make([]byte, 48)
	rand.Read(expected)

	go func() {

		for {
			fd1, err := listener.Accept()
			if err != nil {
				break
			}

			fd2, err := listener.Accept()
			if err != nil {
				break
			}

			sock := NewSocket([]net.Conn{fd1})
			sock.AddSocket(fd2)
			defer sock.Close()

			buffer := make([]byte, 4096)

			for {

				n, err := sock.Read(buffer)
				if err == errSesisonClosed {
					break
				}

				assert.Equal(t, nil, err)

				if n == 0 {
					break
				}

				assert.Equal(t, len(expected), n)
				assert.Equal(t, expected, buffer[0:n])
				wg.Done()
			}
		}

		wg.Done()
	}()

	{
		a, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{a})
		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)

		time.Sleep(time.Duration(100) * time.Millisecond)

		listener.Close()
		a.Close()
	}

	wg.Wait()
}

func TestSocketWriteRead(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:5000")
	var wg sync.WaitGroup
	wg.Add(2)

	addr := listener.Addr().String()
	expected := make([]byte, 48)
	rand.Read(expected)

	go func() {

		for {
			fd1, err := listener.Accept()
			if err != nil {
				break
			}

			fd2, err := listener.Accept()
			if err != nil {
				break
			}

			sock := NewSocket([]net.Conn{fd1})
			sock.AddSocket(fd2)
			defer sock.Close()

			buffer := make([]byte, 4096)

			for {

				n, err := sock.Read(buffer)
				if err == errSesisonClosed {
					break
				}

				assert.Equal(t, nil, err)

				if n == 0 {
					break
				}

				assert.Equal(t, len(expected), n)
				assert.Equal(t, expected, buffer[0:n])
				wg.Done()
			}
		}

		wg.Done()
	}()

	{
		a, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)
		b, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{a, b})
		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)

		time.Sleep(time.Duration(100) * time.Millisecond)

		listener.Close()
		a.Close()
		b.Close()
	}

	wg.Wait()
}

func TestSocketReadMirrorDelay(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:5001")
	var wg sync.WaitGroup
	wg.Add(2)

	addr := listener.Addr().String()
	expected := make([]byte, 48)
	rand.Read(expected)

	go func() {

		for {
			fd1, err := listener.Accept()
			if err != nil {
				break
			}

			fd2, err := listener.Accept()
			if err != nil {
				break
			}

			sock := NewSocket([]net.Conn{fd1})
			sock.AddSocket(fd2)
			defer sock.Close()

			buffer := make([]byte, 4096)

			for {

				n, err := sock.Read(buffer)
				if err == errSesisonClosed {
					break
				}

				assert.Equal(t, nil, err)

				if n == 0 {
					break
				}

				assert.Equal(t, len(expected), n)
				assert.Equal(t, expected, buffer[0:n])
				wg.Done()
			}
		}

		wg.Done()
	}()

	{
		a, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)
		b, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)

		wData, err := proto.Marshal(&resi.Packet{
			Sn:   0,
			Data: expected,
		})
		assert.Equal(t, nil, err)

		n, err := a.Write(wData)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(wData), n)

		time.Sleep(time.Duration(100) * time.Millisecond)

		n, err = b.Write(wData)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(wData), n)

		time.Sleep(time.Duration(100) * time.Millisecond)

		listener.Close()
		a.Close()
		b.Close()
	}

	wg.Wait()
}

func TestSocketWriteReadBig(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:5000")
	var wg sync.WaitGroup
	wg.Add(2)

	addr := listener.Addr().String()
	expected := make([]byte, 16*1024*2)
	rand.Read(expected)

	go func() {

		for {
			fd1, err := listener.Accept()
			if err != nil {
				break
			}

			fd2, err := listener.Accept()
			if err != nil {
				break
			}

			sock := NewSocket([]net.Conn{fd1})
			sock.AddSocket(fd2)
			defer sock.Close()

			buffer := make([]byte, 16*1024*2)
			ptr := 0

			for {

				n, err := sock.Read(buffer[ptr:])
				if err == errSesisonClosed {
					break
				}

				assert.Equal(t, nil, err)

				if n == 0 {
					break
				}

				assert.Equal(t, 16*1024, n)
				assert.Equal(t, expected, buffer[ptr:ptr+n])
				ptr = ptr + n
			}

			wg.Done()
		}

		wg.Done()
	}()

	{
		a, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)
		b, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{a, b})
		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, 16*1024, n)

		n, err = assoc.Write(expected[n:])
		assert.Equal(t, nil, err)
		assert.Equal(t, 16*1024, n)

		time.Sleep(time.Duration(100) * time.Millisecond)

		listener.Close()
		a.Close()
		b.Close()
	}

	wg.Wait()
}

func TestSocketPingPong(t *testing.T) {
	listener, _ := net.Listen("tcp", "localhost:5000")
	var wg sync.WaitGroup
	wg.Add(2)

	addr := listener.Addr().String()
	expected := make([]byte, rand.Intn(16*1024))
	rand.Read(expected)

	go func() {

		for {
			fd1, err := listener.Accept()
			if err != nil {
				break
			}

			fd2, err := listener.Accept()
			if err != nil {
				break
			}

			sock := NewSocket([]net.Conn{fd1})
			sock.AddSocket(fd2)
			defer sock.Close()

			buffer := make([]byte, 16*1024)

			for {
				n, err := sock.Read(buffer)
				if err == errSesisonClosed {
					break
				}

				assert.Equal(t, nil, err)

				if n == 0 {
					break
				}

				n, err = sock.Write(buffer[:n])
				assert.Equal(t, nil, err)
			}

			wg.Done()
		}

		wg.Done()
	}()

	{
		a, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)
		b, err := net.Dial("tcp", addr)
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{a, b})

		buffer := make([]byte, 16*1024)

		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)

		n, err = assoc.Read(buffer)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
		assert.True(t, bytes.Equal(buffer[:n], expected))

		time.Sleep(time.Duration(100) * time.Millisecond)

		listener.Close()
		a.Close()
		b.Close()
	}

	wg.Wait()
}

func TestSocketListener(t *testing.T) {

	a, _ := net.Listen("tcp", "localhost:6000")
	b, _ := net.Listen("tcp", "localhost:6001")

	expected := make([]byte, 512)
	rand.Read(expected)

	l := NewSocketListener([]net.Listener{a, b})

	go func() {
		s := grpc.NewServer()

		helloworld.RegisterGreeterServer(s, &server{})
		if err := s.Serve(l); err != nil {
			return
		}

	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", a.Addr().String())
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{conn})
		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
	}()

	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", b.Addr().String())
		assert.Equal(t, nil, err)

		assoc := NewSocket([]net.Conn{conn})
		n, err := assoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
	}()

	wg.Wait()
	l.Close()
}

func TestSocketAbsorb(t *testing.T) {

	a, _ := net.Listen("tcp", "localhost:6000")
	b, _ := net.Listen("tcp", "localhost:6001")

	time.Sleep(time.Duration(100) * time.Millisecond)

	expected := make([]byte, 512)
	rand.Read(expected)

	l := NewSocketListener([]net.Listener{a, b})

	var globalWg sync.WaitGroup
	globalWg.Add(2)

	go func() {

		// Accept two sockets, read a packet of data from each then combine the
		// two. We should then read a single agreeing message

		aConn, _ := l.Accept()
		bConn, _ := l.Accept()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			buf := make([]byte, len(expected))
			_, _ = aConn.Read(buf)

			if !bytes.Equal(buf, expected) {
				panic(buf)
			}
		}()

		go func() {
			defer wg.Done()
			buf := make([]byte, len(expected))
			_, _ = bConn.Read(buf)

			if !bytes.Equal(buf, expected) {
				panic(buf)
			}
		}()

		wg.Wait()

		((aConn).(*Socket)).Absorb(bConn.(*Socket))

		buf := make([]byte, len(expected))
		n, _ := aConn.Read(buf)

		if !bytes.Equal(buf[:n], expected) {
			panic(buf)
		}

		globalWg.Done()
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	aConn, err := net.Dial("tcp", a.Addr().String())
	assert.Equal(t, nil, err)
	aAssoc := NewSocket([]net.Conn{aConn})

	bConn, err := net.Dial("tcp", b.Addr().String())
	assert.Equal(t, nil, err)
	bAssoc := NewSocket([]net.Conn{bConn})

	go func() {
		defer wg.Done()

		n, err := aAssoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
	}()

	go func() {
		defer wg.Done()

		n, err := bAssoc.Write(expected)
		assert.Equal(t, nil, err)
		assert.Equal(t, len(expected), n)
	}()

	wg.Wait()

	time.Sleep(time.Duration(100) * time.Millisecond) // necessary, some race condition

	aAssoc.Absorb(bAssoc)

	n, err := aAssoc.Write(expected)
	assert.Equal(t, nil, err)
	assert.Equal(t, len(expected), n)

	globalWg.Done()
	globalWg.Wait()
	l.Close()

	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestSocketGRPC(t *testing.T) {

	a, _ := net.Listen("tcp", "localhost:6000")

	time.Sleep(time.Duration(100) * time.Millisecond)

	expected := make([]byte, 512)
	rand.Read(expected)

	fdCh := make(chan net.Conn, 2)
	l := NewSocketListener([]net.Listener{a})
	l.SetAcceptObserver(fdCh)

	s := grpc.NewServer()
	tester := &upgradeTestServer{}

	go func() {

		go func() {
			for {
				fd, ok := <-fdCh
				if !ok {
					break
				}

				tester.sockets = append(tester.sockets, fd.(*Socket))
			}
		}()

		helloworld.RegisterGreeterServer(s, tester)
		if err := s.Serve(l); err != nil {
			return
		}
	}()

	aConn, err := net.Dial("tcp", a.Addr().String())
	assert.Equal(t, nil, err)
	aAssoc := NewSocket([]net.Conn{aConn})

	aOpt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return aAssoc, nil
	})

	aGrpcConn, err := grpc.Dial("", aOpt,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Equal(t, nil, err)

	aC := helloworld.NewGreeterClient(aGrpcConn)

	resp, err := aC.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "Thomas",
	})
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello world", resp.Message)

	l.Close()

	time.Sleep(time.Duration(100) * time.Millisecond)
}

func TestSocketSwitchGRPC(t *testing.T) {

	a, _ := net.Listen("tcp", "localhost:6000")
	b, _ := net.Listen("tcp", "localhost:6001")

	time.Sleep(time.Duration(100) * time.Millisecond)

	expected := make([]byte, 512)
	rand.Read(expected)

	fdCh := make(chan net.Conn, 2)
	l := NewSocketListener([]net.Listener{a, b})
	l.SetAcceptObserver(fdCh)

	s := grpc.NewServer()
	tester := &upgradeTestServer{}

	go func() {

		go func() {
			for {
				fd, ok := <-fdCh
				if !ok {
					break
				}

				tester.sockets = append(tester.sockets, fd.(*Socket))
			}
		}()

		helloworld.RegisterGreeterServer(s, tester)
		if err := s.Serve(l); err != nil {
			return
		}
	}()

	aConn, err := net.Dial("tcp", a.Addr().String())
	assert.Equal(t, nil, err)
	aAssoc := NewSocket([]net.Conn{aConn})

	//aAssoc.Debug = true

	bConn, err := net.Dial("tcp", b.Addr().String())
	assert.Equal(t, nil, err)
	bAssoc := NewSocket([]net.Conn{bConn})

	aOpt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return aAssoc, nil
	})
	bOpt := grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
		return bAssoc, nil
	})

	aGrpcConn, err := grpc.Dial("", aOpt,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Equal(t, nil, err)
	bGrpcConn, err := grpc.Dial("", bOpt,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Equal(t, nil, err)

	aC := helloworld.NewGreeterClient(aGrpcConn)
	bC := helloworld.NewGreeterClient(bGrpcConn)

	time.Sleep(time.Duration(100) * time.Millisecond)

	resp, err := aC.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "Thomas",
	})
	log.Printf("%v\n", err)
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello world", resp.Message)

	time.Sleep(time.Duration(100) * time.Millisecond)

	resp, err = bC.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "Thomas",
	})
	log.Printf("%v\n", err)
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello world", resp.Message)

	aAssoc.Absorb(bAssoc)
	tester.Absorb()

	resp, err = aC.SayHello(context.Background(), &helloworld.HelloRequest{
		Name: "Thomas",
	})
	log.Printf("%v\n", err)
	assert.Equal(t, nil, err)
	assert.Equal(t, "hello world", resp.Message)

	log.Println("FINISH")

	l.Close()

	time.Sleep(time.Duration(100) * time.Millisecond)
}
