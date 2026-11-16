package common

// The content types a parameter value may have, as stored in
// my_config_tree.ContentType and served as "mime" by the API.
const (
	ContentTypeNull     = "application/x-null"
	ContentTypeText     = "text/plain"
	ContentTypeTemplate = "application/x-template"
	ContentTypeJSON     = "application/json"
	ContentTypeYAML     = "application/x-yaml"
	ContentTypeSymlink  = "application/x-symlink"
	ContentTypeCase     = "application/x-case"
	ContentTypeList     = "application/x-list"
	ContentTypeServer   = "application/x-server"
	ContentTypeServer2  = "application/x-server2"
)

// CaseBranch is one entry of a case value: the branch's condition — at most
// one of server, group, datacenter and service, none for the default branch
// — and the content type and value it selects. The conditions are
// NullStrings so that the validator tells a condition present but empty,
// which it refuses, from an absent one; the resolver selects by value and
// treats both as absent. Unknown keys are ignored. The admin validates and
// renders branches and the resolver selects one; nobody writes a branch
// back, the UI edits the JSON text.
type CaseBranch struct {
	Server      NullString `json:"server"`
	Group       NullString `json:"group"`
	Datacenter  NullString `json:"datacenter"`
	Service     NullString `json:"service"`
	ContentType string     `json:"mime"`
	Value       NullString `json:"value"`
}
