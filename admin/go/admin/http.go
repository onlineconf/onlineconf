package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var clientError = map[error]int{
	ErrAccessDenied:    403,
	ErrAlreadyExists:   400,
	ErrVersionNotMatch: 400,
	ErrCommentRequired: 400,
	ErrInvalidValue:    400,
	ErrNotEmpty:        400,
	ErrNotFound:        404,
	ErrParentNotFound:  400,
}

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
		writeClientError(req.Context(), w, 404, "Parameter not exists")
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
	var action string
	if verstr := f.Get("version"); verstr != "" {
		var version int
		version, err = strconv.Atoi(verstr)
		if err != nil {
			writeError(req.Context(), w, err)
			return
		}

		if newPath := f.Get("path"); newPath != "" {
			err = MoveParameter(req.Context(), path, newPath, f.Get("symlink") == "1", version, f.Get("comment"))
			action = "Moved"
		} else {
			err = SetParameter(req.Context(), path, version, f.Get("mime"), f.Get("data"), f.Get("comment"))
			action = "Changed"
		}
	} else if contentType := f.Get("mime"); contentType != "" {
		err = CreateParameter(req.Context(), path, contentType, f.Get("data"), f.Get("summary"), f.Get("description"), f.Get("notification"), f.Get("comment"))
		action = "Created"
	} else {
		_, nOk := f["notification"]
		_, sOk := f["summary"]
		_, dOk := f["description"]
		if len(f) == 1 && nOk {
			err = SetParameterNotification(req.Context(), path, f.Get("notification"))
			action = "NotificationChanged"
		} else if len(f) == 2 && sOk && dOk {
			err = SetParameterDescription(req.Context(), path, f.Get("summary"), f.Get("description"))
			action = "Renamed"
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

	data := struct {
		*Parameter
		Result string `json:"result"`
	}{
		Parameter: p,
		Result:    action,
	}
	writeResponse(req.Context(), w, data, err)
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
	writeResponse(req.Context(), w, map[string]string{"result": "Deleted"}, err)
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
		writeClientError(req.Context(), w, 403, "Allowed to root users only")
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
	data := map[string]string{"result": "Created"}
	writeResponse(req.Context(), w, data, err)
}

func serveDeleteGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := DeleteGroup(req.Context(), mux.Vars(req)["group"])
	data := map[string]string{"result": "Deleted"}
	writeResponse(req.Context(), w, data, err)
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
	data := map[string]string{"result": "Created"}
	writeResponse(req.Context(), w, data, err)
}

func serveDeleteUserFromGroup(w http.ResponseWriter, req *http.Request) {
	if !validateUserIsRoot(w, req) {
		return
	}
	err := DeleteUserFromGroup(req.Context(), mux.Vars(req)["group"], mux.Vars(req)["user"])
	data := map[string]string{"result": "Deleted"}
	writeResponse(req.Context(), w, data, err)
}

func validateUserCanEditAccess(w http.ResponseWriter, req *http.Request) bool {
	canEdit, err := UserCanEditAccess(req.Context(), Username(req.Context()), mux.Vars(req)["path"])
	if err != nil {
		writeError(req.Context(), w, err)
		return false
	}
	if !canEdit {
		writeClientError(req.Context(), w, 403, "Forbidden")
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
	var data interface{}
	if list, err := SelectAccess(req.Context(), mux.Vars(req)["path"]); err == nil {
		for _, access := range list {
			if access.Group == group {
				data = &struct {
					Access
					Result string `json:"result"`
				}{
					Access: access,
					Result: "Changed",
				}
				break
			}
		}
	}
	writeJSON(req.Context(), w, data)
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
	writeResponse(req.Context(), w, map[string]string{"Result": "Deleted"}, err)
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

func writeResponse(ctx context.Context, w http.ResponseWriter, data interface{}, err error) {
	if err == nil {
		writeJSON(ctx, w, data)
	} else {
		writeError(ctx, w, err)
	}
}

func writeJSON(ctx context.Context, w http.ResponseWriter, data interface{}) {
	content, err := json.Marshal(data)
	if err != nil {
		WriteServerError(ctx, w, err)
		return
	}

	header := w.Header()
	header.Add("Content-Type", "application/json")
	header.Add("Content-Length", strconv.Itoa(len(content)))
	w.Write(content)
}

func writeClientError(ctx context.Context, w http.ResponseWriter, code int, message string) {
	content, err := json.Marshal(map[string]string{
		"status":  "Error",
		"message": message,
	})
	if err != nil {
		WriteServerError(ctx, w, err)
		return
	}
	header := w.Header()
	header.Add("Content-Type", "application/json")
	header.Add("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(code)
	w.Write(content)
}

func writeError(ctx context.Context, w http.ResponseWriter, err error) {
	if code, ok := clientError[err]; ok {
		writeClientError(ctx, w, code, err.Error())
	} else {
		WriteServerError(ctx, w, err)
	}
}

func csrfProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if _, ok := req.Header["X-Requested-With"]; ok {
			next.ServeHTTP(w, req)
		} else {
			writeClientError(req.Context(), w, 403, "Forbidden")
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
