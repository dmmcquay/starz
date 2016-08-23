package starz

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/github"
	"github.com/gorilla/sessions"
	githuboauth "golang.org/x/oauth2/github"

	"golang.org/x/oauth2"
)

var store *sessions.CookieStore

var (
	oauthConf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"user:email"},
		Endpoint:     githuboauth.Endpoint,
	}
	oauthStateString = "thisshouldberandom"
)

var Version string = "dev"

var start time.Time

type failure struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func NewFailure(msg string) *failure {
	return &failure{
		Success: false,
		Error:   msg,
	}
}

type Item struct {
	Repo  string `json:"name"`
	Stars int    `json:"stargazers_count"`
}

type Server struct {
	ClientID     string
	ClientSecret string
	ApiToken     string
	CookieSecret string
}

func init() {
	log.SetFlags(log.Ltime)
	start = time.Now()
}

func NewServer(sm *http.ServeMux, clientId, clientSecret, apiToken, cookieSecret, static string) *Server {
	server := &Server{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		ApiToken:     apiToken,
		CookieSecret: cookieSecret,
	}
	addRoutes(sm, server, static)
	return server
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	oauthConf.ClientID = s.ClientID
	oauthConf.ClientSecret = s.ClientSecret
	url := oauthConf.AuthCodeURL(oauthStateString, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *Server) gitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := oauthConf.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get("")
	if err != nil {
		log.Printf("client.Users.Get() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	session, err := store.Get(r, "creds")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = true
	session.Values["uname"] = user.Login
	if err := session.Save(r, w); err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	http.Redirect(w, r, "/static/", http.StatusTemporaryRedirect)
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "creds")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loggedIn := session.Values["authenticated"]; loggedIn != true {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	default:
		b, _ := json.Marshal(NewFailure("Allowed method: GET"))
		http.Error(w, string(b), http.StatusBadRequest)
		return
	case "GET":
		searchreq := r.URL.Path[len(prefix["list"]):]
		if len(searchreq) == 0 {
			b, _ := json.Marshal(NewFailure("url could not be parsed"))
			http.Error(w, string(b), http.StatusBadRequest)
			return
		}
		if searchreq[len(searchreq)-1] != '/' {
			http.Redirect(w, r, prefix["list"]+searchreq+"/", http.StatusMovedPermanently)
			return
		}
		searchReqParsed := strings.Split(searchreq, "/")
		client := github.NewClient(nil)
		if s.ApiToken != "" {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: s.ApiToken},
			)
			tc := oauth2.NewClient(oauth2.NoContext, ts)
			client = github.NewClient(tc)
		}
		opt := &github.RepositoryListOptions{}
		repos, _, err := client.Repositories.List(searchReqParsed[0], opt)
		if err != nil {
			b, _ := json.Marshal(NewFailure("user could not be found"))
			http.Error(w, string(b), http.StatusBadRequest)
			return
		}
		var items []Item
		for _, i := range repos {
			items = append(items, Item{*i.Name, *i.StargazersCount})
		}

		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			b, _ := json.Marshal(NewFailure(err.Error()))
			http.Error(w, string(b), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) {
	output := struct {
		Auth bool `json:"auth"`
	}{
		Auth: false,
	}
	w.Header().Set("Content-Type", "application/json")
	session, err := store.Get(r, "creds")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loggedIn := session.Values["authenticated"]; loggedIn == true {
		output.Auth = true
		json.NewEncoder(w).Encode(output)
		return
	}
	b, _ := json.Marshal(output)
	http.Error(w, string(b), http.StatusUnauthorized)
}

func (s *Server) logout(w http.ResponseWriter, req *http.Request) {
	session, err := store.Get(req, "creds")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	delete(session.Values, "authenticated")
	delete(session.Values, "uname")
	session.Save(req, w)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func (s *Server) serverInfo(w http.ResponseWriter, req *http.Request) {
	output := struct {
		Version string `json:"version"`
		Start   string `json:"start"`
		Uptime  string `json:"uptime"`
	}{
		Version: Version,
		Start:   start.Format("2006-01-02 15:04:05"),
		Uptime:  time.Since(start).String(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(output)
}

func (s *Server) plist(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "creds")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if loggedIn := session.Values["authenticated"]; loggedIn != true {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	data, err := Asset("static/list.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	req := bytes.NewReader(data)
	io.Copy(w, req)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}
