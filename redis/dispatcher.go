package redis

import (
	"errors"
	"io"
)

type TCPConn io.ReadWriteCloser

type TCPConnAccepter func() (TCPConn, error)

type event struct {
	value Value
	conn  TCPConn
}

type Dispatcher struct {
	tcpConnAccepter TCPConnAccepter
	events          chan event
}

func NewDispatcher(tcpConnAccepter TCPConnAccepter) *Dispatcher {
	return &Dispatcher{
		tcpConnAccepter: tcpConnAccepter,
		events:          make(chan event, 512),
	}
}

func (d *Dispatcher) Run() {
	go func() {
		for {
			tcpConn, err := d.tcpConnAccepter()
			if err != nil {
				logError(err)
				return
			}

			go d.handleConn(tcpConn)
		}
	}()

	for e := range d.events {
		text := "+" + e.value.(SimpleString) + "\r\n"
		if _, err := e.conn.Write([]byte(text)); err != nil {
			logError(err)
		}
	}
}

func (d *Dispatcher) handleConn(tcpConn TCPConn) {
	connScanner := NewRESP2Scanner(tcpConn)
	for _, err := range connScanner.ScanAll() {
		if errors.Is(err, io.EOF) {
			return // All input processed
		}
		var errorValue ErrorValue
		if errors.As(err, &errorValue) {
			// TODO: test this: pass as an event to d.events to be written back
			//       to the client via redis.Writer.Write
		}
		if err != nil {
			logError(err)
			return
		}

		// TODO: pass value in _ above to a handler and pass as an event to
		//       d.events to be written back to the client via
		//       redis.Writer.Write

		// TODO: consider adding a handler to support "Inline commands" when
		//       given something that isn't an array:
		//       https://redis.io/docs/latest/develop/reference/protocol-spec/#inline-commands

		d.events <- event{
			value: SimpleString("PONG"),
			conn:  tcpConn,
		}
	}
}
