package ipc

import (
	"testing"
)

func TestLs(t *testing.T) {
	_, err := Ls()
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
