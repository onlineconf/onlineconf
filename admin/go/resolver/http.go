package resolver

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
	"net"
	"net/http"
	"strings"
)

var (
	ErrEmptyIP = errors.New("IP address is empty")
	ErrParseIP = errors.New("Failed to parse IP address")
)

type Server struct {
	Host string
	IP   net.IP
}

func RegisterRoutes(r *mux.Router) {
	r.Path("/client/config").Methods("GET").HandlerFunc(serveConfig)
	r.Path("/client/activity").Methods("POST").HandlerFunc(serveActivity)
}

func serveConfig(w http.ResponseWriter, req *http.Request) {
	server, err := authenticateByIP(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if server == nil {
		http.Error(w, "", 403)
	}
	log.Ctx(req.Context()).Info().Str("host", server.Host).Str("ip", server.IP.String()).Msg("")

	clientVersion := req.Header.Get("X-OnlineConf-Client-Version")
	if clientVersion == "" {
		http.Error(w, "Invalid X-OnlineConf-Client-Version", 400)
		return
	}

	server, clientMTime := serverStatus(w, req)
	if server == nil {
		return
	}

	mtime := treeI.mtime
	w.Header().Add("X-OnlineConf-Admin-Last-Modified", mtime)
	if clientMTime >= mtime {
		http.Error(w, "", 304)
		return
	}

	sg := newServerGraph(req.Context(), &treeI, *server)
	modules := sg.modules(req.Context())
	ser := newSerializer(modules)
	body, err := ser.serialize()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(body)
}

func serveActivity(w http.ResponseWriter, req *http.Request) {
	server, _ := serverStatus(w, req)
	if server != nil {
		w.WriteHeader(200)
	}
}

func serverStatus(w http.ResponseWriter, req *http.Request) (*Server, string) {
	server, err := authenticateByIP(req)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil, ""
	}
	if server == nil {
		http.Error(w, "", 403)
		return nil, ""
	}
	log.Ctx(req.Context()).Info().Str("host", server.Host).Str("ip", server.IP.String()).Msg("")

	clientVersion := req.Header.Get("X-OnlineConf-Client-Version")
	if clientVersion == "" {
		http.Error(w, "Invalid X-OnlineConf-Client-Version", 400)
		return nil, ""
	}

	clientMTime := req.Header.Get("X-OnlineConf-Client-MTime")

	err = updateServerActivity(req.Context(), server, clientMTime, clientVersion)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return nil, ""
	}

	return server, clientMTime
}

func authenticateByIP(req *http.Request) (*Server, error) {
	var ipstr string
	if AdminConfig.Test {
		if ipstr = req.Form.Get("ip"); ipstr != "" {
			if host := req.Form.Get("host"); host != "" {
				ips, err := net.LookupHost(host)
				if err != nil {
					return nil, err
				}
				if len(ips) == 0 {
					return nil, nil
				}
				ipstr = ips[0]
			}
		}
	}
	if ipstr == "" {
		var err error
		ipstr, _, err = net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			return nil, err
		}
	}
	if ipstr == "" {
		return nil, ErrEmptyIP
	}
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, ErrParseIP
	}

	hosts, err := net.LookupAddr(ipstr)
	if err != nil {
		return nil, err
	}
	if len(hosts) == 0 {
		return nil, nil
	}
	host := hosts[0]

	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, nil
	}
	if ips[0] != ipstr {
		return nil, nil
	}

	return &Server{Host: strings.TrimSuffix(host, "."), IP: ip}, nil
}
