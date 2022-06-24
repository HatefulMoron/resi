package main

import (
	"bytes"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	protos "github.com/hatefulmoron/resi/protos"
	resi "github.com/hatefulmoron/resi/protos"
)

var errSesisonClosed = errors.New("session closed")

const maxPacketSize = 16 * 1024

type SocketAddr struct {
	network string
	str     string
}

type WriteStatus struct {
	n   int
	err error
}

type ReadStatus struct {
	fd  net.Conn
	pkt *resi.Packet
}

type WriteOperation struct {
	data   []byte
	status chan WriteStatus
}

type BufferedChunk struct {
	data  []byte
	conns []*ConnObserver
}

type Socket struct {
	lock       sync.Mutex
	sockets    []*ConnObserver
	buffered   map[uint64]*BufferedChunk
	inflight   []byte
	flightLock sync.Mutex
	flightCond *sync.Cond
	tsn        uint64
	Debug      bool
}

func NewSocket(sockets []net.Conn) *Socket {

	sock := &Socket{
		sockets:  make([]*ConnObserver, len(sockets)),
		inflight: make([]byte, 0),
		buffered: make(map[uint64]*BufferedChunk),
		tsn:      0,
	}

	sock.flightCond = sync.NewCond(&sock.flightLock)

	for i := range sockets {
		sock.sockets[i] = NewConnObserver(sockets[i],
			128*1024,
			sock.formatWrite,
			sock.formatRead,
			sock.handleClose)
	}

	return sock
}

func (a *Socket) Absorb(b *Socket) {
	a.lock.Lock()
	defer a.lock.Unlock()
	b.lock.Lock()
	defer b.lock.Unlock()

	for _, fd := range b.sockets {
		fd.WriteObserver = a.formatWrite
		fd.ReadObserver = a.formatRead
		fd.CloseObserver = a.handleClose
	}
	a.sockets = append(a.sockets, b.sockets...)
	b.sockets = []*ConnObserver{}
}

func (a *Socket) formatWrite(conn *ConnObserver, data []byte) []byte {
	a.lock.Lock()
	defer a.lock.Unlock()

	if len(a.sockets) == 1 {
		return data
	} else {
		tsn := a.tsn

		buffer := proto.NewBuffer([]byte{})
		_ = buffer.EncodeMessage(&protos.Packet{
			Sn:   tsn,
			Data: data,
		})
		return buffer.Bytes()
	}
}

func (a *Socket) receivedChunk(
	conn *ConnObserver,
	pkt *protos.Packet,
) ([]byte, error) {
	chunk, ok := a.buffered[pkt.Sn]
	if !ok {
		a.buffered[pkt.Sn] = &BufferedChunk{
			data:  pkt.Data,
			conns: []*ConnObserver{conn},
		}
		return nil, nil
	}

	// if already contains this conn, the other side of the connection is not
	// doing its job correctly
	for _, v := range chunk.conns {
		if v == conn {
			panic("todo")
		}
	}

	chunk.conns = append(chunk.conns, conn)

	// is the data incorrect?
	if !bytes.Equal(pkt.Data, chunk.data) {
		panic("todo bad data")
	}

	// do we have all data? if yes, return it
	if len(chunk.conns) == len(a.sockets) {
		return chunk.data, nil
	} else {
		return nil, nil
	}
}

func (a *Socket) formatRead(conn *ConnObserver, data []byte) ([]byte, error) {
	a.lock.Lock()
	defer a.lock.Unlock()

	if len(a.sockets) == 1 {

		if a.Debug {
			log.Printf("adding %d\n", len(data))
		}

		a.flightLock.Lock()
		a.inflight = append(a.inflight, data...)

		a.flightCond.Signal()
		a.flightLock.Unlock()

		if a.Debug {
			log.Println("signalled")
		}

		return []byte{}, nil
	} else {
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

				data, err := a.receivedChunk(conn, &pkt)
				if err != nil {
					panic(err) // TODO
				}

				if a.Debug {
					log.Printf("read pkt [sn: %d, %d] - %v %v\n", pkt.Sn, len(pkt.Data), data, err)
				}

				if data != nil {
					a.flightLock.Lock()
					a.inflight = append(a.inflight, data...)
					a.flightLock.Unlock()

					if a.Debug {
						log.Printf("appended %d\n", len(data))
					}

					a.flightCond.Signal()
				}

				pkt.Reset()
			}
		}

		return orig[nRead:], nil
	}
}

func (a *Socket) handleClose(conn *ConnObserver, fd net.Conn) {
	a.lock.Lock()
	defer a.lock.Unlock()

	panic("todo")
}

func (addr SocketAddr) Network() string {
	return addr.network
}

func (addr SocketAddr) String() string {
	return addr.str
}

func (a *Socket) DropSocket(fd net.Conn) error {
	panic("todo")
}

func (a *Socket) AddSocket(fd net.Conn) {
	a.sockets = append(a.sockets,
		NewConnObserver(fd, 128*1024, a.formatWrite, a.formatRead, a.handleClose))
}

func (a *Socket) Read(b []byte) (n int, err error) {
	if a.Debug {
		log.Println("enter")
		defer log.Println("exit")
	}

	//for {
	//	a.fragLock.Lock()
	//	if len(a.fragments) > 0 {
	//		//a.fragLock.Unlock()
	//		break
	//	}

	//	a.fragLock.Unlock()
	//	time.Sleep(10 * time.Millisecond)

	//}

	a.flightLock.Lock()
	for len(a.inflight) == 0 {
		a.flightCond.Wait()
	}

	// fragLock locked, fragments has data
	readLen := len(a.inflight)
	if readLen > len(b) {
		readLen = len(b)
	}

	copy(b, a.inflight[:readLen])
	if readLen == len(a.inflight) {
		a.inflight = make([]byte, 0)
	} else {
		a.inflight = a.inflight[readLen:]
	}

	// ..
	a.flightLock.Unlock()
	return readLen, nil
}

func (a *Socket) Write(b []byte) (n int, err error) {

	if a.Debug {
		log.Println("trying to write..")
	}

	for i, fd := range a.sockets {
		if a.Debug {
			log.Printf("> %d\n", i)
		}
		fd.Write(b)
	}

	a.tsn += 1

	if a.Debug {
		log.Printf("write %d\n", len(b))
	}

	return len(b), nil
}

func (a *Socket) Close() error {
	//if !a.closed {
	//	a.closed = true
	//	close(a.writeCh)
	//	close(a.closeAbsorb)

	//	for _, sock := range a.sockets {
	//		err := sock.Close()
	//		if err != nil {
	//			return err
	//		}
	//	}
	//	return nil
	//}
	return nil
}

func (a *Socket) LocalAddr() net.Addr {
	return a.sockets[0].LocalAddr()
}

func (a *Socket) RemoteAddr() net.Addr {
	return a.sockets[0].RemoteAddr()
}

func (a *Socket) SetDeadline(t time.Time) error {
	return nil
}

func (a *Socket) SetReadDeadline(t time.Time) error {
	return nil
}

func (a *Socket) SetWriteDeadline(t time.Time) error {
	return nil
}
