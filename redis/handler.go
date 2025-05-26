package redis

import (
	"bufio"
	"bytes"
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
	// TODO: Refactor to use a multi-producer, single-consumer architecture
	//       like an event loop.
	for {
		tcpConn, err := h.tcpConnAccepter()
		if err != nil {
			logError(err)
			return
		}

		go h.handleConn(tcpConn)
	}
}

var (
	upperPing = []byte("PING\r\n")
	lowerPing = []byte("ping\r\n")
	pong      = []byte("+PONG\r\n")
)

func (h *Handler) handleConn(tcpConn TCPConn) {
	defer closeAndLogError(tcpConn)

	connReader := bufio.NewReader(tcpConn)
	for {
		nextToken, err := connReader.ReadSlice('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logError(err)
			}
			return
		}

		if !bytes.Equal(nextToken, upperPing) &&
			!bytes.Equal(nextToken, lowerPing) {
			continue
		}

		if _, err := tcpConn.Write(pong); err != nil {
			logError(err)
			return
		}
	}
}
