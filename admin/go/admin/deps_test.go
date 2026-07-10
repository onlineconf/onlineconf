//go:build !integration

package admin

import (
	"reflect"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

func validString(s string) NullString {
	var v NullString
	v.Valid = true
	v.String = s
	return v
}

func TestReferrerTargets(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		value       NullString
		want        []string
	}{{
		"symlink",
		"application/x-symlink",
		validString("/a/b"),
		[]string{"/a/b"},
	}, {
		"template with path vars and a context var",
		"application/x-template",
		validString("${/infra/host}:${port}/${/infra/base}"),
		[]string{"/infra/host", "/infra/base"},
	}, {
		"template without path vars",
		"application/x-template",
		validString("plain ${hostname} text"),
		nil,
	}, {
		"case with symlink, template and plain branches",
		"application/x-case",
		validString(`[
			{"mime":"application/x-symlink","value":"/by/group","group":"g"},
			{"mime":"application/x-template","value":"${/tmpl/var}","datacenter":"dc"},
			{"mime":"text/plain","value":"default"}
		]`),
		[]string{"/by/group", "/tmpl/var"},
	}, {
		"nested case",
		"application/x-case",
		validString(`[{"mime":"application/x-case","value":"[{\"mime\":\"application/x-symlink\",\"value\":\"/nested\"}]"}]`),
		[]string{"/nested"},
	}, {
		"invalid case json",
		"application/x-case",
		validString("not json"),
		nil,
	}, {
		"plain value",
		"text/plain",
		validString("/looks/like/a/path"),
		nil,
	}, {
		"null value",
		"application/x-symlink",
		NullString{},
		nil,
	}}
	for _, tt := range tests {
		if got := referrerTargets(tt.contentType, tt.value); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestCaseSymlinkTargets(t *testing.T) {
	value := validString(`[
		{"mime":"application/x-symlink","value":"/one"},
		{"mime":"application/x-template","value":"${/not/a/redirect}"},
		{"mime":"application/x-case","value":"[{\"mime\":\"application/x-symlink\",\"value\":\"/two\"}]"},
		{"mime":"text/plain","value":"/three"}
	]`)
	want := []string{"/one", "/two"}
	if got := caseSymlinkTargets(value); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDepsNeedRebuild(t *testing.T) {
	tests := []struct {
		name        string
		hasStored   bool
		storedDepth int
		configDepth int
		want        bool
	}{
		{"fresh / never built", false, 0, 10, true},
		{"unchanged", true, 10, 10, false},
		{"raised", true, 10, 12, true},
		{"lowered keeps deeper edges", true, 10, 8, false},
	}
	for _, tt := range tests {
		if got := depsNeedRebuild(tt.hasStored, tt.storedDepth, tt.configDepth); got != tt.want {
			t.Errorf("%s: depsNeedRebuild(%v, %d, %d) = %v, want %v",
				tt.name, tt.hasStored, tt.storedDepth, tt.configDepth, got, tt.want)
		}
	}
}

func TestDepsRedirectChanged(t *testing.T) {
	tests := []struct {
		oldType, newType string
		want             bool
	}{
		{"text/plain", "text/plain", false},
		{"application/x-template", "application/x-template", false},
		{"text/plain", "application/x-symlink", true},
		{"application/x-symlink", "text/plain", true},
		{"application/x-symlink", "application/x-symlink", true},
		{"application/x-case", "application/x-case", true},
	}
	for _, tt := range tests {
		if got := depsRedirectChanged(tt.oldType, tt.newType); got != tt.want {
			t.Errorf("depsRedirectChanged(%q, %q) = %v, want %v", tt.oldType, tt.newType, got, tt.want)
		}
	}
}
