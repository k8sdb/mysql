package pool

import (
	"container/list"
	"sync"
)

type Pool struct {
	New   func() interface{}
	Close func(interface{}) error

	sync.Mutex
	data *list.List

	fast chan interface{}
	max  int
}

func New() *Pool {
	p := &Pool{
		data: list.New(),
		fast: make(chan interface{}),
	}
	return p
}

func (p *Pool) MaxLen(n int) {
	p.Lock()
	defer p.Unlock()
	p.max = n
	for p.data.Len() > p.max {
		elem := p.data.Front()
		if elem == nil {
			break
		}
		p.data.Remove(elem)
	}
}

func (p *Pool) Len() int {
	return p.data.Len()
}

func (p *Pool) Clean() {
	p.data.Init()
}

func (p *Pool) Get() (item interface{}) {
	select {
	case item = <-p.fast:
	default:
		p.Lock()
		elem := p.data.Back()
		if elem != nil {
			item = p.data.Remove(elem)
		}
		if item == nil && p.New != nil {
			item = p.New()
		}
		p.Unlock()
	}
	return
}

func (p *Pool) Put(item interface{}) (err error) {
	select {
	case p.fast <- item:
	default:
		p.Lock()
		if p.max == 0 || p.data.Len() < p.max {
			p.data.PushBack(item)
		} else if p.Close != nil {
			err = p.Close(item)
		}
		p.Unlock()
	}
	return err
}
