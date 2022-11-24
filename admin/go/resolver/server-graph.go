package resolver

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/onlineconf/onlineconf/admin/go/common"

	_ "unsafe"

	"github.com/bmatcuk/doublestar/v4"
)

//go:linkname matchWithSeparator github.com/bmatcuk/doublestar/v4.matchWithSeparator
func matchWithSeparator(pattern, name string, separator rune, validate bool) (matched bool, err error)

var serverSortRe = regexp.MustCompile(`(?:[\*\?]+|\{.*?\}|\[.*?\])`)

type serverGraph struct {
	graph
	mtime string
}

func newServerGraph(ctx context.Context, t *tree, server Server) *serverGraph {
	t.rw.RLock()
	defer t.rw.RUnlock()
	return &serverGraph{
		graph: graph{
			root:         t.root.deepClone(),
			caseResolver: *newServerCaseResolver(ctx, t, server),
		},
		mtime: t.mtime,
	}
}

func (graph *serverGraph) suspended(ctx context.Context) bool {
	node := graph.get(ctx, "/onlineconf/suspended")
	return node != nil && node.Value.Valid && node.Value.String != "" && node.Value.String != "0"
}

func (graph *serverGraph) modules(ctx context.Context) map[string]*Param {
	node := graph.get(ctx, "/onlineconf/module")
	if node == nil {
		return nil
	}
	graph.resolveChildren(ctx, &node)
	modules := make(map[string]*Param, len(node.Children))
	for name, childPtr := range node.Children {
		if name != "" {
			modules[name] = *childPtr
		}
	}
	return modules
}

type serverCaseResolver struct {
	server        Server
	datacenter    string
	groups        []string
	services      []string
	serviceGroups []string
	shortname     string
	ip            string
}

func newServerCaseResolver(ctx context.Context, t *tree, server Server) *serverCaseResolver {
	var datacenter string
	if server.Datacenter != "" {
		found := false
		for _, dc := range t.datacenters {
			if dc.name == server.Datacenter {
				found = true
				if dc.trust {
					datacenter = dc.name
				} else {
					log.Ctx(ctx).Warn().Str("datacenter", dc.name).Msg("datacenter value received from updater is not trusted")
				}
				break
			}
		}
		if !found {
			log.Ctx(ctx).Warn().Str("datacenter", server.Datacenter).Msg("unknown datacenter")
		}
	}
	if datacenter == "" {
	dcloop:
		for _, dc := range t.datacenters {
			for _, ipnet := range dc.ipnets {
				if ipnet.Contains(server.IP) {
					datacenter = dc.name
					break dcloop
				}
			}
		}
	}

	groups := make([]string, 0)
	for _, group := range t.groups {
		for _, glob := range group.globs {
			match, err := matchWithSeparator(glob, server.Host, '.', true)
			if err != nil {
				log.Ctx(ctx).Warn().Err(err).Str("group", glob).Msg("invalid glob")
			} else if match {
				groups = append(groups, group.name)
				break
			}
		}
	}

	service := common.Username(ctx)
	services := []string{service}
	for i := len(service) - 1; i >= 0; i-- {
		if service[i] == '/' {
			services = append(services, service[:i])
		}
	}

	serviceGroups := make([]string, 0)
	serviceGroupExists := make(map[string]bool)
	for _, service := range services {
		str := "service:" + service
		for _, group := range t.groups {
			if serviceGroupExists[group.name] {
				continue
			}
			for _, glob := range group.globs {
				match, err := matchWithSeparator(glob, str, '.', true)
				if err != nil {
					log.Ctx(ctx).Warn().Err(err).Str("group", glob).Msg("invalid glob")
				} else if match {
					serviceGroups = append(serviceGroups, group.name)
					serviceGroupExists[group.name] = true
					break
				}
			}
		}
	}

	if e := log.Ctx(ctx).Debug(); e.Enabled() {
		arr := zerolog.Arr()
		for _, group := range groups {
			arr.Str(group)
		}
		svcArr := zerolog.Arr()
		for _, group := range serviceGroups {
			svcArr.Str(group)
		}
		e.Str("datacenter", datacenter).Array("groups", arr).Str("service", service).Array("service_groups", svcArr).Msg("")
	}
	dot := strings.IndexRune(server.Host, '.')
	if dot == -1 {
		dot = len(server.Host)
	}
	return &serverCaseResolver{
		server:        server,
		datacenter:    datacenter,
		groups:        groups,
		services:      services,
		serviceGroups: serviceGroups,
		shortname:     server.Host[:dot],
		ip:            server.IP.String(),
	}
}

func (cr serverCaseResolver) resolveCase(ctx context.Context, param *Param) *Case {
	var data []Case
	err := json.Unmarshal([]byte(param.Value.String), &data)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to decode case json")
		return nil
	}

	byServer := make([]Case, 0)
	byGroup := make(map[string]Case, 0)
	byDatacenter := make(map[string]Case, 0)
	byService := make(map[string]Case, 0)
	var defaultCase Case
	hasDefaultCase := false
	for _, cs := range data {
		if cs.Server != "" {
			byServer = append(byServer, cs)
		} else if cs.Group != "" {
			byGroup[cs.Group] = cs
		} else if cs.Datacenter != "" {
			byDatacenter[cs.Datacenter] = cs
		} else if cs.Service != "" {
			byService[cs.Service] = cs
		} else if !hasDefaultCase {
			defaultCase = cs
			hasDefaultCase = true
		} else {
			log.Ctx(ctx).Error().Str("param", param.Path).Str("value", param.Value.String).Msg("invalid case")
		}
	}
	sort.Slice(byServer, func(i, j int) bool {
		var ic, jc int
		is := serverSortRe.ReplaceAllStringFunc(byServer[i].Server, func(string) string { ic++; return "" })
		js := serverSortRe.ReplaceAllStringFunc(byServer[j].Server, func(string) string { jc++; return "" })
		if len(is) != len(js) {
			return len(is) > len(js)
		}
		if ic != jc {
			return ic < jc
		}
		if len(byServer[i].Server) != len(byServer[j].Server) {
			return len(byServer[i].Server) > len(byServer[j].Server)
		}
		return byServer[i].Server < byServer[j].Server
	})

	for _, cs := range byServer {
		match, err := doublestar.Match(cs.Server, cr.server.Host) // using '/' separator for compatibility reason
		if err != nil {
			log.Ctx(ctx).Warn().Str("server", cs.Server).Msg("invalid server glob")
		} else if match {
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

	for _, service := range cr.services {
		if cs, ok := byService[service]; ok {
			return &cs
		}
	}

	for _, group := range cr.serviceGroups {
		if cs, ok := byGroup[group]; ok {
			return &cs
		}
	}

	if hasDefaultCase {
		return &defaultCase
	}

	return &Case{ContentType: "application/x-null"}
}

func (cr serverCaseResolver) getTemplateVar(name string) (string, bool) {
	switch name {
	case "hostname":
		return cr.server.Host, true
	case "hostname -s", "short_hostname":
		return cr.shortname, true
	case "hostname -i", "ip":
		return cr.ip, true
	default:
		return "", false
	}
}
