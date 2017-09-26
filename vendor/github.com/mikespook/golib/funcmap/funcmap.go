package funcmap

import (
    "errors"
    "reflect"
)

var (
    ErrParamsNotAdapted = errors.New("The number of params is not adapted.")
)

type Funcs map[string]reflect.Value

func New() Funcs {
    return make(Funcs, 2)
}

func (f Funcs) Bind(name string, fn interface{}) (err error) {
    defer func() {
        if e := recover(); e != nil {
            err = errors.New(name + " is not callable.")
        }
    }()
    v := reflect.ValueOf(fn)
    v.Type().NumIn()
    f[name] = v
    return
}

func (f Funcs) Has(name string) bool {
	_, ok := f[name]
    return ok
}

func (f Funcs) Call(name string, params ... interface{}) (result []reflect.Value, err error) {
    if _, ok := f[name]; !ok {
        err = errors.New(name + " does not exist.")
        return
    }
    if len(params) != f[name].Type().NumIn() {
        err = ErrParamsNotAdapted
        return
    }
    in := make([]reflect.Value, len(params))
    for k, param := range params {
        in[k] = reflect.ValueOf(param)
    }
    result = f[name].Call(in)
    return
}
