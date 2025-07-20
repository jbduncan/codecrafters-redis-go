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
	close           func()
	quit            chan struct{}
	events          chan event
	router          *Router
}

// TODO: consider merging these two parameters together into an interface to
//       make their relationship more obvious.

func NewDispatcher(tcpConnAccepter TCPConnAccepter, close func()) *Dispatcher {
	return &Dispatcher{
		tcpConnAccepter: tcpConnAccepter,
		close:           close,
		quit:            make(chan struct{}),
		events:          make(chan event, 512),
		router:          NewRouter(),
	}
}

func (d *Dispatcher) Run() {
	go func() {
		for {
			tcpConn, err := d.tcpConnAccepter()

			if err != nil {
				select {
				case <-d.quit:
					return
				default:
					logError(err)
				}
				continue
			}

			go d.handleConn(tcpConn)
		}
	}()

	for e := range d.events {
		if err := NewRESP2Writer(e.conn).Write(e.value); err != nil {
			logError(err)
		}
	}
}

func (d *Dispatcher) handleConn(tcpConn TCPConn) {
	for value, err := range NewRESP2Scanner(tcpConn).ScanAll() {
		var errorValue ErrorValue
		if errors.As(err, &errorValue) {
			d.events <- event{
				value: errorValue,
				conn:  tcpConn,
			}
			return
		}
		if err != nil {
			logError(err)
			return
		}

		result := d.router.Route(value)
		d.events <- event{
			value: result,
			conn:  tcpConn,
		}
	}
}

func (d *Dispatcher) Stop() {
	// Stops more TCP connections from being accepted and makes Run() wait for
	// d.events to be drained before terminating.
	close(d.quit)
	close(d.events)
	d.close()
}
