package starz

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/sessions"
)

var prefix map[string]string

func addRoutes(sm *http.ServeMux, server *Server, staticFiles string) {
	prefix = map[string]string{
		"info":      "/info/",
		"static":    "/static/s/",
		"protected": "/static/",
		"login":     "/api/v0/login/",
		"logout":    "/api/v0/logout/",
		"list":      "/api/v0/list/",
		"github":    "/api/v0/github_oauth_cb/",
		"auth":      "/api/v0/auth/",
		"health":    "/healthz",
	}

	if staticFiles == "" {
		sm.Handle(
			prefix["static"],
			http.FileServer(
				&assetfs.AssetFS{
					Asset:     Asset,
					AssetDir:  AssetDir,
					AssetInfo: AssetInfo,
				},
			),
		)
		sm.HandleFunc(
			"/",
			func(w http.ResponseWriter, req *http.Request) {
				data, err := Asset("static/s/index.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				r := bytes.NewReader(data)
				io.Copy(w, r)
			},
		)
	} else {
		sm.Handle(
			prefix["static"],
			http.StripPrefix(
				prefix["static"],
				http.FileServer(http.Dir(staticFiles)),
			),
		)
		sm.HandleFunc(
			"/",
			func(w http.ResponseWriter, req *http.Request) {
				http.ServeFile(w, req, filepath.Join(staticFiles, "index.html"))
			},
		)
	}

	store = sessions.NewCookieStore([]byte(server.CookieSecret))
	sm.HandleFunc(prefix["protected"], server.plist)
	sm.HandleFunc(prefix["info"], server.serverInfo)
	sm.HandleFunc(prefix["login"], server.login)
	sm.HandleFunc(prefix["logout"], server.logout)
	sm.HandleFunc(prefix["list"], server.list)
	sm.HandleFunc(prefix["github"], server.gitHubCallback)
	sm.HandleFunc(prefix["auth"], server.auth)
	sm.HandleFunc(prefix["health"], server.health)
}
