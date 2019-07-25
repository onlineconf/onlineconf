package main

import (
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
	"gitlab.corp.mail.ru/mydev/onlineconf/admin/go/admin"
	. "gitlab.corp.mail.ru/mydev/onlineconf/admin/go/common"
	"gitlab.corp.mail.ru/mydev/onlineconf/admin/go/resolver"
)

func main() {
	r := mux.NewRouter()

	registerPProf(r.PathPrefix("/debug/pprof/").Subrouter())

	resolverRouter := r.PathPrefix("/client/").Subrouter()
	resolverRouter.Use(resolver.AuthMiddleware)
	resolver.RegisterRoutes(resolverRouter)

	adminRouter := r.PathPrefix("/").Subrouter()
	adminRouter.Use(admin.AuthMiddleware)
	admin.RegisterRoutes(adminRouter)

	registerStaticFileServer(r.PathPrefix("/").Subrouter())

	handler := hlog.AccessHandler(writeAccessLog)(r)
	if AdminConfig.HTTP.BehindReverseProxy {
		handler = handlers.ProxyHeaders(handler)
	}
	handler = hlog.RequestIDHandler("reqid", "")(handler)
	handler = hlog.NewHandler(log.Logger)(handler)

	server := http.Server{
		Addr:    AdminConfig.HTTP.Listen,
		Handler: handler,
	}
	if AdminConfig.HTTP.TLS.Cert != "" {
		startHTTPRedirectionServer()
		if err := server.ListenAndServeTLS(AdminConfig.HTTP.TLS.Cert, AdminConfig.HTTP.TLS.Key); err != nil {
			log.Fatal().Err(err).Msg("failed to start HTTPS server")
		}
	} else {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("failed to start HTTP server")
		}
	}
}

func writeAccessLog(r *http.Request, status, size int, duration time.Duration) {
	l := hlog.FromRequest(r).Info().
		Str("remote", strings.Split(r.RemoteAddr, ":")[0])
	if r.URL.User != nil {
		l = l.Str("username", r.URL.User.Username())
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
	if AdminConfig.HTTP.StaticRoot != "" {
		handler := &staticFileHandler{http.FileServer(staticFileSystem{http.Dir(AdminConfig.HTTP.StaticRoot)})}
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
	if AdminConfig.HTTP.TLS.RedirectFromHTTP == "" {
		return
	}
	go func() {
		_, port, _ := net.SplitHostPort(AdminConfig.HTTP.Listen)
		err := http.ListenAndServe(AdminConfig.HTTP.TLS.RedirectFromHTTP, &httpRedirectionHandler{port})
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
