// Package api puts the bar on HTTP.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/AndyCross/dotnet-go-layout/go/internal/store"
	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

type server struct {
	menu  *tab.Menu
	store store.Store
	mu    sync.Mutex // one bartender: every change to a tab is a read, a change and a write
}

// New returns the API's handler.
func New(menu *tab.Menu, s store.Store) http.Handler {
	srv := &server{menu: menu, store: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /menu", srv.getMenu)
	mux.HandleFunc("GET /tabs", srv.listTabs)
	mux.HandleFunc("POST /tabs", srv.openTab)
	mux.HandleFunc("GET /tabs/{name}", srv.getTab)
	mux.HandleFunc("GET /tabs/{name}/chalk", srv.chalk)
	mux.HandleFunc("POST /tabs/{name}/rounds", srv.round)
	mux.HandleFunc("POST /tabs/{name}/penultima", srv.penultima)
	mux.HandleFunc("DELETE /tabs/{name}", srv.settle)
	return mux
}

func (s *server) getMenu(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.menu.Items())
}

func (s *server) listTabs(w http.ResponseWriter, r *http.Request) {
	tabs, err := s.store.List()
	if err != nil {
		fail(w, err)
		return
	}
	bills := make([]tab.Bill, 0, len(tabs))
	for _, t := range tabs {
		bills = append(bills, t.Bill(s.menu))
	}
	writeJSON(w, http.StatusOK, bills)
}

func (s *server) openTab(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name string `json:"name"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		problem(w, http.StatusBadRequest, "a tab needs a name")
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := s.store.Get(name); err == nil {
		problem(w, http.StatusConflict, name+" already has a tab open")
		return
	}
	t := tab.New(name)
	if err := s.store.Put(t); err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Location", "/tabs/"+url.PathEscape(name))
	writeJSON(w, http.StatusCreated, t.Bill(s.menu))
}

func (s *server) getTab(w http.ResponseWriter, r *http.Request) {
	t, err := s.store.Get(r.PathValue("name"))
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t.Bill(s.menu))
}

func (s *server) chalk(w http.ResponseWriter, r *http.Request) {
	t, err := s.store.Get(r.PathValue("name"))
	if err != nil {
		fail(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(tab.Chalk(t.Bill(s.menu))))
}

func (s *server) round(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Items []string `json:"items"`
	}
	if !readJSON(w, r, &in) {
		return
	}
	s.change(w, r.PathValue("name"), func(t *tab.Tab) (any, error) {
		if err := t.Order(s.menu, in.Items...); err != nil {
			return nil, err
		}
		return t.Bill(s.menu), nil
	})
}

func (s *server) penultima(w http.ResponseWriter, r *http.Request) {
	s.change(w, r.PathValue("name"), func(t *tab.Tab) (any, error) {
		it, err := t.Penultima(s.menu)
		if err != nil {
			return nil, err
		}
		return struct {
			Poured string   `json:"poured"`
			Bill   tab.Bill `json:"bill"`
		}{it.ID, t.Bill(s.menu)}, nil
	})
}

func (s *server) settle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.store.Get(r.PathValue("name"))
	if err == nil {
		err = s.store.Delete(t.Name)
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t.Bill(s.menu))
}

// change loads a tab, applies f and saves it, under the one lock.
func (s *server) change(w http.ResponseWriter, name string, f func(*tab.Tab) (any, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, err := s.store.Get(name)
	if err != nil {
		fail(w, err)
		return
	}
	out, err := f(t)
	if err == nil {
		err = s.store.Put(t)
	}
	if err != nil {
		fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// fail turns the rules' errors into answers.
func fail(w http.ResponseWriter, err error) {
	var unknown tab.UnknownItemError
	switch {
	case errors.Is(err, store.ErrNotFound):
		problem(w, http.StatusNotFound, err.Error())
	case errors.As(err, &unknown), errors.Is(err, tab.ErrEmptyRound):
		problem(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, tab.ErrNothingToRepeat):
		problem(w, http.StatusConflict, err.Error())
	default:
		problem(w, http.StatusInternalServerError, "something went wrong behind the bar")
	}
}

func readJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		problem(w, http.StatusBadRequest, "the body must be JSON")
		return false
	}
	return true
}

func problem(w http.ResponseWriter, status int, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{
		"title":  http.StatusText(status),
		"status": status,
		"detail": detail,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
