package resolver

//go:generate codecgen -d 2018 -o serialize.generated.go serialize.go

import (
	"context"
	"runtime"
	"strings"

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
	ID          int
	ContentType string
	Value       NullString
	Version     int
	Path        string
	Name        string
	MTime       string
}

type serializer struct {
	data   serializerData
	pushed int
}

func newSerializer(ctx context.Context, sg *serverGraph) *serializer {
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

		for name, childPtr := range param.Children {
			if *childPtr != nil {
				ser.writeParam((*childPtr).Path, name, *childPtr)
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
				ser.writeParam(path, name, *childPtr)
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
		MTime:       param.MTime,
	}
}
