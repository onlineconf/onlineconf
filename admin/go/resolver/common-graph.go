package resolver

import (
	"context"
	"errors"
	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
	"net"
	"sort"
	"strings"
)

var (
	ErrInvalidDatacenter = errors.New("Invalid datacenter configuration")
	ErrInvalidGroup      = errors.New("Invalid group configuration")
)

type datacenter struct {
	name  string
	ipnet net.IPNet
}

type group struct {
	name  string
	globs []glob.Glob
}

type commonGraph struct {
	graph
}

func newCommonGraph(root *Param) *commonGraph {
	return &commonGraph{graph: graph{root: root, caseResolver: commonCaseResolver{}}}
}

func (graph *commonGraph) readDatacenters(ctx context.Context) ([]datacenter, error) {
	dcroot := graph.get(ctx, "/onlineconf/datacenter")
	if dcroot == nil {
		return nil, nil
	}
	graph.resolveChildren(ctx, &dcroot)
	sortedNames := make([]string, 0, len(dcroot.Children))
	for name, childPtr := range dcroot.Children {
		if *childPtr != nil {
			sortedNames = append(sortedNames, name)
		}
	}
	sort.Strings(sortedNames)
	datacenters := make([]datacenter, 0, len(dcroot.Children))
	for _, name := range sortedNames {
		dc := *dcroot.Children[name]
		if !dc.Value.Valid && dc.ContentType == "text/plain" {
			return nil, ErrInvalidDatacenter
		}
		_, ipnet, err := net.ParseCIDR(dc.Value.String)
		if err != nil {
			return nil, err
		}
		datacenters = append(datacenters, datacenter{name, *ipnet})
	}
	return datacenters, nil
}

func (graph *commonGraph) readGroups(ctx context.Context) ([]group, error) {
	grroot := graph.get(ctx, "/onlineconf/group")
	if grroot == nil {
		return nil, nil
	}
	graph.resolveChildren(ctx, &grroot)
	var priority []string
	sortedGroups := make([]string, 0, len(grroot.Children))
	groups := make(map[string]group, len(grroot.Children))
	for name, grPtr := range grroot.Children {
		gr := *grPtr
		if gr == nil {
			continue
		}
		if name == "priority" {
			untrimed := strings.Split(gr.Value.String, ",")
			priority = make([]string, 0, len(untrimed))
			for _, name := range untrimed {
				priority = append(priority, strings.TrimSpace(name))
			}
		} else {
			sortedGroups = append(sortedGroups, name)
			globs := make([]glob.Glob, 0, 1+len(gr.Children))
			params := []*Param{gr}
			for len(params) > 0 {
				node := params[0]
				params = params[1:]
				if node.Value.Valid {
					g, err := glob.Compile(node.Value.String, '.')
					if err == nil {
						globs = append(globs, g)
					} else {
						log.Ctx(ctx).Warn().Err(err).Str("path", gr.Path).Msg("invalid glob")
					}
				}
				for _, childPtr := range node.Children {
					if *childPtr != nil {
						params = append(params, *childPtr)
					}
				}
			}
			groups[name] = group{name, globs}
		}
	}
	sort.Strings(sortedGroups)
	priorityGroups := make([]group, 0, len(priority)+len(sortedGroups))
	star := -1
	for i, name := range priority {
		if name == "*" {
			star = i
		} else if group, ok := groups[name]; ok {
			priorityGroups = append(priorityGroups, group)
			delete(groups, name)
		}
	}
	if star == -1 {
		star = len(priorityGroups)
	}
	starGroups := make([]group, 0, len(sortedGroups))
	for _, name := range sortedGroups {
		if group, ok := groups[name]; ok {
			starGroups = append(starGroups, group)
			delete(groups, name)
		}
	}
	return append(append(priorityGroups[:star], starGroups...), priorityGroups[star:]...), nil
}

type commonCaseResolver struct{}

func (cr commonCaseResolver) resolveCase(context.Context, string) *Case {
	panic("Dummy")
}

func (cr commonCaseResolver) getTemplateVar(string) string {
	panic("Dummy")
}
