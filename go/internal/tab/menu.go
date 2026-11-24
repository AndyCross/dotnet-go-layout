// Package tab is the bar's rules: what is on the menu, what has been chalked
// up against each name, and what that comes to.
package tab

import (
	_ "embed"
	"encoding/json"
	"sync"
)

// menu.json is a copy of ../../../contract/menu.json. go:embed cannot reach
// outside the module, so the copy lives here and TestMenuMatchesContract
// fails if the two drift apart.
//
//go:embed menu.json
var menuJSON []byte

// Item is one thing the bar sells. Price is in euro cents.
type Item struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kind  string `json:"kind"` // "drink" or "tapa"
	Price int    `json:"price"`
}

// Menu is the list of items in the order they are chalked on the board.
type Menu struct {
	items []Item
	byID  map[string]int
}

// NewMenu builds a menu from items, keeping their order.
func NewMenu(items []Item) *Menu {
	m := &Menu{items: items, byID: make(map[string]int, len(items))}
	for i, it := range items {
		m.byID[it.ID] = i
	}
	return m
}

// DefaultMenu is the embedded menu, parsed once.
var DefaultMenu = sync.OnceValue(func() *Menu {
	var items []Item
	if err := json.Unmarshal(menuJSON, &items); err != nil {
		panic("tab: embedded menu.json: " + err.Error())
	}
	return NewMenu(items)
})

// Items returns the menu in board order.
func (m *Menu) Items() []Item { return m.items }

// Find looks an item up by its id.
func (m *Menu) Find(id string) (Item, bool) {
	i, ok := m.byID[id]
	if !ok {
		return Item{}, false
	}
	return m.items[i], true
}
