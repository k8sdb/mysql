package log

import (
	"errors"
	"os"
	"testing"
)

var (
	flags = map[string]int{
		"all":           LogAll,
		"debug|message": LogNone | LogDebug | LogMessage,
		"none":          LogNone,
	}
)

func TestFlag(t *testing.T) {
	f1 := LogAll ^ DisableDebug
	if f1&DisableDebug != 0 {
		t.Error("Flags of log was broken.")
	}
	for flag, level := range flags {
		if StrToLevel(flag) != level {
			t.Errorf("Flag %s needed, but %d got.", flag, level)
		}
	}
}

func TestNewLog(t *testing.T) {
	defer os.Remove("testing.log")

	_, err := NewLog("", LogAll, DefaultCallDepth)
	if err != nil {
		t.Error(err)
	}

	_, err = NewLog("testing.log", LogAll, DefaultCallDepth)
	if err != nil {
		t.Error(err)
	}

	_, err = NewLog("foobar/testing.log", LogAll, DefaultCallDepth)
	if err != nil {
		t.Log(err)
	}
}

func TestLog(t *testing.T) {
	defer os.Remove("testing.log")
	err := Init("testing.log", LogAll, DefaultCallDepth)
	if err != nil {
		t.Error(err)
	}

	Error(errors.New("Test Error."))
	Warning("Test Warning.")
	Message("Test Message.")
	Debug("Test Debug.")
}

func TestDisableLog(t *testing.T) {
	defer os.Remove("testing.log")
	err := Init("testing.log", LogAll^DisableDebug^DisableWarning, DefaultCallDepth)
	if err != nil {
		t.Error(err)
	}

	Error(errors.New("Test Error."))
	Warning("Test Warning.")
	Message("Test Message.")
	Debug("Test Debug.")
}
