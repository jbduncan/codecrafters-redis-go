package redis

import (
	"errors"
	"io"
	"sync"
)

type TCPConn io.ReadWriteCloser

type TCPConnAccepter interface {
	Accept() (TCPConn, error)
	Close()
}

type event struct {
	value Value
	conn  TCPConn
}

type Dispatcher struct {
	tcpConnAccepter TCPConnAccepter
	router          *Router
	events          chan event
	quit            chan struct{}
	wg              *sync.WaitGroup
	stopOnce        *sync.Once
}

func NewDispatcher(tcpConnAccepter TCPConnAccepter) *Dispatcher {
	return &Dispatcher{
		tcpConnAccepter: tcpConnAccepter,
		router:          NewRouter(),
		events:          make(chan event, 512),
		quit:            make(chan struct{}),
		wg:              new(sync.WaitGroup),
		stopOnce:        new(sync.Once),
	}
}

func (d *Dispatcher) Run() {
	go func() {
		for {
			tcpConn, err := d.tcpConnAccepter.Accept()

			if err != nil {
				select {
				case <-d.quit:
					return
				default:
					logError(err)
				}
				continue
			}

			d.wg.Add(1)
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
	defer d.wg.Done()

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
	// Stops more TCP connections from being accepted and allows Run() to drain
	// all remaining events in d.events before terminating.
	d.stopOnce.Do(func() {
		close(d.quit)
		d.tcpConnAccepter.Close()
		d.wg.Wait()
	})
}
