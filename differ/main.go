package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
	"github.com/ugorji/go/codec"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
)

var cborHandle codec.CborHandle

var (
	username string
	password string
	oldUrl   string
	newUrl   string
	workers  int
)

func init() {
	flag.StringVar(&username, "username", "onlineconf", "username")
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&oldUrl, "old-url", "", "URL of the old onlineconf-admin")
	flag.StringVar(&newUrl, "new-url", "", "URL of the new onlineconf-admin")
	flag.IntVar(&workers, "workers", 10, "Total number of workers")
}

type result struct {
	ok   bool
	diff string
}

func main() {
	flag.Parse()
	jobs := make(chan string, workers)
	results := make(chan result, workers)
	for i := 0; i < workers; i++ {
		go func() {
			for s := range jobs {
				ok, diff := compareConfig(s)
				results <- result{ok, diff}
			}
		}()
	}
	servers, err := selectServers()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to fetch servers")
		return
	}
	go func() {
		for _, s := range servers {
			jobs <- s
		}
		close(jobs)
	}()
	i := 0
	failures := 0
	for r := range results {
		i++
		if !r.ok {
			failures++
		}
		fmt.Print(r.diff)
		if i == len(servers) {
			break
		}
	}
	fmt.Printf("Total failed: %d\n", failures)
}

type serverStatus struct {
	Host        string `json:"host"`
	MTimeAlert  bool   `json:"mtime_alert"`
	OnlineAlert bool   `json:"online_alert"`
}

func selectServers() ([]string, error) {
	url := oldUrl + "/monitoring"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(url + ": " + resp.Status)
	}
	dec := json.NewDecoder(resp.Body)
	var servers []serverStatus
	err = dec.Decode(&servers)
	if err != nil {
		return nil, err
	}
	active := make([]string, 0, 0)
	for _, s := range servers {
		if !s.MTimeAlert && !s.OnlineAlert {
			active = append(active, s.Host)
		}
	}
	return active, nil
}

func compareConfig(server string) (bool, string) {
	old, err := getConfig(oldUrl, server)
	if err != nil {
		log.Error().Err(err).Str("server", server).Msg("failed to fetch old config")
		return false, ""
	}
	new, err := getConfig(newUrl, server)
	if err != nil {
		log.Error().Err(err).Str("server", server).Msg("failed to fetch new config")
		return false, ""
	}
	fixOld(old, new)
	diff, err := compare(old, new)
	if err != nil {
		log.Error().Err(err).Str("server", server).Msg("failed to compare")
		return false, ""
	}
	return diff == "", "diff " + server + "\n" + diff
}

type config struct {
	Modules []string `json:"modules"`
	Nodes   []struct {
		Path        string
		ContentType string
		Value       []uint8
	} `json:"nodes"`
}

func getConfig(baseUrl, server string) (*config, error) {
	ips, err := net.LookupHost(server)
	if err != nil {
		return nil, err
	}
	ip := ips[0]
	_ = ip
	url := baseUrl + "/client/config"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	req.Header.Add("X-Real-IP", ip)
	req.Header.Add("X-OnlineConf-Client-Version", "TEST")
	req.Header.Add("X-OnlineConf-Client-MTime", "0000-00-00 00:00:00")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(url + ": " + resp.Status)
	}
	dec := codec.NewDecoder(resp.Body, &cborHandle)
	var data config
	err = dec.Decode(&data)
	if err != nil {
		return nil, err
	}
	sort.Strings(data.Modules)
	sort.Slice(data.Nodes, func(i, j int) bool {
		return data.Nodes[i].Path < data.Nodes[j].Path
	})
	return &data, nil
}

func compare(old, new *config) (string, error) {
	cfg := spew.Config
	cfg.DisablePointerAddresses = true
	cfg.DisableCapacities = true
	cfg.SortKeys = true
	a := cfg.Sdump(old)
	b := cfg.Sdump(new)

	ar, aw, err := os.Pipe()
	if err != nil {
		return "", err
	}

	br, bw, err := os.Pipe()
	if err != nil {
		return "", err
	}

	go func() {
		_, err := aw.WriteString(a)
		if err != nil {
			log.Error().Err(err).Msg("failed to write a")
		}
		aw.Close()
	}()

	go func() {
		_, err := bw.WriteString(b)
		if err != nil {
			log.Error().Err(err).Msg("failed to write b")
		}
		bw.Close()
	}()

	cmd := exec.Command("diff", "-u", "/dev/fd/3", "/dev/fd/4")
	cmd.ExtraFiles = []*os.File{ar, br}
	out, err := cmd.Output()
	if err != nil && err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus() != 1 {
		return "", err
	}
	return string(out), nil
}

func fixOld(old, new *config) {
	hasTree := false
	for _, n := range old.Nodes {
		if n.Path == "/onlineconf/module/TREE" {
			hasTree = true
			break
		}
	}
	if hasTree {
		nodes := old.Nodes
		old.Nodes = nil
		for _, n := range nodes {
			if strings.HasPrefix(n.Path, "/onlineconf/module/") {
				old.Nodes = append(old.Nodes, n)
			}
		}
	}
	for i, n := range old.Nodes {
		switch n.ContentType {
		case "application/x-template":
			old.Nodes[i].ContentType = "text/plain"
		case "application/x-case":
			if len(new.Nodes) > i {
				old.Nodes[i].ContentType = new.Nodes[i].ContentType
			}
		}
	}
}
