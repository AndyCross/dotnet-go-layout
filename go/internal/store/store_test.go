package store

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

// Both stores must behave the same, so the same test runs against each.
func TestStores(t *testing.T) {
	stores := map[string]Store{
		"memory": NewMemory(),
		"file":   File{Path: filepath.Join(t.TempDir(), "chalk.json")},
	}
	for name, s := range stores {
		t.Run(name, func(t *testing.T) {
			if _, err := s.Get("Mari"); !errors.Is(err, ErrNotFound) {
				t.Fatalf("Get on empty store: %v", err)
			}
			mari := tab.New("Mari")
			mari.Marks["fino"] = 2
			if err := s.Put(mari); err != nil {
				t.Fatal(err)
			}
			if err := s.Put(tab.New("Antonio")); err != nil {
				t.Fatal(err)
			}
			got, err := s.Get("Mari")
			if err != nil || got.Marks["fino"] != 2 {
				t.Fatalf("Get = %+v, %v", got, err)
			}
			all, _ := s.List()
			if len(all) != 2 || all[0].Name != "Antonio" {
				t.Fatalf("List = %+v", all)
			}
			if err := s.Delete("Mari"); err != nil {
				t.Fatal(err)
			}
			if err := s.Delete("Mari"); !errors.Is(err, ErrNotFound) {
				t.Fatalf("second Delete: %v", err)
			}
		})
	}
}
