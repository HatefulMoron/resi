package internal

import (
	"net"
)

type WriteFormatter func(*ConnObserver, []byte) []byte
type ReadObserver func(*ConnObserver, []byte) ([]byte, error)
type CloseObserver func(*ConnObserver, net.Conn)

type ConnObserver struct {
	buffer        []byte
	ptr           int
	fd            net.Conn
	writeCh       chan WriteOperation
	WriteObserver WriteFormatter
	ReadObserver  ReadObserver
	CloseObserver CloseObserver
}

func NewConnObserver(
	fd net.Conn,
	sz int,
	write WriteFormatter,
	read ReadObserver,
	cl CloseObserver,
) *ConnObserver {
	conn := &ConnObserver{
		buffer:        make([]byte, sz),
		ptr:           0,
		fd:            fd,
		writeCh:       make(chan WriteOperation, 1024),
		WriteObserver: write,
		ReadObserver:  read,
		CloseObserver: cl,
	}

	go conn.readPump()
	go conn.writePump()

	return conn
}

func (c *ConnObserver) readPump() {
	for {
		n, err := c.fd.Read(c.buffer[c.ptr:])
		if err != nil {
			c.fd.Close()
			return
		}

		back, err := c.ReadObserver(c, c.buffer[:c.ptr+n])
		if err != nil {
			c.fd.Close()
			return
		}

		copy(c.buffer, back)
		c.ptr = len(back)
	}
}

func (c *ConnObserver) writePump() {
	for {
		op, ok := <-c.writeCh
		if !ok {
			return
		}

		formatted := op.data

		if c.WriteObserver != nil {
			formatted = c.WriteObserver(c, op.data)
		}

		ptr := 0
		for {
			if ptr == len(formatted) {
				break
			}

			n, err := c.fd.Write(formatted[ptr:])
			if err != nil {
				c.fd.Close()
				return
			}

			ptr += n
		}

		op.status <- WriteStatus{
			n:   len(formatted),
			err: nil,
		}
	}
}

func (c *ConnObserver) Write(buf []byte) {
	ch := make(chan WriteStatus)
	c.writeCh <- WriteOperation{
		data:   buf,
		status: ch,
	}

	_ = <-ch
	//status := <-ch
	//return status.err
}

func (c *ConnObserver) Close() error {
	if c.CloseObserver != nil {
		c.CloseObserver(c, c.fd)
	}

	close(c.writeCh)
	return c.fd.Close()
}

func (c *ConnObserver) LocalAddr() net.Addr {
	return c.fd.LocalAddr()
}

func (c *ConnObserver) RemoteAddr() net.Addr {
	return c.fd.RemoteAddr()
}

func (c *ConnObserver) Fd() net.Conn {
	return c.fd
}
