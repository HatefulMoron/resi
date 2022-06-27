package internal

import (
	"encoding/hex"
	"net"
	"time"

	"github.com/nknorg/ncp-go"
	"github.com/nknorg/nkn-sdk-go"
)

type NKNListener struct {
	account *nkn.Account
	mc      *nkn.MultiClient
}

type NKNConn struct {
	account *nkn.Account
	mc      *nkn.MultiClient
	session net.Conn
}

func (c *NKNConn) Read(b []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return c.session.Read(b)
}

func (c *NKNConn) Write(b []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return c.session.Write(b)
}

func (c *NKNConn) Close() error {
	return c.session.Close()
}

func (c *NKNConn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

func (c *NKNConn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

func (c *NKNConn) SetDeadline(t time.Time) error {
	return c.session.SetDeadline(t)
}

func (c *NKNConn) SetReadDeadline(t time.Time) error {
	return c.session.SetReadDeadline(t)
}

func (c *NKNConn) SetWriteDeadline(t time.Time) error {
	return c.session.SetWriteDeadline(t)
}

func NewNKNListener(seed string) (*NKNListener, error) {
	ncpConfig := &ncp.Config{
		NonStream: false,
	}
	config := &nkn.ClientConfig{
		//RPCConcurrency: 8,
		SessionConfig:  ncpConfig,
		ConnectRetries: 1,
	}

	seedHex, err := hex.DecodeString(seed)
	if err != nil {
		return nil, err
	}

	account, err := nkn.NewAccount(seedHex)
	if err != nil {
		return nil, err
	}

	mc, err := nkn.NewMultiClient(account, "", 4, false, config)
	if err != nil {
		return nil, err
	}

	<-mc.OnConnect.C

	err = mc.Listen(nil)
	if err != nil {
		return nil, err
	}

	return &NKNListener{
		account: account,
		mc:      mc,
	}, nil
}

func (l *NKNListener) Accept() (net.Conn, error) {
	return l.mc.Accept()
}

func (l *NKNListener) Close() error {
	return l.mc.Close()
}

func (l *NKNListener) Addr() net.Addr {
	return l.mc.Addr()
}

func Dial(addr string) (*NKNConn, error) {

	ncpConfig := &ncp.Config{
		NonStream: false,
	}
	config := &nkn.ClientConfig{
		//RPCConcurrency: 8,
		SessionConfig:  ncpConfig,
		ConnectRetries: 1,
	}
	dialConfig := &nkn.DialConfig{
		DialTimeout: 5000,
	}

	account, err := nkn.NewAccount(nil)
	if err != nil {
		return nil, err
	}

	mc, err := nkn.NewMultiClient(account, "", 4, false, config)
	if err != nil {
		return nil, err
	}

	<-mc.OnConnect.C

	session, err := mc.DialWithConfig(addr, dialConfig)
	if err != nil {
		return nil, err
	}

	return &NKNConn{
		account: account,
		mc:      mc,
		session: session,
	}, nil
}
