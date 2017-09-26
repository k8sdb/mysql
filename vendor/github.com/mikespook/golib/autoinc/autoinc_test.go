package autoinc

import (
	"testing"
)

const (
	Start = 0
	End   = 100
	Step  = 2
)

func TestAutoInc(t *testing.T) {
	ai := New(Start, Step)
	defer ai.Close()
	for i := Start; i < End; i = i + Step {
		if id := ai.Id(); id != i {
			t.Errorf("The id %d wanted, but got %d", i, id)
			return
		}
	}
}
