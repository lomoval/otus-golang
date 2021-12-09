package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, ioutil.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("incorrect host", func(t *testing.T) {
		c := NewTelnetClient("nonexisthost.ru", 10*time.Second, os.Stdin, os.Stdout)
		require.Error(t, c.Connect())
	})

	t.Run("non connected port", func(t *testing.T) {
		c := NewTelnetClient("ya.ru:999", 1*time.Second, os.Stdin, os.Stdout)
		require.Error(t, c.Connect())
	})

	t.Run("timeout check", func(t *testing.T) {
		c := NewTelnetClient("ya.ru:999", 500*time.Millisecond, os.Stdin, os.Stdout)
		err := c.Connect()
		require.Eventually(t, func() bool { return err != nil }, 550*time.Millisecond, 100)
	})

	t.Run("get bad request", func(t *testing.T) {
		expected := "HTTP/1.1 400 Bad Request\r\n"

		in := &bytes.Buffer{}
		out := &bytes.Buffer{}

		c := NewTelnetClient(net.JoinHostPort("httpbin.org", "80"), 500*time.Millisecond, io.NopCloser(in), out)
		require.NoError(t, c.Connect())

		in.WriteString("get me bad request")
		require.Nil(t, c.Send())

		c.Receive()
		n, err := out.ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, expected, n)
	})
}
