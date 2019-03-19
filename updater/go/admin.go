package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

var ErrNotModified = errors.New("config not modified")

func getModules(config AdminConfig, hostname, mtime string) (string, map[string][]moduleParam, error) {
	respMtime, data, err := getConfigData(config, hostname, mtime)
	if err != nil {
		return "", nil, err
	}
	modules := prepareModules(data)
	return respMtime, modules, nil
}

func getConfigData(config AdminConfig, hostname, mtime string) (string, *ConfigData, error) {
	client := http.Client{}

	var uri string
	switch config.Port {
	case 80:
		uri = "http://" + config.Host
	case 443:
		uri = "https://" + config.Host
	default:
		uri = "http://" + config.Host + ":" + strconv.Itoa(config.Port)
	}
	uri += "/client/config"
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", nil, err
	}

	if hostname != "" {
		req.Header.Add("X-OnlineConf-Client-Host", hostname)
	}
	req.Header.Add("X-OnlineConf-Client-Version", version)
	req.SetBasicAuth(config.Username, config.Password)

	req.Header.Add("X-OnlineConf-Client-Mtime", mtime)

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		respMtime := resp.Header.Get("X-OnlineConf-Admin-Last-Modified")
		data, err := deserializeConfigData(resp.Body)
		return respMtime, data, err
	case 304:
		return "", nil, ErrNotModified
	default:
		return "", nil, errors.New(resp.Status)
	}
}

func prepareModules(data *ConfigData) (modules map[string][]moduleParam) {
	modules = make(map[string][]moduleParam, len(data.Modules))
	for _, m := range data.Modules {
		modules[m] = []moduleParam{}
	}
	sort.Slice(data.Nodes, func(i, j int) bool {
		return data.Nodes[i].Path < data.Nodes[j].Path
	})
	for _, param := range data.Nodes {
		if !strings.HasPrefix(param.Path, "/onlineconf/module/") {
			switch param.Path {
			case "/", "/onlineconf", "/onlineconf/module":
			default:
				log.Warn().Str("path", param.Path).Msg("parameter is out of '/onlineconf/module/' subtree")
			}
			continue
		}

		pc := strings.Split(param.Path, "/")
		moduleName := pc[3]
		if modules[moduleName] == nil {
			modules[moduleName] = []moduleParam{}
		}
		if len(pc) == 4 {
			continue
		}
		if param.ContentType == "application/x-null" {
			continue
		}

		var mParam moduleParam
		if moduleName == "TREE" {
			mParam.path = strings.Join(append([]string{""}, pc[4:]...), "/")
		} else {
			mParam.path = strings.Join(pc[4:], ".")
		}

		switch param.ContentType {
		case "application/json":
			var buf bytes.Buffer
			err := json.Compact(&buf, []byte(param.Value.String))
			if err != nil {
				log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to compact json")
				continue
			}
			mParam.value = buf.String()
			mParam.json = true
			if mParam.value == "null" {
				continue
			}
		case "application/x-yaml":
			v, err := YAMLToJSON([]byte(param.Value.String))
			if err != nil {
				log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to convert yaml to json")
				continue
			}
			mParam.value = string(v)
			mParam.json = true
			if mParam.value == "null" {
				continue
			}
		default:
			mParam.value = param.Value.String
		}

		modules[moduleName] = append(modules[moduleName], mParam)
	}
	return
}
