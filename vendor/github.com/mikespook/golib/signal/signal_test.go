package signal

import (
	"os"
	"testing"
	"time"
)

func TestSignalHandler(t *testing.T) {
	Bind(os.Interrupt, func() uint { return BreakExit })
	go func() {
		time.Sleep(time.Millisecond * 100)
		if err := Send(os.Getpid(), os.Interrupt); err != nil {
			t.Error(err)
		}
	}()
	s := Wait()
	t.Log(s)
}

func TestKillHandler(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Log(err)
		}
	}()
	Bind(os.Kill, func() uint { return BreakExit })
	t.Fatal("`SIGKILL` did not handle correctly")
	go func() {
		time.Sleep(time.Millisecond * 100)
		if err := Send(os.Getpid(), os.Interrupt); err != nil {
			t.Error(err)
		}
	}()
	s := Wait()
	t.Log(s)
}
