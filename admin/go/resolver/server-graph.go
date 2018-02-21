package resolver

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.corp.mail.ru/mydev/onlineconf/admin/go/hostname"
	"path"
	"regexp"
	"sort"
	"strings"
)

type serverGraph struct {
	graph
}

func newServerGraph(ctx context.Context, t *tree, server Server) *serverGraph {
	return &serverGraph{
		graph: graph{
			root:         t.clone(),
			caseResolver: *newServerCaseResolver(ctx, t, server),
		},
	}
}

func (graph *serverGraph) modules(ctx context.Context) map[string]*Param {
	node := graph.get(ctx, "/onlineconf/module")
	graph.resolveChildren(ctx, &node)
	modules := make(map[string]*Param, len(node.Children))
	for name, childPtr := range node.Children {
		modules[name] = *childPtr
	}
	return modules
}

type serverCaseResolver struct {
	server     Server
	datacenter string
	groups     []string
	shortname  string
	ip         string
}

func newServerCaseResolver(ctx context.Context, t *tree, server Server) *serverCaseResolver {
	var datacenter string
	for _, dc := range t.datacenters {
		if dc.ipnet.Contains(server.IP) {
			datacenter = dc.name
			break
		}
	}
	groups := make([]string, 0)
	for _, group := range t.groups {
		for _, glob := range group.globs {
			if matched, _ := hostname.Match(glob, server.Host); matched {
				groups = append(groups, group.name)
				break
			}
		}
	}
	if e := log.Ctx(ctx).Debug(); e.Enabled() {
		arr := zerolog.Arr()
		for _, group := range groups {
			arr.Str(group)
		}
		e.Str("datacenter", datacenter).Array("groups", arr).Msg("")
	}
	dot := strings.IndexRune(server.Host, '.')
	if dot == -1 {
		dot = len(server.Host)
	}
	return &serverCaseResolver{
		server:     server,
		datacenter: datacenter,
		groups:     groups,
		shortname:  server.Host[:dot],
		ip:         server.IP.String(),
	}
}

func (cr serverCaseResolver) resolveCase(ctx context.Context, value string) *Case {
	var data []Case
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("value", value).Msg("failed to decode case json")
		return nil
	}

	byServer := make(CasesByServer, 0)
	byGroup := make(map[string]Case, 0)
	byDatacenter := make(map[string]Case, 0)
	var defaultCase Case
	hasDefaultCase := false
	for _, cs := range data {
		if cs.Server != "" {
			byServer = append(byServer, cs)
		} else if cs.Group != "" {
			byGroup[cs.Group] = cs
		} else if cs.Datacenter != "" {
			byDatacenter[cs.Datacenter] = cs
		} else if !hasDefaultCase {
			defaultCase = cs
			hasDefaultCase = true
		} else {
			log.Ctx(ctx).Error().Str("value", value).Msg("invalid case")
		}
	}
	sort.Sort(byServer)

	for _, cs := range byServer {
		if matched, _ := path.Match(cs.Server, cr.server.Host); matched { // TODO replace to hostname.Match()
			return &cs
		}
	}

	for _, group := range cr.groups {
		if cs, ok := byGroup[group]; ok {
			return &cs
		}
	}

	if cs, ok := byDatacenter[cr.datacenter]; ok {
		return &cs
	}

	if hasDefaultCase {
		return &defaultCase
	}

	return &Case{}
}

func (cr serverCaseResolver) getTemplateVar(name string) string {
	switch name {
	case "hostname":
		return cr.server.Host
	case "hostname -s", "short_hostname":
		return cr.shortname
	case "hostname -i", "ip":
		return cr.ip
	default:
		return ""
	}
}

var serverSortRe = regexp.MustCompile(`(?:[\*\?]+|\{.*?\}|\[.*?\])`)

type CasesByServer []Case

func (cases CasesByServer) Len() int {
	return len(cases)
}

func (cases CasesByServer) Less(i, j int) bool {
	is := serverSortRe.ReplaceAllString(cases[i].Server, "")
	js := serverSortRe.ReplaceAllString(cases[j].Server, "")
	if len(is) != len(js) {
		return len(is) > len(js)
	}
	if is != js {
		return is < js
	}
	if len(cases[i].Server) != len(cases[j].Server) {
		return len(cases[i].Server) > len(cases[j].Server)
	}
	return cases[i].Server < cases[j].Server
}

func (cases CasesByServer) Swap(i, j int) {
	cases[i], cases[j] = cases[j], cases[i]
}
