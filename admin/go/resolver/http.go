package resolver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"net"
	"net/http"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/onlineconf/onlineconf/admin/go/common"
)

var (
	ErrEmptyIP = errors.New("IP address is empty")
	ErrParseIP = errors.New("Failed to parse IP address")
)

type Server struct {
	Host string
	IP   net.IP
}

var configSemaphore = make(chan struct{}, int(0.8*float32(runtime.NumCPU())))

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if username, password, ok := req.BasicAuth(); ok {
			pwdHash, usernameExists := treeI.getServicePwdHash(username)
			if usernameExists && pwdHashEquals(req.Context(), pwdHash, password) {
				req = AddUsernameToRequest(req, username)
				next.ServeHTTP(w, req)
				return
			}
		}
		w.Header().Add("WWW-Authenticate", "Basic realm=onlineconf-updater")
		w.WriteHeader(401)
	})
}

func RegisterRoutes(r *mux.Router) {
	r.Use(authMiddleware)

	r.Path("/config").Methods("GET").HandlerFunc(serveConfig)
	r.Path("/activity").Methods("POST").HandlerFunc(serveActivity)
}

func serveConfig(w http.ResponseWriter, req *http.Request) {
	server, clientMTime := serverStatus(w, req)
	if server == nil {
		return
	}

	mtime := treeI.mtime
	if clientMTime >= mtime {
		w.Header().Add("X-OnlineConf-Admin-Last-Modified", mtime)
		http.Error(w, "", 304)
		return
	}

	configSemaphore <- struct{}{}
	defer func() { <-configSemaphore }()

	sg := newServerGraph(req.Context(), &treeI, *server)
	if sg.suspended(req.Context()) {
		log.Ctx(req.Context()).Info().Msg("suspended")
		w.Header().Set("X-OnlineConf-Suspended", "true")
		w.Header().Add("X-OnlineConf-Admin-Last-Modified", clientMTime)
		http.Error(w, "", 304)
		return
	}

	ser := newSerializer(req.Context(), sg)
	body, err := ser.serialize()
	if err != nil {
		WriteServerError(req.Context(), w, err)
		return
	}
	w.Header().Add("X-OnlineConf-Admin-Last-Modified", sg.mtime)
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
		WriteServerError(req.Context(), w, err)
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

	if clientVersion != "TEST" {
		err = updateServerActivity(req.Context(), server, clientMTime, clientVersion)
		if err != nil {
			WriteServerError(req.Context(), w, err)
			return nil, ""
		}
	}

	return server, clientMTime
}

func authenticateByIP(req *http.Request) (*Server, error) {
	ipstr := strings.Split(req.RemoteAddr, ":")[0]
	if ipstr == "" {
		return nil, ErrEmptyIP
	}
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return nil, ErrParseIP
	}

	host := req.Header.Get("X-OnlineConf-Client-Host")
	if host == "" {
		hosts, err := net.LookupAddr(ipstr)
		if err != nil {
			return nil, err
		}
		if len(hosts) == 0 {
			return nil, nil
		}
		if len(hosts) > 1 {
			arr := zerolog.Arr()
			for _, host := range hosts {
				arr.Str(host)
			}
			log.Ctx(req.Context()).Warn().Str("ip", ipstr).Array("hosts", arr).Msg("more then one PTR record found")
		}
		host = hosts[0]
	}

	ips, err := net.LookupHost(host)
	if err != nil {
		return nil, err
	}

	ok := false
	for _, ip := range ips {
		if ip == ipstr {
			ok = true
			break
		}
	}
	if !ok {
		return nil, nil
	}

	return &Server{Host: strings.TrimSuffix(host, "."), IP: ip}, nil
}

func pwdHashEquals(ctx context.Context, pwdHash []byte, password string) bool {
	h := sha256.New()
	if _, err := h.Write([]byte(password)); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to calculate sha256 hash")
		return false
	}
	return bytes.Equal(pwdHash, h.Sum(nil))
}
