package resolver

import (
	"context"
	"encoding/hex"
	"errors"
	"net"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidIPNets = errors.New("Invalid IP networks list")
)

type datacenter struct {
	name   string
	ipnets []net.IPNet
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

func (graph *commonGraph) readEphemeralIPs(ctx context.Context) ([]net.IPNet, error) {
	param := graph.get(ctx, "/onlineconf/ephemeral-ip")
	if param == nil {
		return nil, nil
	}
	return readIPNets(param)
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
		ipnets, err := readIPNets(*dcroot.Children[name])
		if err != nil {
			return nil, err
		}
		datacenters = append(datacenters, datacenter{name, ipnets})
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
	groupNames := make([]string, 0, len(grroot.Children))
	globsByName := make(map[string][]string, len(grroot.Children))
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
			groupNames = append(groupNames, name)
			globs := make([]string, 0, 1+len(gr.Children))
			params := []*Param{gr}
			for len(params) > 0 {
				node := params[0]
				params = params[1:]
				if node.Value.Valid {
					globs = append(globs, node.Value.String)
				}
				for _, childPtr := range node.Children {
					if *childPtr != nil {
						params = append(params, *childPtr)
					}
				}
			}
			globsByName[name] = globs
		}
	}
	sort.Strings(groupNames)
	byPriority := groupsSortedByPriority(groupNames, priority)
	sorted := groupsSortedByInclusion(byPriority, globsByName)
	result := make([]group, 0, len(sorted))
	for _, name := range sorted {
		globs := make([]glob.Glob, 0, len(globsByName[name]))
		for _, str := range globsByName[name] {
			g, err := glob.Compile(str, '.')
			if err == nil {
				globs = append(globs, g)
			} else {
				log.Ctx(ctx).Warn().Err(err).Str("group", name).Msg("invalid glob")
			}
		}
		if len(globs) > 0 {
			result = append(result, group{name, globs})
		}
	}
	return result, nil
}

func (graph *commonGraph) readServices(ctx context.Context) (map[string][]byte, error) {
	uroot := graph.get(ctx, "/onlineconf/service")
	if uroot == nil {
		return nil, nil
	}
	graph.resolveChildren(ctx, &uroot)

	res := map[string][]byte{}
	graph.fillServices(ctx, res, uroot, "")
	return res, nil
}

func (graph *commonGraph) fillServices(ctx context.Context, services map[string][]byte, node *Param, prefix string) {
	for name, childPtr := range node.Children {
		child := *childPtr
		if child == nil {
			continue
		}
		serviceName := prefix + name
		pwdHash, err := hex.DecodeString(child.Value.String)
		if err != nil {
			log.Ctx(ctx).Warn().Err(err).Str("service", serviceName).Msg("invalid password string")
			continue
		}
		services[serviceName] = pwdHash
		graph.fillServices(ctx, services, child, serviceName+"/")
	}
}

type commonCaseResolver struct{}

func (cr commonCaseResolver) resolveCase(context.Context, *Param) *Case {
	panic("Dummy")
}

func (cr commonCaseResolver) getTemplateVar(string) string {
	panic("Dummy")
}

func readIPNets(param *Param) ([]net.IPNet, error) {
	if !(param.Value.Valid && (param.ContentType == "text/plain" || param.ContentType == "application/x-list")) {
		return nil, ErrInvalidIPNets
	}
	ipnets := []net.IPNet{}
	for _, ipstr := range strings.Split(param.Value.String, ",") {
		_, ipnet, err := net.ParseCIDR(ipstr)
		if err != nil {
			return nil, err
		}
		ipnets = append(ipnets, *ipnet)
	}
	return ipnets, nil
}

func groupsSortedByPriority(groups []string, priority []string) []string {
	exist := make(map[string]bool, len(groups))
	for _, name := range groups {
		exist[name] = true
	}
	hasStar := false
	sorted := make([]string, 0, len(groups)+len(priority))
	for _, name := range priority {
		if name == "*" {
			hasStar = true
			sorted = append(sorted, subtract(groups, priority)...)
		} else if exist[name] {
			sorted = append(sorted, name)
		}
	}
	if !hasStar {
		sorted = append(sorted, groups...)
	}
	return uniqStrings(sorted)
}

func groupsSortedByInclusion(groups []string, globs map[string][]string) []string {
	existInHead := make(map[string]bool, len(groups))
	for _, g := range groups {
		existInHead[g] = true
	}
	tail := make([]string, 0, len(groups))
	for len(groups) > 0 {
		name := groups[len(groups)-1]
		groups = groups[:len(groups)-1]
		existInHead[name] = false
		for i := len(tail) - 1; i >= 0; i-- {
			n := tail[i]
			if name != n && !existInHead[n] && strictSuperset(globs[name], globs[n]) {
				groups = append(groups, n)
				existInHead[n] = true
			}
		}
		tail = append(tail, name)
	}
	nonuniq := make([]string, 0, len(tail))
	for i := len(tail) - 1; i >= 0; i-- {
		nonuniq = append(nonuniq, tail[i])
	}
	return uniqStrings(nonuniq)
}

func uniqStrings(orig []string) []string {
	uniq := make([]string, 0, len(orig))
	seen := make(map[string]bool, len(orig))
	for _, str := range orig {
		if !seen[str] {
			uniq = append(uniq, str)
			seen[str] = true
		}
	}
	return uniq
}

func subtract(orig, rem []string) []string {
	result := make([]string, 0, len(orig))
	exist := make(map[string]bool, len(rem))
	for _, s := range rem {
		exist[s] = true
	}
	for _, s := range orig {
		if !exist[s] {
			result = append(result, s)
		}
	}
	return result
}

func strictSuperset(outter, inner []string) bool {
	for _, i := range inner {
		ok := false
		for _, o := range outter {
			if i == o {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return len(outter) > len(inner)
}
