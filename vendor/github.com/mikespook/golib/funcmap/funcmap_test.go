package funcmap

import (
    "testing"
)

var (
    testcases = map[string]interface{}{
        "hello": func() {print("hello")},
        "foobar": func(a, b, c int) int {return a+b+c},
        "errstring": "Can not call this as a function",
        "errnumeric": 123456789,
    }
    funcs = New()
)

func TestBind(t *testing.T) {
    for k, v := range testcases {
        err := funcs.Bind(k, v)
        if k[:3] == "err" {
            if err == nil {
                t.Error("Bind %s: %s", k, "an error should be paniced.")
            }
        } else {
            if err != nil {
                t.Error("Bind %s: %s", k, err)
            }
        }
    }
}

func TestCall(t *testing.T) {
    if _, err := funcs.Call("foobar"); err == nil {
        t.Error("Call %s: %s", "foobar", "an error should be paniced.")
    }
    if _, err := funcs.Call("foobar", 0, 1, 2); err != nil {
        t.Error("Call %s: %s", "foobar", err)
    }
    if _, err := funcs.Call("errstring", 0, 1, 2); err == nil {
        t.Error("Call %s: %s", "errstring", "an error should be paniced.")
    }
}
