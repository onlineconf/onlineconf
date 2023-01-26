package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

var ErrNotModified = errors.New("config not modified")

var templateRe = regexp.MustCompile(`\$\{(.*?)\}`)

func getModules(config AdminConfig, hostname, datacenter, mtime string, vars map[string]string, force bool) (string, map[string][]moduleParam, error) {
	respMtime, data, err := getConfigData(config, hostname, datacenter, mtime, force)
	if err != nil {
		return "", nil, err
	}
	modules := prepareModules(data, vars)
	return respMtime, modules, nil
}

func getConfigData(config AdminConfig, hostname, datacenter, mtime string, force bool) (string, *ConfigData, error) {
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, config.URI+"/client/config", nil)
	if err != nil {
		return "", nil, err
	}

	if hostname != "" {
		req.Header.Add("X-OnlineConf-Client-Host", hostname)
	}
	if datacenter != "" {
		req.Header.Add("X-OnlineConf-Client-Datacenter", datacenter)
	}
	req.Header.Add("X-OnlineConf-Client-Version", version)
	req.SetBasicAuth(config.Username, config.Password)

	if !force {
		req.Header.Add("X-OnlineConf-Client-Mtime", mtime)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		respMtime := resp.Header.Get("X-OnlineConf-Admin-Last-Modified")
		data, err := deserializeConfigData(resp.Body)
		ResolveChecker.CleanStorage()
		return respMtime, data, err
	case 304:
		return "", nil, ErrNotModified
	default:
		return "", nil, errors.New(resp.Status)
	}
}

func prepareModules(data *ConfigData, vars map[string]string) (modules map[string][]moduleParam) {
	modules = make(map[string][]moduleParam, len(data.Modules))
	// ResolveChecker.storage.clean()
	for _, m := range data.Modules {
		modules[m] = []moduleParam{}
	}
	sort.Slice(data.Nodes, func(i, j int) bool {
		return data.Nodes[i].Path < data.Nodes[j].Path
	})
	defaultDelimiter := ""
	delimiter := defaultDelimiter
	for _, param := range data.Nodes {
		if !strings.HasPrefix(param.Path, "/onlineconf/module/") {
			switch param.Path {
			case "/", "/onlineconf":
			case "/onlineconf/module":
				defaultDelimiter = readModuleConfig(param).Delimiter
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
			delimiter = readModuleConfig(param).Delimiter
			if delimiter == "" {
				if defaultDelimiter != "" {
					delimiter = defaultDelimiter
				} else if moduleName == "TREE" {
					delimiter = "/"
				} else {
					delimiter = "."
				}
			}
			continue
		}
		if param.ContentType == "application/x-null" {
			continue
		}

		var mParam moduleParam
		if delimiter == "/" {
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
		case "application/x-template":
			mParam.value = templateRe.ReplaceAllStringFunc(param.Value.String, func(match string) string {
				name := match[2 : len(match)-1]

				val := vars[name]
				if val == "" {
					val, ok := TryResolveByResolverModule(context.TODO(), name) // todo: fix after major patch
					if !ok {
						log.Warn().Str("key", name).Msg("no resolver could get the value")
						return ""
					}

					return val
				}

				return val
			})
		default:
			mParam.value = param.Value.String
		}

		modules[moduleName] = append(modules[moduleName], mParam)
	}
	return
}

type moduleConfig struct {
	Delimiter string
}

func readModuleConfig(param ConfigParam) (cfg moduleConfig) {
	var err error
	value := []byte(param.Value.String)
	switch param.ContentType {
	case "application/x-yaml":
		value, err = YAMLToJSON(value)
		if err != nil {
			log.Warn().Err(err).Str("param", param.Path).Str("value", param.Value.String).Msg("failed to convert yaml to json")
			break
		}
		fallthrough
	case "application/json":
		err = json.Unmarshal(value, &cfg)
		if err != nil {
			log.Warn().Err(err).Str("param", param.Path).Str("value", string(value)).Msg("failed to parse json")
		}
	}
	return
}
