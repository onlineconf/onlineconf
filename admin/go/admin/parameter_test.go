package admin

import (
	"reflect"
	"strings"
	"testing"
)

func TestFormatSymlinkCheckQuery(t *testing.T) {
	tests := []struct {
		name         string
		in           string
		isSymlink    bool
		placeholders int
		want         []any
	}{{
		"single component path",
		"/path",
		false,
		4,
		[]any{"/path", "/path", "%${/path}%", "%${/path}%"},
	}, {
		"multiple path elements",
		"/path/to/key",
		false,
		8,
		[]any{"/path", "/path/to", "/path/to/key", "/path", "/path/to", "/path/to/key", "%${/path/to/key}%", "%${/path/to/key}%"},
	}, {
		"single component path, the value is a symlink",
		"/path",
		true,
		6,
		[]any{"/path", "/path/%", "/path", "/path/%", "%${/path}%", "%${/path}%"},
	}, {
		"multiple path elements, the value is a symlink",
		"/path/to/key",
		true,
		10,
		[]any{"/path", "/path/to", "/path/to/key", "/path/to/key/%", "/path", "/path/to", "/path/to/key", "/path/to/key/%", "%${/path/to/key}%", "%${/path/to/key}%"},
	}}

	for _, tt := range tests {
		query, params := formatSymlinkCheckQuery(tt.in, tt.isSymlink)
		placeholders := strings.Count(query, "?")

		if placeholders != tt.placeholders {
			t.Errorf("%s: got %d placeholders, want %d", tt.name, placeholders, tt.placeholders)
		}

		if !reflect.DeepEqual(params, tt.want) {
			t.Errorf("%s: query params: got %v, want %v", tt.name, params, tt.want)
		}
	}
}
