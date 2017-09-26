AutoInc
=======

`AutoInc` can generate integers from `start` with `step`. This library is safe for concurrent use.

Example
-------

```go
import (
	"fmt"
	"sync"
	// import the library
	"github.com/mikespook/golib/autoinc"
)

const (
	Start = 0
	End   = 100
	Step  = 2
)

func main() {
	var wg sync.WaitGroup
	
	// get the instance
	ai := autoinc.New(Start, Step)
	// close the counter
	defer ai.Close()
	for i := Start; i < End; i++ {
		wg.Add(1)
		go func() {
			// get an integer and print it
			fmt.Println(ai.Id())
			wg.Done()
		}()
	}
	wg.Wait()
}
```
