package signal_test

import (
	"fmt"
	"os"
	"syscall"
	"time"

	// Import the library
	"github.com/mikespook/golib/signal"
)

func main() {
	signal.Bind(syscall.SIGUSR1, func() uint {
		fmt.Println("SIGUSR1 handler #1")
		return signal.Continue
	})
	signal.Bind(syscall.SIGUSR1, func() uint {
		fmt.Println("SIGUSR1 handler #2")
		return signal.Break
	})

	// Because the previous handler returns `Break` permanently,
	// this handler will never be excuted.
	signal.Bind(syscall.SIGUSR1, func() uint {
		fmt.Println("SIGUSR1 handler #3")
		return signal.Continue
	})

	// Bind and Unbind
	handler := signal.Bind(syscall.SIGUSR1, func() uint {
		fmt.Println("SIGUSR1 handler #4")
		return signal.Continue
	})
	handler.Unbind()
	// Another alternative way is:
	// signal.Unbind(syscall.SIGUSR1, handler.Id)

	signal.Bind(syscall.SIGINT, func() uint { return signal.BreakExit })

	// Stop automatically after 2 minutes
	go func() {
		time.Sleep(time.Second * 120)
		if err := signal.Send(os.Getpid(), os.Interrupt); err != nil {
			fmt.Println(err)
		}
	}()

	// Block here
	s := signal.Wait()
	fmt.Printf("Exit by signal: %s\n", s)
}
