package starz

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
)

var server = Server{}

func init() {
	store = sessions.NewCookieStore([]byte(""))
}

func TestHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.health)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestInfo(t *testing.T) {
	req, err := http.NewRequest("GET", "/info", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.serverInfo)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestListWithAuth(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/list/dmmcquay/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.list)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "dmmcquay"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("could not store session info")
		return
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestListWithoutAuth(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/list/dmmcquay/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.list)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = false
	session.Values["uname"] = "bobo"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("could not store session info")
		return
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}
}

func TestListWrongMethod(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v0/list/dmmcquay/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.list)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "bobo"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("could not store session info")
		return
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestListBadURL(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/list/ ", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.list)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "bobo"
	if err := session.Save(req, rr); err != nil {
		t.Errorf("could not store session info")
		return
	}
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusMovedPermanently)
	}
}

func TestLogin(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/login/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.login)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}
}

func TestAuthNotAuthed(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/auth/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.auth)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
	expected := `{"auth":false}`
	actual := strings.Trim(rr.Body.String(), "\n ")
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}
}

func TestAuthAuthed(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/auth/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.auth)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "bobo"
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"auth":true}`
	actual := strings.Trim(rr.Body.String(), "\n ")
	if actual != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}
}

func TestLogout(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v0/logout/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.logout)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "bobo"
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusSeeOther)
	}
	if session.Values["authenticated"] != nil {
		t.Errorf("handler returned unexpected body: got %v want %v",
			session.Values["authenticated"], nil)
	}
}

func TestPlistNotAuthed(t *testing.T) {
	req, err := http.NewRequest("GET", "/static/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.plist)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}
}

func TestPlistAuthed(t *testing.T) {
	req, err := http.NewRequest("GET", "/static/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.plist)
	session, _ := store.Get(req, "creds")
	session.Values["authenticated"] = true
	session.Values["uname"] = "bobo"
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestStaticAssests(t *testing.T) {
	req, err := http.NewRequest("GET", "/static/s/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(
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
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAddRoutes(t *testing.T) {
	//addRoutes(sm *http.ServeMux, server *Server, staticFiles string)
	sm := http.NewServeMux()
	_ = NewServer(
		sm,
		"",
		"",
		"",
		"",
		"",
	)
	expected := "/api/v0/github_oauth_cb/"
	if prefix["github"] != expected {
		t.Errorf("prefix didn't get populated correctly: got %v want %v",
			prefix["github"], expected)
	}
}
