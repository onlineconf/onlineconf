package updater

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/rs/zerolog/log"
)

var ErrNotModified = errors.New("config not modified")

var templateRe = regexp.MustCompile(`\$\{(.*?)\}`)

type adminClient struct {
	client *http.Client
	config AdminConfig
}

func newAdminClient(config AdminConfig) *adminClient {
	client := &http.Client{}
	return &adminClient{
		config: config,
		client: client,
	}
}

func (a *adminClient) getModules(hostname, datacenter, mtime string, vars map[string]string) (string, map[string][]moduleParam, map[string]moduleConfig, error) {
	respMtime, data, err := a.getConfigData(hostname, datacenter, mtime)
	if err != nil {
		return "", nil, nil, err
	}
	modules, moduleConfigs := prepareModules(data, vars)
	return respMtime, modules, moduleConfigs, nil
}

func (a *adminClient) getConfigData(hostname, datacenter, mtime string) (string, *ConfigData, error) {
	req, err := http.NewRequest("GET", a.config.URI+"/client/config", nil)
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
	req.SetBasicAuth(a.config.Username, a.config.Password)

	req.Header.Add("X-OnlineConf-Client-Mtime", mtime)

	resp, err := a.client.Do(req)
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

func prepareModules(data *ConfigData, vars map[string]string) (modules map[string][]moduleParam, moduleConfigs map[string]moduleConfig) {
	moduleConfigs = make(map[string]moduleConfig, len(data.Modules))
	modules = make(map[string][]moduleParam, len(data.Modules))
	for _, m := range data.Modules {
		modules[m] = []moduleParam{}
	}
	sort.Slice(data.Nodes, func(i, j int) bool {
		return data.Nodes[i].Path < data.Nodes[j].Path
	})
	defaultModuleConfig := moduleConfig{}
	delimiter := ""
	for _, param := range data.Nodes {
		if !strings.HasPrefix(param.Path, "/onlineconf/module/") {
			switch param.Path {
			case "/", "/onlineconf":
			case "/onlineconf/module":
				defaultModuleConfig = readModuleConfig(param)
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
		if _, ok := moduleConfigs[moduleName]; !ok {
			moduleConfigs[moduleName] = defaultModuleConfig
		}

		if len(pc) == 4 {
			moduleConfig := readModuleConfig(param)
			if moduleConfig.Owner == "" {
				moduleConfig.Owner = defaultModuleConfig.Owner
			}
			if moduleConfig.Mode == "" {
				moduleConfig.Mode = defaultModuleConfig.Mode
			}
			if moduleConfig.Delimiter == "" {
				moduleConfig.Delimiter = defaultModuleConfig.Delimiter
			}
			moduleConfigs[moduleName] = moduleConfig
			delimiter = moduleConfig.Delimiter
			if delimiter == "" {
				if moduleName == "TREE" {
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
				return vars[name]
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
	Owner     string
	Mode      string

	ownerUid    int
	ownerGid    int
	ownerCached bool
}

func (config moduleConfig) getFileInfo(file string) (int, int, int, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return 0, 0, 0, err
	}
	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, 0, fmt.Errorf("can't fetch stat info for file %s", file)
	}
	return int(stat.Uid), int(stat.Gid), int(fileInfo.Mode()), nil
}

func (config moduleConfig) getNewFileAttrs(file string) (int, int, int, error) {
	newUID, newGID, err := config.getOwnerCached()
	if err != nil {
		return -1, -1, -1, err
	}

	fileUID, fileGID, fileMode, err := config.getFileInfo(file)
	if err != nil {
		return -1, -1, -1, err
	}

	if fileUID == newUID && fileGID == newGID {
		newUID = -1
		newGID = -1
	}

	newMode := -1
	if mode := config.Mode; mode != "" {
		modeUint, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return -1, -1, -1, fmt.Errorf("convert file mode %o failure...%w", modeUint, err)
		}
		if os.FileMode(modeUint) != os.FileMode(fileMode) {
			newMode = int(modeUint)
		}
	}

	return newUID, newGID, newMode, nil
}

func (config *moduleConfig) getOwnerCached() (uid int, gid int, err error) {
	if config.ownerCached {
		return config.ownerUid, config.ownerGid, nil
	}

	newUID, newGID, err := config.parseOwnerString()
	if err != nil {
		return -1, -1, err
	}

	config.ownerCached = true
	config.ownerUid = newUID
	config.ownerGid = newGID

	return newUID, newGID, nil
}

func (config moduleConfig) parseOwnerString() (uid int, gid int, err error) {
	ownerStr := strings.TrimSpace(config.Owner)
	if ownerStr == "" {
		return -1, -1, nil
	}

	// available formats: "user:group", "user", "user:", ":group", ":"
	uidStr, grpStr, found := strings.Cut(ownerStr, ":")

	uid = -1
	gid = -1
	if uidStr != "" {
		if uid64, err := strconv.ParseInt(uidStr, 10, 64); err == nil {
			uid = int(uid64)
		} else {
			usr, err := user.Lookup(uidStr)
			if err != nil {
				return 0, 0, err
			}
			uid, err = strconv.Atoi(usr.Uid)
			if err != nil {
				return 0, 0, err
			}

			// format: "user:"
			if found && grpStr == "" {

				gid, err = strconv.Atoi(usr.Gid)
				if err != nil {
					return 0, 0, err
				}
			}
		}
	}

	if grpStr != "" {
		if gid64, err := strconv.ParseInt(grpStr, 10, 64); err == nil {
			gid = int(gid64)
		} else {
			grp, err := user.LookupGroup(grpStr)
			if err != nil {
				return 0, 0, err
			}
			gid, err = strconv.Atoi(grp.Gid)
			if err != nil {
				return 0, 0, err
			}
		}
	}

	return uid, gid, nil
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
