package main

import (
	"testing"
)

func TestFormURI(t *testing.T) {
	cases := []struct {
		URI    string
		Result string
		Error  bool
	}{
		{
			URI:    "http://example.com/path/prefix/",
			Result: "http://example.com/path/prefix",
		},
		{
			URI:    "http://example.com/",
			Result: "http://example.com",
		},
		{
			URI:    "http://example.com",
			Result: "http://example.com",
		},
		{
			URI:   "//example.com",
			Error: true,
		},
		{
			URI:   "/client/config",
			Error: true,
		},
	}

	for i, c := range cases {
		res, err := formURI(c.URI)
		if err != nil {
			if !c.Error {
				t.Fatalf("case[%d]: %v", i, err)
			}
			continue
		}
		if c.Error {
			t.Fatalf("case[%d]: expected error", i)
		}
		if res != c.Result {
			t.Fatalf("case[%d]: result %s not equal to %s", i, res, c.Result)
		}
	}
}
