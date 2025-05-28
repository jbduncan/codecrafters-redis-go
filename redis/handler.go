package redis

import (
	"bufio"
	"bytes"
	"encoding"
	"errors"
	"io"
)

type TCPConn io.ReadWriteCloser

type TCPConnAccepter func() (TCPConn, error)

var (
	upperPing = []byte("PING\r\n")
	lowerPing = []byte("ping\r\n")
	pong      = []byte("+PONG\r\n")
)

type pingCommand struct{}

func (p pingCommand) MarshalText() ([]byte, error) {
	return pong, nil
}

type command encoding.TextMarshaler

type event struct {
	cmd  command
	conn TCPConn
}

type Handler struct {
	tcpConnAccepter TCPConnAccepter
	events          chan event
}

func NewHandler(tcpConnAccepter TCPConnAccepter) *Handler {
	return &Handler{
		tcpConnAccepter: tcpConnAccepter,
		events:          make(chan event, 512),
	}
}

func (h *Handler) Handle() {
	go func() {
		for {
			tcpConn, err := h.tcpConnAccepter()
			if err != nil {
				logError(err)
				return
			}

			go h.handleConn(tcpConn)
		}
	}()

	for e := range h.events {
		text, err := e.cmd.MarshalText()
		if err != nil {
			// TODO: ERR response
		}
		if _, err := e.conn.Write(text); err != nil {
			logError(err)
		}
	}
}

func (h *Handler) handleConn(tcpConn TCPConn) {
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

		h.events <- event{
			cmd:  pingCommand{},
			conn: tcpConn,
		}
	}
}
