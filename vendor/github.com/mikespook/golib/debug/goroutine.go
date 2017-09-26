package debug

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func PrintGoroutines(allFrame bool) {
	FprintGoroutines(os.Stderr, allFrame)
}

func FprintGoroutines(w io.Writer, allFrame bool) {
	ng := runtime.NumGoroutine()
	p := make([]runtime.StackRecord, ng)
	n, ok := runtime.GoroutineProfile(p)
	if !ok {
		panic("The slice is too small too put all records")
		return
	}
	for i := 0; i < n; i++ {
		printStackRecord(w, i, p[i].Stack(), allFrame)
	}
}

// stolen from `runtime/pprof`
func printStackRecord(w io.Writer, index int, stk []uintptr, allFrame bool) {
	wasPanic := false
	for i, pc := range stk {
		f := runtime.FuncForPC(pc)
		if f == nil {
			fmt.Fprintf(w, "#\t%#x\n", pc)
			wasPanic = false
		} else {
			tracepc := pc
			// Back up to call instruction.
			if i > 0 && pc > f.Entry() && !wasPanic {
				if runtime.GOARCH == "386" || runtime.GOARCH == "amd64" {
					tracepc--
				} else {
					tracepc -= 4 // arm, etc
				}
			}
			file, line := f.FileLine(tracepc)
			name := f.Name()
			wasPanic = name == "runtime.panic"
			if name == "runtime.goexit" || !allFrame && (strings.HasPrefix(name, "runtime.") || name == "bgsweep" || name == "runfinq" || name == "main.printGoroutines") {
				continue
			}
			fmt.Fprintf(w, "%d.\t%#x\t%s+%#x\t%s:%d\n", index, pc, name, pc-f.Entry(), file, line)
		}
	}
}
