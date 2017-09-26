package cache

import (
    "time"
    "errors"
)

type object struct {
    time time.Time
    timeout time.Duration
    obj interface{}
}

type cachetable map[string]*object

var (
    cache cachetable
    ErrTimeOut error = errors.New("The cache has been timeout.")
    ErrKeyNotFound error = errors.New("The key was not found.")
    ErrTypeAssertion error = errors.New("Type assertion error.")
)

func init() {
    cache = make(cachetable, 1000)
    go gc()
}

func gc() {
    for {
        for k, v := range cache {
            if v.time.Add(v.timeout).Before(time.Now()) {
                delete(cache, k)
            }
            time.Sleep(time.Microsecond)
        }
        time.Sleep(time.Second)
    }
}

func Set(key string, obj interface{}, timeout time.Duration) {
    cache[key] = &object{time.Now(), timeout, obj}
}

func Get(key string) (obj interface{}, err error) {
    c, ok := cache[key]
    if ok {
        now := time.Now()
        if c.time.Add(c.timeout).After(now) {
            c.time = now
            return c.obj, nil
        }
        delete(cache, key)
        return nil, ErrTimeOut
    }
    return nil, ErrKeyNotFound
}

func Delete(key string) {
    delete(cache, key)
}

func HasKey(key string) bool {
    _, ok := cache[key]
    return ok
}
