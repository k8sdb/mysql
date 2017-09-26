package cache

import (
    "time"
    "testing"
)

func TestCache(t *testing.T) {
    foobar := "foobar"
    Set("foobar", foobar, time.Duration(10 * time.Second))
    obj, err := Get("foobar")
    if err != nil {
        t.Error(err)
    }
    
    str, ok := obj.(string)
    if !ok {
        t.Error("Type assertions error")
    }
    if str != foobar {
        t.Error("Set/Get were not conform.")
    }
    go Set("foobar", foobar, time.Duration(1 * time.Microsecond))
    time.Sleep(20 * time.Microsecond)
    obj, err = Get("foobar")
    if err == nil {
        t.Error("Time out is not working.")
    }

    go Set("foobar", foobar, time.Duration(1 * time.Microsecond))
    time.Sleep(time.Minute)
    if HasKey("foobar") {
       t.Error("GC is not working.")
    }
}
