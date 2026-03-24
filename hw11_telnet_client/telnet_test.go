package main

import (
	"bytes"
	"io"
	"net"
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

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
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
}

func TestTelnetClient_ConnectionError(t *testing.T) {
	timeout := 2 * time.Second

	in := &bytes.Buffer{}
	out := &bytes.Buffer{}

	client := NewTelnetClient("127.0.0.1:65534", timeout, io.NopCloser(in), out)

	err := client.Connect()
	require.EqualError(t, err, "dial tcp 127.0.0.1:65534: connect: connection refused")
	require.NoError(t, client.Close())
}

func TestTelnetClient_CloseWithoutConnect(t *testing.T) {
	// Close должен быть безопасен до Connect и при повторном вызове.
	client := NewTelnetClient(
		"127.0.0.1:1",
		time.Second,
		io.NopCloser(bytes.NewBuffer(nil)),
		&bytes.Buffer{},
	)

	require.NoError(t, client.Close())
	require.NoError(t, client.Close())
}

func TestTelnetClient_SendWithoutConnect(t *testing.T) {
	// Попытка отправки без соединения должна вернуть ожидаемую ошибку.
	client := NewTelnetClient(
		"127.0.0.1:1",
		time.Second,
		io.NopCloser(bytes.NewBufferString("ping")),
		&bytes.Buffer{},
	)

	err := client.Send()
	require.ErrorIs(t, err, errNotConnected)
}

func TestTelnetClient_ReceiveWithoutConnect(t *testing.T) {
	// Попытка чтения без соединения должна вернуть ожидаемую ошибку.
	client := NewTelnetClient(
		"127.0.0.1:1",
		time.Second,
		io.NopCloser(bytes.NewBuffer(nil)),
		&bytes.Buffer{},
	)

	err := client.Receive()
	require.ErrorIs(t, err, errNotConnected)
}
