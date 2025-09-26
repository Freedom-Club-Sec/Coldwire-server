package httpserver

import (
	"fmt"
    "mime"
    "path/filepath"
	"net/http"
    "embed"

	"github.com/Freedom-Club-Sec/Coldwire-server/internal/authenticate"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/config"
	"github.com/Freedom-Club-Sec/Coldwire-server/internal/data"
)

type Server struct {
	addr   string
	mux    *http.ServeMux
	Cfg    *config.Config
	DbSvcs *DBServices
}

type DBServices struct {
	UserService *authenticate.UserService
	DataService *data.DataService
}

//go:embed web/**
var webFiles embed.FS

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/authenticate/init", s.authenticateInitHandler)
	s.mux.HandleFunc("/authenticate/verify", s.authenticateVerificationHandler)

	s.mux.Handle("/data/longpoll", s.jwtMiddleware(http.HandlerFunc(s.dataLongpollHandler)))
	s.mux.Handle("/data/send", s.jwtMiddleware(http.HandlerFunc(s.newDataHandler)))

	s.mux.HandleFunc("/federation/info", s.federationInfoHandler)
	s.mux.HandleFunc("/federation/send", s.federationSendHandler)


    s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path
        if path == "/" {
            path = "/index.html"
        }

        // Serve file from embedded FS
        data, err := webFiles.ReadFile("web" + path)
        if err != nil {
            http.NotFound(w, r)
            return
        }

        ext := filepath.Ext(path)
        if ctype := mime.TypeByExtension(ext); ctype != "" {
            w.Header().Set("Content-Type", ctype)
        } else {
            w.Header().Set("Content-Type", "application/octet-stream")
        }

        w.Write(data)
    })

}

func New(host string, port int, cfg *config.Config, dbSvcs *DBServices) *Server {
	mux := http.NewServeMux()

	srv := &Server{
		addr:   fmt.Sprintf("%s:%d", host, port),
		mux:    mux,
		Cfg:    cfg,
		DbSvcs: dbSvcs,
	}
	srv.registerRoutes()

	return srv
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) Addr() string {
	return s.addr
}
