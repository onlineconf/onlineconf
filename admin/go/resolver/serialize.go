package resolver

//go:generate go run github.com/ugorji/go/codec/codecgen -d 2018 -o serialize.generated.go serialize.go

import (
	"context"
	"database/sql"
	"encoding/json"
	"path"
	"runtime"
	"sort"
	"strings"

	zLog "github.com/rs/zerolog/log"
	"github.com/ugorji/go/codec"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var cborHandle codec.CborHandle = codec.CborHandle{BasicHandle: codec.BasicHandle{EncodeOptions: codec.EncodeOptions{Raw: true}}}

var stringValueReplacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")

type serializerData struct {
	Modules []string      `json:"modules"`
	Nodes   []interface{} `json:"nodes"`
}

type serializerParam struct {
	//Name        string
	Path        string
	ContentType string
	Value       NullString
}

type serializer struct {
	data   serializerData
	pushed int
}

func newSerializer(ctx context.Context, sg *serverGraph, compat bool) *serializer {
	modules := sg.modules(ctx)

	ser := serializer{
		data: serializerData{
			Modules: make([]string, 0, len(modules)),
			Nodes:   make([]interface{}, 0, 1024),
		},
	}

	for name := range modules {
		ser.data.Modules = append(ser.data.Modules, name)
	}

	if tree := modules["TREE"]; tree != nil && tree.Path == "/" && compat {
		ser.writeParam("/", "", tree)
		return &ser
	}

	for _, path := range []string{"/", "/onlineconf", "/onlineconf/module"} {
		if node := sg.get(ctx, path); node != nil {
			ser.data.Nodes = append(ser.data.Nodes, newSerializerParam(path, node))
		}
	}
	for name, module := range modules {
		ser.writeParam("/onlineconf/module/"+name, name, module)
	}

	return &ser
}

func (ser *serializer) writeParam(path string, name string, param *Param) {
	if param == nil {
		return
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

		ser.writeChildren(param, func(_ string, childPtr *Param) string {
			return childPtr.Path
		})
	} else {
		ser.data.Nodes = append(ser.data.Nodes, newSerializerParam(path, param))

		ser.writeChildren(param, func(name string, _ *Param) string {
			base := ""
			if path != "/" {
				base = path
			}

			return base + "/" + name
		})
	}

	ser.pushed++
	if ser.pushed%1000 == 0 {
		runtime.Gosched()
	}
}

func (ser *serializer) writeChildren(param *Param, pathFunc func(name string, childPtr *Param) string) {
	if len(param.Children) == 0 {
		return
	}

	childrenNames := make([]string, 0, len(param.Children))

	for name, childPtr := range param.Children {
		ser.writeParam(pathFunc(name, *childPtr), name, *childPtr)

		childrenNames = append(childrenNames, name)
	}

	listPath := pathFunc("", param)
	listName := path.Base(listPath) + "/"

	if !strings.HasSuffix(listPath, "/") {
		listPath += "/"
	}

	sort.Strings(childrenNames)

	chilrenNamesJson, err := json.Marshal(childrenNames)
	if err != nil {
		zLog.Err(err).Msg("error encoding children names list") // should not happen
	}

	ser.writeParam(listPath, listName, &Param{
		//Name:        listName,
		Path:        listPath,
		ContentType: "application/json",
		Value: NullString{NullString: sql.NullString{
			String: string(chilrenNamesJson),
			Valid:  true,
		}},
	})

	return
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
		//Name:        param.Name,
		Path:        path,
		ContentType: param.ContentType,
		Value:       value,
	}
}
