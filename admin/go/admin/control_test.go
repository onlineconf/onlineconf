//go:build !integration

package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTruthy(t *testing.T) {
	for value, want := range map[string]bool{"": false, "0": false, "1": true, "yes": true, " 0": true, "00": true} {
		if got := truthy(value); got != want {
			t.Errorf("truthy(%q) = %v, want %v", value, got, want)
		}
	}
}

func TestPolicyMemoMiddleware(t *testing.T) {
	var memo *policyMemo
	h := policyMemoMiddleware(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		memo, _ = req.Context().Value(policyMemoKey{}).(*policyMemo)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	if memo == nil {
		t.Errorf("the request carries no policy memo")
	}
}
