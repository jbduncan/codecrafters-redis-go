package redis

import (
	"github.com/codecrafters-io/redis-starter-go/errorsx"
	"io"
)

type TCPConn io.ReadWriteCloser

type TCPConnAccepter func() (TCPConn, error)

type Handler struct {
	tcpConnAccepter TCPConnAccepter
}

func NewHandler(tcpConnAccepter TCPConnAccepter) *Handler {
	return &Handler{
		tcpConnAccepter: tcpConnAccepter,
	}
}

func (h *Handler) Handle() {
	tcpConn, err := h.tcpConnAccepter()
	if err != nil {
		errorsx.Log(err)
		return
	}
	defer errorsx.Close(tcpConn)
	if _, err := io.WriteString(tcpConn, "+PONG\r\n"); err != nil {
		errorsx.Log(err)
		return
	}
}
