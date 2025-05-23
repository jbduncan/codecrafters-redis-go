package redis

import (
	"io"
)

type TCPConn io.ReadWriteCloser

type TCPListener func() (TCPConn, error)

type Handler struct {
	tcpListener TCPListener
}

func NewHandler(tcpListener TCPListener) *Handler {
	return &Handler{
		tcpListener: tcpListener,
	}
}

func (h *Handler) Handle() {
	tcpConn, err := h.tcpListener()
	if err != nil {
		// TODO: handle error
	}
	if _, err := io.WriteString(tcpConn, "+PONG\r\n"); err != nil {
		// TODO: handle error
	}
}

func (h *Handler) Close() error {
	// TODO
	panic("implement me")
}
