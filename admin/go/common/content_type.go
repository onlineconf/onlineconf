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
