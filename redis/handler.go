package redis

import (
	"bufio"
	"errors"
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
		logError(err)
		return
	}
	defer closeAndLogError(tcpConn)

	connReader := bufio.NewReader(tcpConn)
	for {
		nextToken, err := connReader.ReadString('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logError(err)
			}
			return
		}

		if nextToken != "PING\r\n" && nextToken != "ping\r\n" {
			continue
		}

		if _, err := io.WriteString(tcpConn, "+PONG\r\n"); err != nil {
			logError(err)
			return
		}
	}
}
