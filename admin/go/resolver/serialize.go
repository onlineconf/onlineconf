package resolver

//go:generate go run github.com/ugorji/go/codec/codecgen -d 2018 -o serialize.generated.go serialize.go

import (
	"context"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/ugorji/go/codec"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var cborHandle = codec.CborHandle{BasicHandle: codec.BasicHandle{EncodeOptions: codec.EncodeOptions{Raw: true}}}

var stringValueReplacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")

type serializerData struct {
	Modules []string      `json:"modules"`
	Nodes   []interface{} `json:"nodes"`
}

type serializerParam struct {
	ID          int
	ContentType string
	Value       NullString
	Version     int
	Path        string
	Name        string
}

type serializer struct {
	data   serializerData
	pushed int
	compat bool
}

func newSerializer(ctx context.Context, sg *serverGraph, compat bool) *serializer {
	modules := sg.modules(ctx)

	ser := serializer{
		data: serializerData{
			Modules: make([]string, 0, len(modules)),
			Nodes:   make([]interface{}, 0, 1024),
		},
		compat: compat,
	}

	for name := range modules {
		ser.data.Modules = append(ser.data.Modules, name)
	}

	if tree := modules["TREE"]; tree != nil && tree.Path == "/" && compat {
		ser.writeParam("/", tree, true)
		return &ser
	}

	for _, path := range []string{"/", "/onlineconf", "/onlineconf/module"} {
		if node := sg.get(ctx, path); node != nil {
			ser.data.Nodes = append(ser.data.Nodes, newSerializerParam(path, node))
		}
	}
	for name, module := range modules {
		ser.writeParam("/onlineconf/module/"+name, module, getModuleCompat(ctx, compat, module))
	}

	return &ser
}

type moduleConfig struct {
	ChildLists bool `json:"child_lists" yaml:"child_lists"`
}

// getModuleCompat returns true if the request is received from perl updater, or if
// "child_lists" json/yaml module parameter (see [updater.moduleConfig]) isn't set (i.e. is missing or false)
func getModuleCompat(ctx context.Context, compat bool, module *Param) bool {
	if compat {
		return true
	}

	var modCfg moduleConfig

	if err := module.GetStruct(&modCfg); err != nil {
		log.Ctx(ctx).Err(err).Str("module", module.Path).Msg("error decoding module config")
		return true
	}

	return !modCfg.ChildLists
}

func (ser *serializer) writeParam(path string, param *Param, compat bool) {
	if param == nil {
		return
	}

	if compat && path != "/" && strings.HasSuffix(path, "/") {
		return // skip child lists for legacy perl updater
	}

	if !param.common {
		if param.seen {
			return
		}
		param.seen = true
		defer func() { param.seen = false }()
	}

	if param.common && param.Path == path {
		ser.data.Nodes = append(ser.data.Nodes, codec.Raw(param.serialized))

		for _, childPtr := range param.Children {
			if *childPtr != nil {
				ser.writeParam((*childPtr).Path, *childPtr, compat)
			}
		}
	} else {
		ser.data.Nodes = append(ser.data.Nodes, newSerializerParam(path, param))

		for name, childPtr := range param.Children {
			base := ""
			if path != "/" {
				base = path
			}
			path := base + "/" + name

			if *childPtr != nil {
				ser.writeParam(path, *childPtr, compat)
			}
		}
	}

	ser.pushed++
	if ser.pushed%1000 == 0 {
		runtime.Gosched()
	}
}

func (ser *serializer) serialize() ([]byte, error) {
	runtime.Gosched()
	defer runtime.Gosched()

	buf := make([]byte, 0, 100*len(ser.data.Nodes))
	enc := codec.NewEncoderBytes(&buf, &cborHandle)
	err := enc.Encode(ser.data)
	return buf, err
}

func serializeParam(param *Param) error {
	node := newSerializerParam(param.Path, param)

	buf := make([]byte, 0, 50)
	enc := codec.NewEncoderBytes(&buf, &cborHandle)
	err := enc.Encode(node)

	if err == nil {
		param.serialized = buf
	}

	return err
}

func newSerializerParam(path string, param *Param) *serializerParam {
	value := param.Value
	switch param.ContentType {
	case "application/json", "application/x-yaml", "application/x-null":
		break
	default:
		value.String = stringValueReplacer.Replace(value.String)
	}

	return &serializerParam{
		ID:          param.ID,
		ContentType: param.ContentType,
		Value:       value,
		Version:     param.Version,
		Path:        path,
		Name:        param.Name,
	}
}
