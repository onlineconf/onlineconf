package resolver

//go:generate go run github.com/ugorji/go/codec/codecgen -d 2018 -o serialize.generated.go serialize.go

import (
	"context"
	"database/sql"
	"path"
	"runtime"
	"strings"

	"github.com/ugorji/go/codec"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

// FolderIDBase is added to folder's ID to get a child list ID (/-suffixed), so they are predictable and repeatable.
const FolderIDBase = 1_000_000

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

func (ser *serializer) writeParam(path string, name string, param *Param) (maxMTime string) {
	if param == nil {
		return "" // valid "zero" value for mtime when using string comparison
	}

	if !param.common {
		if param.seen {
			return param.MTime
		}
		param.seen = true
		defer func() { param.seen = false }()
	}

	if param.common && param.Path == path {
		ser.data.Nodes = append(ser.data.Nodes, codec.Raw(param.serialized))

		maxMTime = ser.writeChildren(param, func(_ string, childPtr *Param) string {
			return childPtr.Path
		})
	} else {
		ser.data.Nodes = append(ser.data.Nodes, newSerializerParam(path, param))

		maxMTime = ser.writeChildren(param, func(name string, _ *Param) string {
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

	return maxMTime
}

func (ser *serializer) writeChildren(param *Param, pathFunc func(name string, childPtr *Param) string) (maxMTime string) {
	if len(param.Children) == 0 {
		return param.MTime
	}

	maxMTime = param.MTime
	childrenNames := make([]string, 0, len(param.Children))

	for name, childPtr := range param.Children {
		childMTime := ser.writeParam(pathFunc(name, *childPtr), name, *childPtr)
		if maxMTime < childMTime {
			maxMTime = childMTime
		}

		childrenNames = append(childrenNames, name)
	}

	listPath := pathFunc("", param)
	listName := path.Base(listPath) + "/"

	if !strings.HasSuffix(listPath, "/") {
		listPath += "/"
	}

	ser.writeParam(listPath, listName, &Param{
		ID:          param.ID + FolderIDBase,
		Name:        listName,
		Path:        listPath,
		MTime:       maxMTime,
		Version:     1,
		ContentType: "text/plain",
		Value: NullString{NullString: sql.NullString{
			String: strings.Join(childrenNames, ","),
			Valid:  true,
		}},
	})

	return maxMTime
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
