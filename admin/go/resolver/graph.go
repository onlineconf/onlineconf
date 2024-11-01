package resolver

import (
	"context"
	"regexp"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var templateRe = regexp.MustCompile(`\$\{(.*?)\}`)

type Case struct {
	ContentType string     `json:"mime"`
	Value       NullString `json:"value"`
	Server      string     `json:"server"`
	Group       string     `json:"group"`
	Datacenter  string     `json:"datacenter"`
	Service     string     `json:"service"`
}

type caseResolver interface {
	resolveCase(context.Context, *Param) *Case
	getTemplateVar(string) (string, bool)
}

type graph struct {
	root         *Param
	caseResolver caseResolver
	resolved     int
}

func (g *graph) get(ctx context.Context, path string) *Param {
	nodePtr := &g.root
	segments := strings.Split(strings.TrimRight(path, "/"), "/")[1:]
	for _, name := range segments {
		g.resolve(ctx, nodePtr)
		if childPtr := (*nodePtr).Children[name]; childPtr != nil && *childPtr != nil {
			nodePtr = childPtr
		} else {
			return nil
		}
	}
	g.resolve(ctx, nodePtr)
	return *nodePtr
}

func (g *graph) resolve(ctx context.Context, paramPtr **Param) {
	defer func() {
		g.resolved++
		if g.resolved%1000 == 0 {
			runtime.Gosched()
		}
	}()

	param := *paramPtr
	for *paramPtr != nil {
		switch (*paramPtr).ContentType {
		case "application/x-symlink":
			g.resolveSymlink(ctx, paramPtr)
		case "application/x-case":
			g.resolveCase(ctx, *paramPtr)
		case "application/x-template":
			g.resolveTemplate(ctx, *paramPtr)
			return
		default:
			return
		}
	}
	log.Ctx(ctx).Warn().Str("param", param.Path).Str("content_type", param.ContentType).Str("value", param.Value.String).Msg("resolve failed")
}

func (g *graph) resolveSymlink(ctx context.Context, paramPtr **Param) {
	param := *paramPtr
	*paramPtr = nil
	if target := g.get(ctx, param.Value.String); target != nil {
		*paramPtr = target
		return
	}
	log.Ctx(ctx).Warn().Str("param", param.Path).Str("value", param.Value.String).Msg("symlink is pointing to nowhere")
}

func (g *graph) resolveCase(ctx context.Context, param *Param) {
	cs := g.caseResolver.resolveCase(ctx, param)
	if cs != nil {
		param.ContentType = cs.ContentType
		param.Value = cs.Value
	} else {
		param.ContentType = "application/x-null"
		param.Value = NullString{}
	}
}

func (g *graph) resolveTemplate(ctx context.Context, param *Param) {
	if param.seen {
		return
	}
	param.seen = true
	expanded := true
	param.Value.String = templateRe.ReplaceAllStringFunc(param.Value.String, func(match string) string {
		name := match[2 : len(match)-1]
		if strings.HasPrefix(name, "/") {
			if target := g.get(ctx, name); target != nil {
				if target.ContentType == "application/x-template" {
					expanded = false
				}
				return target.Value.String
			} else {
				log.Ctx(ctx).Warn().Str("param", param.Path).Str("value", param.Value.String).Str("var", name).Msg("failed to expand variable")
				return ""
			}
		} else if value, ok := g.caseResolver.getTemplateVar(name); ok {
			return value
		} else {
			expanded = false
			return match
		}
	})
	if expanded {
		param.ContentType = "text/plain"
	}
	param.seen = false
}

func (g *graph) resolveChildren(ctx context.Context, paramPtr **Param) {
	if paramPtr == nil || *paramPtr == nil || (*paramPtr).deepResolved {
		return
	}
	g.resolve(ctx, paramPtr)
	node := *paramPtr
	if node == nil || node.seen || node.deepResolved {
		return
	}
	node.seen = true
	for _, childPtr := range node.Children {
		g.resolveChildren(ctx, childPtr)
	}
	node.deepResolved = true
	node.seen = false
}
