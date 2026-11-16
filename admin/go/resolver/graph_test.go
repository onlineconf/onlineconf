//go:build !integration

package resolver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// mk builds a tree node as selectTree would deliver it: the value already
// plaintext, the marker as stored.
func mk(path, contentType, value string, encrypted bool, children ...*Param) *Param {
	p := &Param{
		Path:        path,
		Name:        path[strings.LastIndex(path, "/")+1:],
		ContentType: contentType,
		Encrypted:   encrypted,
		Children:    map[string]**Param{},
	}
	if contentType != ContentTypeNull {
		p.Value = NullStringOf(value)
	}
	for _, c := range children {
		child := c
		p.Children[child.Name] = &child
	}
	return p
}

// recorder is a Synchronized target that keeps what it was delivered.
type recorder struct {
	param  *Param
	called bool
}

func (r *recorder) Update(_ context.Context, param *Param) {
	r.param, r.called = param, true
}

// The common graph belongs to no server: a case at a synchronized path has
// no branch to select and resolves to absent, a template with a host
// variable stays unexpanded — neither panics the reload.
func TestCommonGraphCaseAndHostVariable(t *testing.T) {
	root := mk("/", ContentTypeNull, "", false,
		mk("/onlineconf", ContentTypeNull, "", false,
			mk("/onlineconf/case", ContentTypeCase, `[{"mime":"text/plain","value":"x"}]`, false),
			mk("/onlineconf/tmpl", ContentTypeTemplate, "${hostname}", false),
		),
	)
	ctx := context.Background()
	r := &recorder{}
	newCommonGraph(root.deepClone()).synchronize(ctx, "/onlineconf/case", r)
	if !r.called || r.param == nil || r.param.ContentType != ContentTypeNull || r.param.Value.Valid {
		t.Errorf("a case was not delivered as absent: called=%v param=%+v", r.called, r.param)
	}
	r = &recorder{}
	newCommonGraph(root.deepClone()).synchronize(ctx, "/onlineconf/tmpl", r)
	if !r.called || r.param == nil || r.param.ContentType != ContentTypeTemplate || r.param.Value.String != "${hostname}" {
		t.Errorf("a host-variable template was not delivered unexpanded: called=%v param=%+v", r.called, r.param)
	}
}

// fakeCaseResolver selects the default branch of a case and knows no host
// variable, which is all a taint test needs of a server.
type fakeCaseResolver struct{}

func (fakeCaseResolver) resolveCase(_ context.Context, param *Param) *Case {
	var branches []Case
	if err := json.Unmarshal([]byte(param.Value.String), &branches); err != nil {
		return nil
	}
	for i := range branches {
		b := &branches[i]
		if b.Server == "" && b.Group == "" && b.Datacenter == "" && b.Service == "" {
			return b
		}
	}
	return nil
}

func (fakeCaseResolver) getTemplateVar(context.Context, *Param, string) (string, bool) {
	return "", false
}

// A secret keeps its marker through resolution: a case whose selected branch
// is a symlink to an encrypted parameter resolves to that parameter, a
// template expanding one carries the taint, and a log statement prints
// <encrypted> for both and the value otherwise.
func TestSecretTaint(t *testing.T) {
	root := mk("/", ContentTypeNull, "", false,
		mk("/secret", ContentTypeText, "s3cr3t", true),
		mk("/plain", ContentTypeText, "shown", false),
		mk("/case", ContentTypeCase, `[{"server":"other*","mime":"text/plain","value":"x"},{"mime":"application/x-symlink","value":"/secret"}]`, false),
		mk("/tmpl", ContentTypeTemplate, "user:${/secret}@${/plain}", false),
	)
	g := &graph{root: root, caseResolver: fakeCaseResolver{}}
	ctx := context.Background()
	if p := g.get(ctx, "/case"); p == nil || !p.Encrypted || p.Value.String != "s3cr3t" || p.logValue() != "<encrypted>" {
		t.Errorf("case with a symlink branch to a secret: %+v", p)
	}
	if p := g.get(ctx, "/tmpl"); p == nil || !p.Encrypted || p.ContentType != ContentTypeText || p.Value.String != "user:s3cr3t@shown" || p.logValue() != "<encrypted>" {
		t.Errorf("template expanding a secret: %+v", p)
	}
	if p := g.get(ctx, "/plain"); p == nil || p.Encrypted || p.logValue() != "shown" {
		t.Errorf("plain parameter: %+v", p)
	}
}
