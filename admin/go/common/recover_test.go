//go:build !integration

package common

import "testing"

// RecoverPanic must stop a panic from propagating out of the goroutine it
// guards (it logs a recovered-panic line, which is expected test output).
func TestRecoverPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic escaped RecoverPanic: %v", r)
		}
	}()
	func() {
		defer RecoverPanic("test")
		panic("boom")
	}()
	// reached only if the panic was recovered
}
