package ipc

import (
	"testing"
)

func TestPs(t *testing.T) {
	_, err := Ps()
	t.Log(err)
}

func TestStop(t *testing.T) {
	err := Stop()
	t.Log(err)
}

func TestStart(t *testing.T) {
	err := Start()
	t.Log(err)
}
