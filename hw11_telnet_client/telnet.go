package main

import (
	"io"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	in      io.ReadCloser
	out     io.Writer
	dialer  net.Dialer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{address: address, in: in, out: out, dialer: net.Dialer{Timeout: timeout}}
}

func (t *telnetClient) Connect() error {
	var err error
	t.conn, err = t.dialer.Dial("tcp", t.address)
	if err != nil {
		return err
	}

	return nil
}

func (t *telnetClient) Close() error {
	return t.conn.Close()
}

func (t *telnetClient) Send() error {
	_, err := io.Copy(t.conn, t.in)
	return err
}

func (t *telnetClient) Receive() error {
	_, err := io.Copy(t.out, t.conn)
	return err
}
