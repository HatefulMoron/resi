package main

import (
	"errors"
	"net"
	"sync"
)

var errListenerClosed = errors.New("listener closed.")

type SocketListener struct {
	listeners      []net.Listener
	acceptCh       chan net.Conn
	acceptObserver chan net.Conn
}

func NewSocketListener(listeners []net.Listener) *SocketListener {
	l := &SocketListener{
		listeners:      listeners,
		acceptCh:       make(chan net.Conn, 8),
		acceptObserver: nil,
	}

	go l.pumpChannel()

	return l
}

func (l *SocketListener) SetAcceptObserver(ch chan net.Conn) {
	l.acceptObserver = ch
}

func (l *SocketListener) Addr() net.Addr {
	return l.listeners[0].Addr()
}

func (l *SocketListener) Close() error {
	for _, fd := range l.listeners {
		_ = fd.Close()
	}
	return nil
}

func (l *SocketListener) pumpChannel() {
	var wg sync.WaitGroup
	wg.Add(len(l.listeners))
	defer func() {
		wg.Wait()
		close(l.acceptCh)
	}()

	for _, fd := range l.listeners {
		lFd := fd
		go func() {
			defer wg.Done()

			for {

				conn, err := lFd.Accept()

				if err != nil {
					l.Close()
					return
				}

				sock := NewSocket([]net.Conn{conn})

				l.acceptCh <- sock

				if l.acceptObserver != nil {
					l.acceptObserver <- sock
				}

			}
		}()
	}
}

func (l *SocketListener) Accept() (net.Conn, error) {
	fd, ok := <-l.acceptCh
	if !ok {
		return nil, errListenerClosed
	}

	return fd, nil
}
