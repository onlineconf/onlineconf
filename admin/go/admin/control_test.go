//go:build !integration

package admin

import "testing"

func TestTruthy(t *testing.T) {
	for value, want := range map[string]bool{"": false, "0": false, "1": true, "yes": true, " 0": true, "00": true} {
		if got := truthy(value); got != want {
			t.Errorf("truthy(%q) = %v, want %v", value, got, want)
		}
	}
}
