package pool

import (
	"container/list"
	"testing"
)

const (
	itemsNum = 10
	maxNum   = 4
)

func TestPool(t *testing.T) {
	p := New()
	if p.Get() != nil {
		t.Error("Should be nil")
		t.Fail()
	}

	if err := p.Put(nil); err != nil {
		t.Errorf("Unexpected error: %s", err)
		t.Fail()
	}
	i := 0
	p.New = func() interface{} {
		i++
		return i
	}
	for x := 1; x <= itemsNum; x++ {
		a := p.Get()
		if b, ok := a.(int); !ok {
			t.Errorf("Wrong value: %v", a)
			t.Fail()
		} else if b != x {
			t.Errorf("Wrong number: %d, Expected: %d", b, x)
			t.Fail()
		} else {
			t.Logf("Number: %d", b)
		}
	}

	if p.Len() != 0 {
		t.Errorf("Wrong Pool size: %d, expected: %d", p.Len(), 0)
		t.Fail()
	}

	for x := 1; x <= itemsNum; x++ {
		if err := p.Put(x); err != nil {
			t.Errorf("Unexpected error: %s", err)
			t.Fail()
		}
	}

	if p.Len() != itemsNum {
		t.Errorf("Wrong Pool size: %d, expected: %d", p.Len(), itemsNum)
		t.Fail()
	}

	p.MaxLen(maxNum)

	if p.Len() != maxNum {
		t.Errorf("Wrong Pool size: %d, expected: %d", p.Len(), maxNum)
		t.Fail()
	}
}

func dump(t *testing.T, data *list.List) {
	t.Log("----dump----BGN----")
	for elem := data.Front(); elem.Next() != nil; elem = elem.Next() {
		t.Logf("%v", elem.Value)
	}
	t.Log("----dump----END----")
}
