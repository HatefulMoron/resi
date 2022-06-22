package main

import (
	"crypto/sha256"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
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

type Socket struct {
	sockets     []net.Conn
	lock        sync.Mutex
	rdLock      sync.Mutex
	sn          uint64
	rsn         uint64
	writeCh     chan WriteOperation
	closed      bool
	closeAbsorb chan struct{}
	absorbed    bool
	Debug       bool
}

func NewSocket(sockets []net.Conn) *Socket {
	sock := &Socket{
		sockets:     sockets,
		sn:          0,
		rsn:         0,
		writeCh:     make(chan WriteOperation),
		closed:      false,
		closeAbsorb: make(chan struct{}),
		absorbed:    false,
		Debug:       false,
	}

	go sock.writeRoutine()
	return sock
}

func (a *Socket) Absorb(b *Socket) {

	if a.sn != b.sn {
		log.Printf("a: %d\nb: %d\n", a.sn, b.sn)
		panic("incorrect usage")
	}

	for _, fd := range b.sockets {
		a.sockets = append(a.sockets, fd)
	}

	b.absorbed = true
}

func (addr SocketAddr) Network() string {
	return addr.network
}

func (addr SocketAddr) String() string {
	return addr.str
}

func (a *Socket) DropSocket(fd net.Conn) error {
	a.lock.Lock()
	defer a.lock.Unlock()

	for i, sock := range a.sockets {
		if sock == fd {
			a.sockets[i] = a.sockets[len(a.sockets)-1]
			a.sockets = a.sockets[:len(a.sockets)-1]

			err := sock.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Socket) AddSocket(fd net.Conn) {
	a.sockets = append(a.sockets, fd)
}

func (a *Socket) Read(b []byte) (n int, err error) {

	a.rdLock.Lock()
	defer a.rdLock.Unlock()

	if a.Debug {
		log.Println("=== starting read")
	}

	if a.absorbed {
		_, _ = <-a.closeAbsorb
		return 0, errors.New("closed")
	}

	var wg sync.WaitGroup
	wg.Add(len(a.sockets))

	datas := make([]*ReadStatus, len(a.sockets))

	for i, fd := range a.sockets {
		lI := i
		lFd := fd
		go func() {

			// While we don't have a full packet, read the fd
			buf := make([]byte, maxPacketSize)
			ptr := 0
			defer func() {
				wg.Done()
			}()

			for {
				if a.Debug {
					log.Printf("inner %s: waiting\n", lFd.LocalAddr().String())
				}
				n, err := lFd.Read(buf[ptr:])

				if a.Debug {
					log.Printf("inner %s: %d %v\n", lFd.LocalAddr().String(), n, err)
				}

				if err != nil {

					if a.Debug {
						log.Printf("dropping: %v\n", err)
					}

					a.DropSocket(lFd) // TODO
					break
				} else if n == 0 && err == nil {
					break
				}

				if len(a.sockets) > 1 {

					var pkt resi.Packet
					err = proto.Unmarshal(buf[:ptr+n], &pkt)
					if err == nil && n == 0 {
						break
					} else if err == nil {
						datas[lI] = &ReadStatus{
							fd:  lFd,
							pkt: &pkt,
						}
						break
					}
				} else {

					datas[lI] = &ReadStatus{
						fd: lFd,
						pkt: &resi.Packet{
							Sn:   0,
							Data: buf[:ptr+n],
						},
					}
					break

				}

				ptr += n
			}
		}()
	}

	wg.Wait()

	var checksum *[32]byte
	checksum = nil

	for _, v := range datas {
		if v != nil {
			check := sha256.Sum256(v.pkt.Data)

			if checksum == nil {
				checksum = &check
				continue
			} else {
				if check != *checksum {
					if a.Debug {
						log.Println("read: 0 incorrectdata")
					}
					return 0, errors.New("mirror: incorrect data")
				}
			}
		}
	}

	if checksum == nil {
		if len(a.sockets) == 0 {
			if a.Debug {
				log.Println("read: 0 closed")
			}
			return 0, errSesisonClosed
		} else {
			if a.Debug {
				log.Println("read: 0 nil")
			}
			return 0, nil
		}
	}

	copy(b, datas[0].pkt.Data)

	if a.Debug {
		log.Printf("%s: read: %d nil\n", a.sockets[0].LocalAddr().String(), len(datas[0].pkt.Data))
	}

	return len(datas[0].pkt.Data), nil
}

func (a *Socket) writeRoutine() {
	for {
		op, ok := <-a.writeCh
		if !ok {
			return
		}

		n := len(a.sockets)

		if a.Debug {
			log.Printf("<<< %d\n", len(a.writeCh))
			log.Printf("doing write with %d\n", n)
		}

		var wg sync.WaitGroup
		wg.Add(n)

		var wData []byte
		var err error

		if len(a.sockets) > 1 {

			sn := a.sn
			a.sn = a.sn + 1

			wData, err = proto.Marshal(&resi.Packet{
				Sn:   sn,
				Data: op.data,
			})
			if err != nil {
				op.status <- WriteStatus{
					n:   0,
					err: err,
				}
				continue
			}

		} else {

			wData = op.data

		}

		// Try to write to all sockets
		for _, fd := range a.sockets {
			lfd := fd
			go func() {
				defer wg.Done()

				ptr := 0
				for {
					if ptr == len(wData) {
						break
					}

					n, err := lfd.Write(wData[ptr:])
					if err != nil {
						a.Close()
						return
					}

					ptr += n
				}

			}()
		}

		if a.Debug {
			log.Println("waiting on write")
		}
		wg.Wait()
		if a.Debug {
			log.Println("write done")
		}

		op.status <- WriteStatus{
			n:   len(op.data),
			err: nil,
		}
		if a.Debug {
			log.Println("write status sent")
		}
	}
}

func (a *Socket) Write(b []byte) (n int, err error) {
	if a.Debug {
		log.Printf("starting write for %d\n", len(b))
	}

	size := len(b)
	if size > maxPacketSize {
		size = maxPacketSize
	}

	resp := make(chan WriteStatus)
	a.writeCh <- WriteOperation{
		data:   b[:size],
		status: resp,
	}

	status := <-resp

	if a.Debug {
		log.Printf("write: %d %v\n", status.n, status.err)
	}

	time.Sleep(time.Duration(50) * time.Millisecond)

	return status.n, status.err
}

func (a *Socket) ReallyClose() error {
	if !a.closed {
		a.closed = true
		close(a.writeCh)
		close(a.closeAbsorb)

		for _, sock := range a.sockets {
			err := sock.Close()
			if err != nil {
				return err
			}
		}
		return nil
	}
	return nil
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
