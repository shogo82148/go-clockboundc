package clockboundc

import (
	"context"
	"testing"
	"time"
)

func TestNow(t *testing.T) {
	response := []byte{
		// header
		1, // version
		1, // type Now
		0, // Chrony is not synchronized
		0, // Reserved
		// EARLIEST
		0x16, 0xb9, 0x6a, 0x09, 0x7f, 0x4d, 0x06, 0x87,
		// LATEST
		0x16, 0xb9, 0x6a, 0x09, 0x7f, 0x50, 0xd7, 0xbc,
	}
	d, err := NewMockDaemon(context.Background(), response)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	c, err := NewWithPath(d.local)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	now, err := c.Now()
	if err != nil {
		t.Fatal(err)
	}

	request := <-d.Requests()
	if len(request) != 4 {
		t.Fatalf("unexpected the length of the request: want 4, got %d", len(request))
	}
	// Version
	if request[0] != 1 {
		t.Errorf("unexpected version: want 1, got %d", request[0])
	}
	// Request Type
	if request[1] != 1 {
		t.Errorf("unexpected request type: want Now(1), got %d", request[1])
	}
	// request[2] and request[3] are reserved for future use.

	if now.Header.Version != 1 {
		t.Errorf("unexpected version: want 1, got %d", now.Header.Version)
	}
	if now.Header.Type != 1 {
		t.Errorf("unexpected type: want 1, got %d", now.Header.Type)
	}
	if now.Header.Unsynchronized {
		t.Errorf("want synchronized, but not")
	}
	if got, want := now.Bound.Earliest.UnixNano(), int64(0x16b96a09_7f4d0687); got != want {
		t.Errorf("Earliest is missmatch: want %d, got %d", want, got)
	}
	if got, want := now.Bound.Latest.UnixNano(), int64(0x16b96a09_7f50d7bc); got != want {
		t.Errorf("Latest is missmatch: want %d, got %d", want, got)
	}
	if got, want := now.Time.UnixNano(), int64(0x16b96a09_7f4eef22); got != want {
		t.Errorf("Time is missmatch: want %d, got %d", want, got)
	}
}

func TestNow_CheckNotOverflow(t *testing.T) {
	response := []byte{
		// header
		1, // version
		1, // type Now
		0, // Chrony is not synchronized
		0, // Reserved

		// it's maximum timesamp ClockBound can handle.
		// EARLIEST
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		// LATEST
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	}
	d, err := NewMockDaemon(context.Background(), response)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	c, err := NewWithPath(d.local)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	now, err := c.Now()
	if err != nil {
		t.Fatal(err)
	}

	// it's maximum timesamp ClockBound can handle.
	want, err := time.Parse(time.RFC3339Nano, "2554-07-21T23:34:33.709551615Z")
	if err != nil {
		t.Fatal(err)
	}

	// (time.Time).UnixNano overflows, so we can't use it.
	if !now.Bound.Earliest.Equal(want) {
		t.Errorf("Earliest is missmatch: want %s, got %s", want, now.Bound.Earliest)
	}
	if !now.Bound.Latest.Equal(want) {
		t.Errorf("Earliest is missmatch: want %s, got %s", want, now.Bound.Latest)
	}
	if !now.Time.Equal(want) {
		t.Errorf("Earliest is missmatch: want %s, got %s", want, now.Time)
	}
}

func TestBefore(t *testing.T) {
	response := []byte{
		// header
		1, // version
		2, // type Before
		0, // Chrony is not synchronized
		0, // Reserved

		// Before Response
		1, // 1 (true)
	}
	d, err := NewMockDaemon(context.Background(), response)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	c, err := NewWithPath(d.local)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	now := time.Unix(0, 0x1234_5678_9abc_def0)
	before, err := c.Before(now)
	if err != nil {
		t.Fatal(err)
	}

	request := <-d.Requests()
	if len(request) != 12 {
		t.Fatalf("unexpected the length of the request: want 4, got %d", len(request))
	}
	// Version
	if request[0] != 1 {
		t.Errorf("unexpected version: want 1, got %d", request[0])
	}
	// Request Type
	if request[1] != 2 {
		t.Errorf("unexpected request type: want Before(2), got %d", request[1])
	}
	// request[2] and request[3] are reserved for future use.

	for i, want := range [...]byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0} {
		if request[4+i] != want {
			t.Errorf("unexpected request[%d]: want %x, got %x", 4+i, want, request[4+i])
		}
	}

	if before.Header.Version != 1 {
		t.Errorf("unexpected version: want 1, got %d", before.Header.Version)
	}
	if before.Header.Type != 2 {
		t.Errorf("unexpected type: want 2, got %d", before.Header.Type)
	}
	if before.Header.Unsynchronized {
		t.Errorf("want synchronized, but not")
	}
	if !before.Before {
		t.Error("want true, got false")
	}
}

func TestAfter(t *testing.T) {
	response := []byte{
		// header
		1, // version
		3, // type After
		0, // Chrony is not synchronized
		0, // Reserved

		// After Response
		1, // 1 (true)
	}
	d, err := NewMockDaemon(context.Background(), response)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()

	c, err := NewWithPath(d.local)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	now := time.Unix(0, 0x1234_5678_9abc_def0)
	after, err := c.After(now)
	if err != nil {
		t.Fatal(err)
	}

	request := <-d.Requests()
	if len(request) != 12 {
		t.Fatalf("unexpected the length of the request: want 4, got %d", len(request))
	}
	// Version
	if request[0] != 1 {
		t.Errorf("unexpected version: want 1, got %d", request[0])
	}
	// Request Type
	if request[1] != 3 {
		t.Errorf("unexpected request type: want After(3), got %d", request[1])
	}
	// request[2] and request[3] are reserved for future use.

	for i, want := range [...]byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0} {
		if request[4+i] != want {
			t.Errorf("unexpected request[%d]: want %x, got %x", 4+i, want, request[4+i])
		}
	}

	if after.Header.Version != 1 {
		t.Errorf("unexpected version: want 1, got %d", after.Header.Version)
	}
	if after.Header.Type != 3 {
		t.Errorf("unexpected type: want 2, got %d", after.Header.Type)
	}
	if after.Header.Unsynchronized {
		t.Errorf("want synchronized, but not")
	}
	if !after.After {
		t.Error("want true, got false")
	}
}
