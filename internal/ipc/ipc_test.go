package ipc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"testing"
)

func fakeDaemon(started, done chan struct{}, expected Response) {
	defer func() { done <- struct{}{} }()

	os.Remove(SocketAddr)
	l, err := net.Listen("unix", SocketAddr)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	started <- struct{}{}

	log.Println("listen..")
	conn, err := l.Accept()
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	b := make([]byte, 4096)
	n, err := conn.Read(b)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return
		}
		log.Println(err)
		return
	}

	var req Request
	if err := json.Unmarshal(b[:n], &req); err != nil {
		log.Println(err)
		return
	}

	if err := json.NewEncoder(conn).Encode(expected); err != nil {
		log.Println(err)
		return
	}

}

func TestPs(t *testing.T) {
	t.Run("DialFails", func(t *testing.T) {
		if _, err := Ps(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("RspWithErr", func(t *testing.T) {
		started := make(chan struct{})
		done := make(chan struct{})

		const expectedErr = "bad thing happened"
		go fakeDaemon(started, done, Response{Err: expectedErr})

		<-started
		defer func() { <-done }()

		if _, err := Ps(); err.Error() != expectedErr {
			t.Fatalf("expected '%s', have '%s'", err.Error(), expectedErr)
		}
	})

	t.Run("RspWithWrongDataType", func(t *testing.T) {
		started := make(chan struct{})
		done := make(chan struct{})

		go fakeDaemon(started, done, Response{Data: "dummy"})

		<-started
		defer func() { <-done }()

		if _, err := Ps(); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("Ok", func(t *testing.T) {
		started := make(chan struct{})
		done := make(chan struct{})

		expected := []PsResult{
			{Name: "a", Status: "x"},
			{Name: "b", Status: "y"},
		}

		go fakeDaemon(started, done, Response{Data: expected})

		<-started
		defer func() { <-done }()

		result, err := Ps()
		if err != nil {
			t.Fatal(err)
		}

		success := func() error {
			if len(result) != len(expected) {
				return fmt.Errorf("different len: %d %d", len(result), len(expected))
			}
			for i := range result {
				if result[i] != expected[i] {
					return fmt.Errorf("different element at index %d: %s %s", i, result[i], expected[i])
				}
			}
			return nil
		}

		if err := success(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestStop(t *testing.T) {
	started := make(chan struct{})
	done := make(chan struct{})

	go fakeDaemon(started, done, Response{})

	<-started
	defer func() { <-done }()

	if err := Stop(); err != nil {
		t.Fatal(err)
	}
}

func TestStart(t *testing.T) {
	started := make(chan struct{})
	done := make(chan struct{})

	go fakeDaemon(started, done, Response{})

	<-started
	defer func() { <-done }()

	if err := Start(); err != nil {
		t.Fatal(err)
	}
}

func TestOp(t *testing.T) {
	tests := []struct {
		op   Op
		name string
	}{
		{OpPs, "ps"},
		{OpStart, "start"},
		{OpStop, "stop"},
		{Op(255), "unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if v := test.op.String(); v != test.name {
				t.Fatalf("expected %s, have %s", test.name, v)
			}
		})
	}
}
