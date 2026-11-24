// Package store keeps open tabs between requests or between runs of the CLI.
package store

import (
	"encoding/json"
	"errors"
	"os"
	"sort"
	"sync"

	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

// ErrNotFound means no tab is open under that name.
var ErrNotFound = errors.New("no tab open under that name")

// Store is where open tabs live. It is declared here, beside the two
// implementations, because both programs need the same small set.
type Store interface {
	Get(name string) (*tab.Tab, error)
	Put(t *tab.Tab) error
	Delete(name string) error
	List() ([]*tab.Tab, error)
}

// Memory is a Store for the API: gone when the process stops.
type Memory struct {
	mu   sync.Mutex
	tabs map[string]*tab.Tab
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory { return &Memory{tabs: map[string]*tab.Tab{}} }

func (m *Memory) Get(name string) (*tab.Tab, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tabs[name]
	if !ok {
		return nil, ErrNotFound
	}
	return clone(t), nil
}

func (m *Memory) Put(t *tab.Tab) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tabs[t.Name] = clone(t)
	return nil
}

func (m *Memory) Delete(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tabs[name]; !ok {
		return ErrNotFound
	}
	delete(m.tabs, name)
	return nil
}

func (m *Memory) List() ([]*tab.Tab, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*tab.Tab, 0, len(m.tabs))
	for _, t := range m.tabs {
		out = append(out, clone(t))
	}
	sortByName(out)
	return out, nil
}

// File is a Store for the CLI: one JSON file, read and written whole on
// every call, which is plenty for one bar.
type File struct{ Path string }

func (f File) load() (map[string]*tab.Tab, error) {
	tabs := map[string]*tab.Tab{}
	data, err := os.ReadFile(f.Path)
	if errors.Is(err, os.ErrNotExist) {
		return tabs, nil
	}
	if err != nil {
		return nil, err
	}
	return tabs, json.Unmarshal(data, &tabs)
}

func (f File) save(tabs map[string]*tab.Tab) error {
	data, err := json.MarshalIndent(tabs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.Path, append(data, '\n'), 0o644)
}

func (f File) Get(name string) (*tab.Tab, error) {
	tabs, err := f.load()
	if err != nil {
		return nil, err
	}
	t, ok := tabs[name]
	if !ok {
		return nil, ErrNotFound
	}
	return t, nil
}

func (f File) Put(t *tab.Tab) error {
	tabs, err := f.load()
	if err != nil {
		return err
	}
	tabs[t.Name] = t
	return f.save(tabs)
}

func (f File) Delete(name string) error {
	tabs, err := f.load()
	if err != nil {
		return err
	}
	if _, ok := tabs[name]; !ok {
		return ErrNotFound
	}
	delete(tabs, name)
	return f.save(tabs)
}

func (f File) List() ([]*tab.Tab, error) {
	tabs, err := f.load()
	if err != nil {
		return nil, err
	}
	out := make([]*tab.Tab, 0, len(tabs))
	for _, t := range tabs {
		out = append(out, t)
	}
	sortByName(out)
	return out, nil
}

func clone(t *tab.Tab) *tab.Tab {
	c := tab.New(t.Name)
	for k, v := range t.Marks {
		c.Marks[k] = v
	}
	return c
}

func sortByName(ts []*tab.Tab) {
	sort.Slice(ts, func(i, j int) bool { return ts[i].Name < ts[j].Name })
}
