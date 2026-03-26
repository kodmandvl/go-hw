package main

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

var errNotConnected = errors.New("telnet client is not connected")

// TelnetClient - интерфейс.
type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

// telnetclient - структура, хранит состояние подключения и потоков ввода/вывода.
type telnetclient struct {
	address   string
	timeout   time.Duration
	in        io.ReadCloser
	out       io.Writer
	conn      net.Conn
	closeOnce sync.Once
	closeErr  error
}

// NewTelnetClient создает новый экземпляр клиента для указанного адреса.
func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetclient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

// Connect устанавливает TCP-соединение с таймаутом.
func (c *telnetclient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// Close закрывает сетевое соединение; повторные вызовы безопасны.
func (c *telnetclient) Close() error {
	c.closeOnce.Do(func() {
		if c.conn != nil {
			c.closeErr = c.conn.Close()
		}
	})
	return c.closeErr
}

// Send передает данные из input-потока в сетевое соединение.
func (c *telnetclient) Send() error {
	if c.conn == nil {
		return errNotConnected
	}
	_, err := io.Copy(c.conn, c.in)
	return err
}

// Receive читает данные из сети и пишет их в output-поток.
func (c *telnetclient) Receive() error {
	if c.conn == nil {
		return errNotConnected
	}
	_, err := io.Copy(c.out, c.conn)
	return err
}
