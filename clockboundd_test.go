package clockboundc

import (
	"context"
	"log"
	"net"
	"os"
)

type MockDaemon struct {
	local    string
	conn     net.PacketConn
	ch       chan []byte
	closed   chan struct{}
	response []byte
}

func NewMockDaemon(ctx context.Context, response []byte) (*MockDaemon, error) {
	local, err := newSockPath()
	if err != nil {
		return nil, err
	}

	config := &net.ListenConfig{}
	conn, err := config.ListenPacket(ctx, "unixgram", local)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(local, 0777); err != nil {
		conn.Close()
		os.Remove(local)
		return nil, err
	}

	d := &MockDaemon{
		local:    local,
		conn:     conn,
		ch:       make(chan []byte, 8),
		closed:   make(chan struct{}),
		response: response,
	}
	go d.serve()

	return d, nil
}

func (d *MockDaemon) Requests() <-chan []byte {
	return d.ch
}

func (d *MockDaemon) serve() {
	for {
		select {
		case <-d.closed:
			return
		default:
		}

		buf := make([]byte, 16)
		n, raddr, err := d.conn.ReadFrom(buf)
		if err != nil {
			select {
			case <-d.closed:
				return
			default:
			}
			log.Println(err)
			continue
		}
		buf = buf[:n]
		d.ch <- buf

		if _, err := d.conn.WriteTo(d.response, raddr); err != nil {
			select {
			case <-d.closed:
				return
			default:
			}
			log.Println(err)
			continue
		}
	}
}

func (d *MockDaemon) Close() error {
	close(d.closed)
	err := d.conn.Close()
	if err1 := os.Remove(d.local); err == nil && err1 != nil {
		err = err1
	}
	return err
}
