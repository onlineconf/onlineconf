package main

import (
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/onlineconf/onlineconf/admin/go/admin"
	"github.com/onlineconf/onlineconf/admin/go/botapi"
	. "github.com/onlineconf/onlineconf/admin/go/common"
	"github.com/onlineconf/onlineconf/admin/go/resolver"
)

var configFile = flag.String("config", "/usr/local/etc/onlineconf-admin.yaml", "config file")

type ConfigFile struct {
	HTTP struct {
		Listen string
		TLS    struct {
			Cert             string
			Key              string
			RedirectFromHTTP string `yaml:"redirect_from_http"`
		}
		BehindReverseProxy bool   `yaml:"behind_reverse_proxy"`
		StaticRoot         string `yaml:"static_root"`
	}
	CommonConfig      `yaml:",inline"`
	admin.AdminConfig `yaml:",inline"`
}

var config *ConfigFile

func main() {
	flag.Parse()
	config = readConfigFile(*configFile)
	CommonInitialize(config.CommonConfig)
	admin.Initialize(config.AdminConfig)
	resolver.Initialize()

	r := mux.NewRouter()

	registerPProf(r.PathPrefix("/debug/pprof/").Subrouter())

	resolverRouter := r.PathPrefix("/client/").Subrouter()
	resolver.RegisterRoutes(resolverRouter)

	adminRouter := r.PathPrefix("/").Subrouter()
	admin.RegisterRoutes(adminRouter)

	botapiRouter := r.PathPrefix("/botapi/").Subrouter()
	botapi.RegisterRoutes(botapiRouter)

	registerStaticFileServer(r.PathPrefix("/").Subrouter())

	handler := hlog.AccessHandler(writeAccessLog)(r)
	handler = UsernameMiddleware(handler)
	if config.HTTP.BehindReverseProxy {
		handler = handlers.ProxyHeaders(handler)
	}
	handler = hlog.RequestIDHandler("reqid", "")(handler)
	handler = hlog.NewHandler(log.Logger)(handler)

	server := http.Server{
		Addr:    config.HTTP.Listen,
		Handler: handler,
	}
	if config.HTTP.TLS.Cert != "" {
		startHTTPRedirectionServer()
		if err := server.ListenAndServeTLS(config.HTTP.TLS.Cert, config.HTTP.TLS.Key); err != nil {
			log.Fatal().Err(err).Msg("failed to start HTTPS server")
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("failed to start HTTP server")
		}
	}
}

func readConfigFile(filename string) *ConfigFile {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to read config file")
	}
	var config ConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal().Err(err).Str("file", filename).Msg("failed to parse config file")
	}

	if config.Auth.Method == "header" && !config.HTTP.BehindReverseProxy {
		log.Fatal().Msg("header authenticator can be used behind reverse proxy only")
	}

	if config.Database.MaxConn <= 0 {
		config.Database.MaxConn = 100
	}

	return &config
}

func writeAccessLog(r *http.Request, status, size int, duration time.Duration) {
	l := hlog.FromRequest(r).Info().
		Str("remote", strings.Split(r.RemoteAddr, ":")[0])
	if username := Username(r.Context()); username != "" {
		l = l.Str("username", username)
	}
	l.Str("method", r.Method).
		Str("uri", r.RequestURI).
		Int("status", status).
		Int("size", size).
		Dur("duration", duration).
		Msg("")
}

func registerPProf(r *mux.Router) {
	r.Use(admin.RootUsersOnly)

	r.Path("/cmdline").HandlerFunc(pprof.Cmdline)
	r.Path("/profile").HandlerFunc(pprof.Profile)
	r.Path("/symbol").HandlerFunc(pprof.Symbol)
	r.Path("/trace").HandlerFunc(pprof.Trace)
	r.PathPrefix("/").HandlerFunc(pprof.Index)
}

func registerStaticFileServer(r *mux.Router) {
	if config.HTTP.StaticRoot != "" {
		handler := &staticFileHandler{http.FileServer(staticFileSystem{http.Dir(config.HTTP.StaticRoot)})}
		r.PathPrefix("/").Methods("GET").Handler(handler)
	}
}

var aliases = map[string]string{
	"/history/":      "/",
	"/server/":       "/",
	"/access-group/": "/",
}

type staticFileHandler struct {
	http.Handler
}

func (f *staticFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/classic/" {
		w.Header().Set("Content-Type", "application/xhtml+xml")
	} else if p, ok := aliases[r.URL.Path]; ok {
		r.URL.Path = p
	}
	f.Handler.ServeHTTP(w, r)
}

type staticFileSystem struct {
	http.FileSystem
}

func (fs staticFileSystem) Open(name string) (http.File, error) {
	f, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	d, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if d.IsDir() {
		index, err := fs.FileSystem.Open(strings.TrimSuffix(name, "/") + "/index.html")
		if err != nil {
			return nil, os.ErrNotExist
		}
		index.Close()
	}
	return f, nil
}

func startHTTPRedirectionServer() {
	if config.HTTP.TLS.RedirectFromHTTP == "" {
		return
	}
	go func() {
		_, port, _ := net.SplitHostPort(config.HTTP.Listen)
		err := http.ListenAndServe(config.HTTP.TLS.RedirectFromHTTP, &httpRedirectionHandler{port})
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start HTTP->HTTPS redirection")
		}
	}()
}

type httpRedirectionHandler struct {
	port string
}

func (h *httpRedirectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	url := r.URL
	url.Scheme = "https"
	url.Host, _, _ = net.SplitHostPort(r.Host)
	if h.port != "443" {
		url.Host = net.JoinHostPort(url.Host, h.port)
	}
	http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
}
