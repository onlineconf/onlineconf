package common

import (
	"fmt"
	"runtime"

	"github.com/rs/zerolog/log"
)

// RecoverPanic recovers a panic and logs it with a stack trace, so a panic in
// library code does not crash the whole process. Defer it at the top of a
// background goroutine — unlike net/http request handlers, bare goroutines have
// no recovery of their own:
//
//	go func() {
//		defer RecoverPanic("dependency rebuild")
//		...
//	}()
func RecoverPanic(what string) {
	r := recover()
	if r == nil {
		return
	}
	stack := make([]byte, 64<<10)
	stack = stack[:runtime.Stack(stack, false)]
	log.Error().
		Str("goroutine", what).
		Str("panic", fmt.Sprint(r)).
		Str("stack", string(stack)).
		Msg("recovered panic in background goroutine")
}
