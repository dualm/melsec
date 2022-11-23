package melsec

import (
	"encoding/binary"
	"net"
	"reflect"
	"testing"
)

func makeListener(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		t.Fatal(err)
	}

	for {
		_, err := listener.Accept()
		if err != nil {
			t.Error(err)
		}
	}
}

func TestSetCPUTimer(t *testing.T) {
	go func() {
		makeListener(t)
	}()

	n := uint16(20)

	conn, err := NewConn("localhost", "8080", SetCPUTimer(n))
	if err != nil {
		t.Fatal(err)
	}

	b := make([]byte, 2)

	binary.LittleEndian.PutUint16(b, n)

	if !reflect.DeepEqual(McMessage(b), conn.option.getCPUTimer()) {
		t.Fatalf("want %v, got %v", b, conn.option.getCPUTimer())
	}
}
