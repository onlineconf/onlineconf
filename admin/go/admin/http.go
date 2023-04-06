package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var writeError = WriteErrorFunc(map[error]ErrorResponse{
	ErrAccessDenied:    {HTTPCode: 403, ErrorCode: "AccessDenied"},
	ErrAlreadyExists:   {HTTPCode: 400, ErrorCode: "AlreadyExists"},
	ErrVersionNotMatch: {HTTPCode: 400, ErrorCode: "VersionNotMatch"},
	ErrCommentRequired: {HTTPCode: 400, ErrorCode: "CommentRequired"},
	ErrInvalidValue:    {HTTPCode: 400, ErrorCode: "InvalidValue"},
	ErrNotEmpty:        {HTTPCode: 400, ErrorCode: "NotEmpty"},
	ErrNotFound:        {HTTPCode: 404, ErrorCode: "NotFound"},
	ErrParentNotFound:  {HTTPCode: 400, ErrorCode: "ParentNotFound"},
})
var writeResponse = WriteResponseOrErrorFunc(writeError)

func RegisterRoutes(r *mux.Router) {
	r.Use(authMiddleware)
	r.Use(csrfProtection)

	c := r.Path("/config{path:(?:/|(?:/[^/]+)+)}").Subrouter()
	c.Methods("GET").HandlerFunc(serveGetConfig)
	c.Methods("POST").HandlerFunc(serveSetConfig)
	c.Methods("DELETE").HandlerFunc(serveDeleteConfig)
	r.Path("/batch/GET/config").Methods("POST").HandlerFunc(serveBatchGetConfig)

	r.Path("/whoami").Methods("GET").HandlerFunc(serveWhoami)
	r.Path("/user").Methods("GET").HandlerFunc(serveGetUsers)

	g := r.PathPrefix("/group/").Subrouter()
	g.Path("/").Methods("GET").HandlerFunc(serveGetGroups)
	g.Path("/{group}").Methods("GET").HandlerFunc(serveGetGroupUsers)
	g.Path("/{group}").Methods("POST").HandlerFunc(serveCreateGroup)
	g.Path("/{group}").Methods("DELETE").HandlerFunc(serveDeleteGroup)
	g.Path("/{group}/{user}").Methods("POST").HandlerFunc(serveAddUserToGroup)
	g.Path("/{group}/{user}").Methods("DELETE").HandlerFunc(serveDeleteUserFromGroup)

	a := r.Path("/access{path:(?:/|(?:/[^/]+)+)}").Subrouter()
	a.Methods("GET").HandlerFunc(serveGetAccess)
	a.Methods("POST").HandlerFunc(serveSetAccess)
	a.Methods("DELETE").HandlerFunc(serveDeleteAccess)

	r.Path("/search").Methods("GET").HandlerFunc(serveSearch)
	r.Path("/global-log").Methods("GET").HandlerFunc(serveGlobalLog)
	r.Path("/log{path:(?:/|(?:/[^/]+)+)}").Methods("GET").HandlerFunc(serveParameterLog)

	r.Path("/monitoring").Methods("GET").HandlerFunc(serveMonitoring)
	r.Path("/monitoring/{host}").Methods("DELETE").HandlerFunc(serveDeleteServerFromMonitoring)

	r.Path("/ui-config").Methods("GET").HandlerFunc(serveUIConfig)
}

func serveGetConfig(w http.ResponseWriter, req *http.Request) {
	path := mux.Vars(req)["path"]
	var parameter *Parameter
	var err error
	switch req.URL.Query().Get("symlink") {
	case "":
		parameter, err = SelectParameter(req.Context(), path)
	case "follow":
		parameter, err = SelectParameterFollowingSymlink(req.Context(), path)
	case "resolve":
		parameter, err = SelectParameterResolvingSymlink(req.Context(), path)
	}
	if err != nil {
		writeError(req.Context(), w, err)
		return
	} else if parameter == nil {
		writeError(req.Context(), w, ErrNotFound)
		return
	}
	var data interface{}
	switch req.URL.Query().Get("depth") {
	case "", "children":
		data, err = parameter.WithChildren(req.Context())
	case "subtree":
		data, err = parameter.WithSubtree(req.Context())
	default:
		data = parameter
	}
	writeResponse(req.Context(), w, data, err)
}

func serveSetConfig(w http.ResponseWriter, req *http.Request) {
	path := mux.Vars(req)["path"]
	req.ParseForm()
	f := req.PostForm

	var err error
	if verstr := f.Get("version"); verstr != "" {
		var version int
		version, err = strconv.Atoi(verstr)
		if err != nil {
			writeError(req.Context(), w, err)
			return
		}

		if newPath := f.Get("path"); newPath != "" {
			err = MoveParameter(req.Context(), path, newPath, f.Get("symlink") == "1", version, f.Get("comment"))
		} else {
			err = SetParameter(req.Context(), path, version, f.Get("mime"), f.Get("data"), f.Get("comment"))
		}
	} else if contentType := f.Get("mime"); contentType != "" {
		err = CreateParameter(req.Context(), path, contentType, f.Get("data"), optValue(f, "summary"), optValue(f, "description"), optValue(f, "notification"), f.Get("comment"))
	} else {
		_, nOk := f["notification"]
		_, sOk := f["summary"]
		_, dOk := f["description"]
		if len(f) == 1 && nOk {
			err = SetParameterNotification(req.Context(), path, f.Get("notification"))
		} else if len(f) == 2 && sOk && dOk {
			err = SetParameterDescription(req.Context(), path, f.Get("summary"), f.Get("description"))
		} else {
			http.Error(w, "invalid request", 400)
			return
		}
	}

	if err != nil {
		writeError(req.Context(), w, err)
		return
	}

	p, err := SelectParameter(req.Context(), path)
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}

	writeResponse(req.Context(), w, p, err)
}

func serveDeleteConfig(w http.ResponseWriter, req *http.Request) {
	origMethod := req.Method
	req.Method = "POST"
	req.ParseForm()
	req.Method = origMethod
	version, err := strconv.Atoi(req.PostFormValue("version"))
	if err == nil {
		err = DeleteParameter(req.Context(), mux.Vars(req)["path"], version, req.PostFormValue("comment"))
	}
	writeResponse(req.Context(), w, struct{}{}, err)
}

func serveBatchGetConfig(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}
	data, err := SelectWithChildrenMulti(req.Context(), req.PostForm["id[]"])
	writeResponse(req.Context(), w, data, err)
}

func serveWhoami(w http.ResponseWriter, req *http.Request) {
	username := Username(req.Context())
	canEdit, err := UserIsRoot(req.Context(), username)
	data := map[string]interface{}{
		"username":        username,
		"can_edit_groups": canEdit,
	}
	writeResponse(req.Context(), w, data, err)
}

func serveGetUsers(w http.ResponseWriter, req *http.Request) {
	list, err := authenticator.SelectUsers(req.Context(), req.URL.Query().Get("term"))
	writeResponse(req.Context(), w, list, err)
}

func validateUserIsRoot(w http.ResponseWriter, req *http.Request) bool {
	canEdit, err := UserIsRoot(req.Context(), Username(req.Context()))
	if err != nil {
		writeError(req.Context(), w, err)
		return false
	}
	if !canEdit {
		WriteResponse(req.Context(), w, ErrorResponse{
			HTTPCode:  403,
			ErrorCode: "RootRequired",
			Message:   "Allowed to root users only",
		})
		return false
	}
	return true
}

func serveGetGroups(w http.ResponseWriter, req *http.Request) {
	list, err := SelectGroups(req.Context())
	writeResponse(req.Context(), w, list, err)
}

func serveCreateGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := CreateGroup(req.Context(), mux.Vars(req)["group"])
	writeResponse(req.Context(), w, struct{}{}, err)
}

func serveDeleteGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := DeleteGroup(req.Context(), mux.Vars(req)["group"])
	writeResponse(req.Context(), w, struct{}{}, err)
}

func serveGetGroupUsers(w http.ResponseWriter, req *http.Request) {
	list, err := SelectGroupUsers(req.Context(), mux.Vars(req)["group"])
	writeResponse(req.Context(), w, list, err)
}

func serveAddUserToGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := AddUserToGroup(req.Context(), mux.Vars(req)["group"], mux.Vars(req)["user"])
	writeResponse(req.Context(), w, struct{}{}, err)
}

func serveDeleteUserFromGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := DeleteUserFromGroup(req.Context(), mux.Vars(req)["group"], mux.Vars(req)["user"])
	writeResponse(req.Context(), w, struct{}{}, err)
}

func validateUserCanEditAccess(w http.ResponseWriter, req *http.Request) bool {
	canEdit, err := UserCanEditAccess(req.Context(), Username(req.Context()), mux.Vars(req)["path"])
	if err != nil {
		writeError(req.Context(), w, err)
		return false
	}
	if !canEdit {
		WriteResponse(req.Context(), w, ErrorResponse{HTTPCode: 403, ErrorCode: "Forbidden", Message: "Forbidden"})
		return false
	}
	return true
}

func writeModifyAccessResponse(w http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		writeError(req.Context(), w, err)
		return
	}
	group := req.PostFormValue("group")
	if list, err := SelectAccess(req.Context(), mux.Vars(req)["path"]); err == nil {
		for _, access := range list {
			if access.Group == group {
				WriteResponse(req.Context(), w, access)
				return
			}
		}
	}
	WriteResponse(req.Context(), w, struct{}{})
}

func serveGetAccess(w http.ResponseWriter, req *http.Request) {
	data, err := SelectAccess(req.Context(), mux.Vars(req)["path"])
	writeResponse(req.Context(), w, data, err)
}

func serveSetAccess(w http.ResponseWriter, req *http.Request) {
	if !validateUserCanEditAccess(w, req) {
		return
	}
	var rw NullBool
	switch req.PostFormValue("rw") {
	case "true":
		rw.Valid = true
		rw.Bool = true
	case "false":
		rw.Valid = true
		rw.Bool = false
	}
	err := SetAccess(req.Context(), mux.Vars(req)["path"], req.PostFormValue("group"), rw)
	writeModifyAccessResponse(w, req, err)
}

func serveDeleteAccess(w http.ResponseWriter, req *http.Request) {
	if !validateUserCanEditAccess(w, req) {
		return
	}
	origMethod := req.Method
	req.Method = "POST"
	req.ParseForm()
	req.Method = origMethod
	err := DeleteAccess(req.Context(), mux.Vars(req)["path"], req.PostFormValue("group"))
	writeModifyAccessResponse(w, req, err)
}

func serveSearch(w http.ResponseWriter, req *http.Request) {
	list, err := SearchParameters(req.Context(), req.URL.Query().Get("term"))
	writeResponse(req.Context(), w, list, err)
}

func serveGlobalLog(w http.ResponseWriter, req *http.Request) {
	filter := LogFilter{
		Author: req.URL.Query().Get("author"),
		Branch: req.URL.Query().Get("branch"),
		From:   req.URL.Query().Get("from"),
		Till:   req.URL.Query().Get("till"),
		All:    req.URL.Query().Get("all") != "",
	}
	lastID, _ := strconv.Atoi(req.URL.Query().Get("lastid"))
	list, err := SelectLog(req.Context(), filter, lastID)
	writeResponse(req.Context(), w, list, err)
}

func serveParameterLog(w http.ResponseWriter, req *http.Request) {
	filter := LogFilter{
		Path: mux.Vars(req)["path"],
		All:  true,
	}
	lastID, _ := strconv.Atoi(req.URL.Query().Get("lastid"))
	list, err := SelectLog(req.Context(), filter, lastID)
	writeResponse(req.Context(), w, list, err)
}

func serveMonitoring(w http.ResponseWriter, req *http.Request) {
	list, err := SelectServerStatus(req.Context(), req.URL.Query().Get("sort"))
	writeResponse(req.Context(), w, list, err)
}

func serveDeleteServerFromMonitoring(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := DeleteServerStatus(req.Context(), mux.Vars(req)["host"])
	writeResponse(req.Context(), w, struct{}{}, err)
}

const avatarPath = "/onlineconf/ui/avatar"

var errInvalidAvatarContentType = errors.New("invalid content-type")

type UIConfig struct {
	Avatar *Avatar `json:"avatar,omitempty"`
}

type Avatar struct {
	URI      string            `json:"uri"`
	Domain   string            `json:"domain"`
	Gravatar bool              `json:"gravatar"`
	Rename   map[string]string `json:"rename,omitempty"`
	Link     *struct {
		URI    string            `json:"uri"`
		Rename map[string]string `json:"rename,omitempty"`
	} `json:"link,omitempty"`
}

// All data read by this function from the configuration tree is visible
// to all users without any restrictions through `/ui-config`.
// To prevent a security hole the following rules must be applied:
// - symlinks MUST NOT be supported
// - any data MUST be decoded using an explicit schema
func serveUIConfig(w http.ResponseWriter, req *http.Request) {
	var config UIConfig
	avatarParam, err := SelectParameter(req.Context(), avatarPath)
	if avatarParam != nil {
		var avatar Avatar
		switch avatarParam.ContentType {
		case "application/json":
			err = json.Unmarshal([]byte(avatarParam.Value.String), &avatar)
		case "application/x-yaml":
			err = yaml.Unmarshal([]byte(avatarParam.Value.String), &avatar)
		default:
			err = errInvalidAvatarContentType
		}
		if err == nil {
			config.Avatar = &avatar
		}
	}
	if err != nil {
		log.Warn().Err(err).Str("path", avatarPath).Msg("failed to read avatar config")
	}
	writeResponse(req.Context(), w, config, err)
}

func csrfProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := req.Header["X-Requested-With"]; ok {
			next.ServeHTTP(w, req)
		} else {
			WriteResponse(req.Context(), w, ErrorResponse{HTTPCode: 403, ErrorCode: "Forbidden", Message: "Forbidden"})
		}
	})
}

func RootUsersOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if validateUserIsRoot(w, req) {
			next.ServeHTTP(w, req)
		}
	})
}

func optValue(value url.Values, key string) NullString {
	var opt NullString
	if _, ok := value[key]; ok {
		opt.Valid = true
		opt.String = value.Get(key)
	}
	return opt
}
