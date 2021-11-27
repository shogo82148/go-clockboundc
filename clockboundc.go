package clockboundc

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// DefaultSocketPath is the default path for connecting clockboundd.
const DefaultSocketPath = "/run/clockboundd/clockboundd.sock"

// SocketNamePrefix is the prefix for paths of temporary unixgram sockets.
const SocketNamePrefix = "clockboundc"

// CommandType is a type for requests and responses.
type CommandType uint8

const (
	CommandTypeError  CommandType = 0
	CommandTypeNow    CommandType = 1
	CommandTypeBefore CommandType = 2
	CommandTypeAfter  CommandType = 3
)

// Bound is a clock error bound.
// "true time" exists between Earliest and Latest.
type Bound struct {
	Earliest time.Time
	Latest   time.Time
}

// Header is a header clockboundd returns.
type Header struct {
	// The protocol version of this response.
	Version uint8

	// The response type.
	Type CommandType

	// Unsynchronized true if Chrony is not synchronized. It is false otherwise.
	Unsynchronized bool
}

// Now is the current system time with a clock error bound.
type Now struct {
	// Time is the current system time without the bound.
	time.Time

	// Header is the header clockboundd returns.
	Header Header

	// Bound is the clock error bound.
	// "true time" exists between Earliest and Latest.
	Bound Bound
}

// Before is the before response from clockboundd.
type Before struct {
	// Header is the header clockboundd returns.
	Header Header

	// Before is true if the requested time happened
	// before the earliest error bound of the current system time, otherwise false.
	Before bool
}

// After is the after response from clockboundd.
type After struct {
	// Header is the header clockboundd returns.
	Header Header

	// After is true if the requested time happened
	// after the latest error bound of the current system time, otherwise false.
	After bool
}

// Client is a Go implementation of clockboundc.
type Client struct {
	local string
	conn  *net.UnixConn
}

// New returns a new clockboundc that connects to DefaultSocketPath.
func New() (*Client, error) {
	return NewWithPath(DefaultSocketPath)
}

// NewWithPath returns a new clockboundc that connects to path.
func NewWithPath(path string) (*Client, error) {
	raddr, err := net.ResolveUnixAddr("unixgram", path)
	if err != nil {
		return nil, err
	}
	local, err := newSockPath()
	if err != nil {
		return nil, err
	}
	laddr, err := net.ResolveUnixAddr("unixgram", local)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unixgram", laddr, raddr)
	if err != nil {
		return nil, err
	}
	if err := os.Chmod(local, 0666); err != nil {
		return nil, err
	}
	return &Client{
		local: local,
		conn:  conn,
	}, nil
}

func newSockPath() (string, error) {
	var buf [16]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return "", err
	}
	name := SocketNamePrefix + "-" + base64.RawURLEncoding.EncodeToString(buf[:]) + ".sock"
	return filepath.Join(os.TempDir(), name), nil
}

// Now gets the current system time with the bound.
func (c *Client) Now() (Now, error) {
	var buf [20]byte
	buf[0] = 1                    // Version
	buf[1] = byte(CommandTypeNow) // Command Type: Now (1)
	buf[2] = 0                    // Reserved
	buf[3] = 0                    // Reserved
	_, err := c.conn.Write(buf[0:4])
	if err != nil {
		return Now{}, err
	}

	n, err := c.conn.Read(buf[:])
	if err != nil {
		return Now{}, err
	}
	if n != len(buf) {
		return Now{}, fmt.Errorf("clockboundc: invalid response length: %d", n)
	}

	version := buf[0]
	typ := CommandType(buf[1])
	unsyncFlag := buf[2] != 0
	earliest := binary.BigEndian.Uint64(buf[4:])
	latest := binary.BigEndian.Uint64(buf[12:])
	return Now{
		Time: fromUnixNano(latest - (latest-earliest)/2),
		Header: Header{
			Version:        version,
			Type:           typ,
			Unsynchronized: unsyncFlag,
		},
		Bound: Bound{
			Earliest: fromUnixNano(earliest),
			Latest:   fromUnixNano(latest),
		},
	}, nil
}

// Before returns whether the requested time happened
// before the earliest error bound of the current system time.
func (c *Client) Before(t time.Time) (Before, error) {
	var buf [12]byte
	buf[0] = 1                       // Version
	buf[1] = byte(CommandTypeBefore) // Command Type: Before (2)
	buf[2] = 0                       // Reserved
	buf[3] = 0                       // Reserved
	binary.BigEndian.PutUint64(buf[4:], toUnixNano(t))
	_, err := c.conn.Write(buf[:])
	if err != nil {
		return Before{}, err
	}

	_, err = c.conn.Read(buf[:5])
	if err != nil {
		return Before{}, nil
	}
	version := buf[0]
	typ := CommandType(buf[1])
	unsyncFlag := buf[2] != 0
	before := buf[4] != 0
	return Before{
		Header: Header{
			Version:        version,
			Type:           typ,
			Unsynchronized: unsyncFlag,
		},
		Before: before,
	}, nil
}

// After returns whether the requested time happened
// after the latest error bound of the current system time.
func (c *Client) After(t time.Time) (After, error) {
	var buf [12]byte
	buf[0] = 1                      // Version
	buf[1] = byte(CommandTypeAfter) // Command Type: After (3)
	buf[2] = 0                      // Reserved
	buf[3] = 0                      // Reserved
	binary.BigEndian.PutUint64(buf[4:], toUnixNano(t))
	_, err := c.conn.Write(buf[:])
	if err != nil {
		return After{}, err
	}

	_, err = c.conn.Read(buf[:5])
	if err != nil {
		return After{}, nil
	}
	version := buf[0]
	typ := CommandType(buf[1])
	unsyncFlag := buf[2] != 0
	after := buf[4] != 0
	return After{
		Header: Header{
			Version:        version,
			Type:           typ,
			Unsynchronized: unsyncFlag,
		},
		After: after,
	}, nil
}

// Close closes the client.
func (c *Client) Close() error {
	err := c.conn.Close()
	if err1 := os.Remove(c.local); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func fromUnixNano(nano uint64) time.Time {
	return time.Unix(int64(nano/1e9), int64(nano%1e9))
}

// toUnixNano is unsigned edition of (time.Time).UnixNano.
// It can handle after year 1970 from 2554.
// (time.Time).UnixNano can't handle after year 2262.
func toUnixNano(t time.Time) uint64 {
	t.UnixNano()
	sec := t.Unix()
	nano := t.Nanosecond()
	return uint64(sec)*1e9 + uint64(nano)
}
