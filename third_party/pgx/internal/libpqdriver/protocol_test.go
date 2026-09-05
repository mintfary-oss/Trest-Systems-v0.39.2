package libpqdriver

import (
	"context"
	"database/sql/driver"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func pgmsg(c net.Conn, kind byte, payload []byte) error {
	b := make([]byte, 5+len(payload))
	b[0] = kind
	binary.BigEndian.PutUint32(b[1:5], uint32(4+len(payload)))
	copy(b[5:], payload)
	_, err := c.Write(b)
	return err
}

// A protocol fixture verifies the libpq API selection, not PostgreSQL semantics.
func TestNoParametersUseSimpleProtocol(t *testing.T) {
	l, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	defer l.Close()
	seen := make(chan string, 1)
	errs := make(chan error, 1)
	go func() {
		c, e := l.Accept()
		if e != nil {
			errs <- e
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(10 * time.Second))
		b := make([]byte, 4)
		if _, e = io.ReadFull(c, b); e != nil {
			errs <- e
			return
		}
		n := int(binary.BigEndian.Uint32(b))
		b = make([]byte, n-4)
		if _, e = io.ReadFull(c, b); e != nil {
			errs <- e
			return
		}
		pgmsg(c, 'R', []byte{0, 0, 0, 0})
		pgmsg(c, 'S', []byte("server_version\x0016.0\x00"))
		pgmsg(c, 'S', []byte("client_encoding\x00UTF8\x00"))
		pgmsg(c, 'Z', []byte{'I'})
		h := make([]byte, 5)
		if _, e = io.ReadFull(c, h); e != nil {
			errs <- e
			return
		}
		n = int(binary.BigEndian.Uint32(h[1:]))
		b = make([]byte, n-4)
		if _, e = io.ReadFull(c, b); e != nil {
			errs <- e
			return
		}
		seen <- string(h[0]) + ":" + string(b)
		pgmsg(c, 'C', []byte("CREATE TABLE\x00"))
		pgmsg(c, 'C', []byte("INSERT 0 1\x00"))
		pgmsg(c, 'Z', []byte{'I'})
		io.Copy(io.Discard, c)
	}()
	conn, e := (Driver{}).Open("host=127.0.0.1 port=" + strings.Split(l.Addr().String(), ":")[1] + " user=test dbname=test sslmode=disable connect_timeout=5")
	if e != nil {
		t.Fatal(e)
	}
	defer conn.Close()
	q := "CREATE TABLE x(id int); INSERT INTO x VALUES(1);"
	if _, e = conn.(driver.ExecerContext).ExecContext(context.Background(), q, nil); e != nil {
		t.Fatal(e)
	}
	select {
	case msg := <-seen:
		if msg != "Q:"+q+"\x00" {
			t.Fatalf("expected simple Q message, got %q", msg)
		}
	case err := <-errs:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("no protocol message")
	}
}
