package tab

import (
	"errors"
	"fmt"
	"strings"
)

// ErrEmptyRound is returned for a round with nothing in it.
var ErrEmptyRound = errors.New("a round needs at least one item")

// ErrNothingToRepeat is returned by Penultima before any drink is chalked.
var ErrNothingToRepeat = errors.New("nothing to repeat yet: order a drink first")

// UnknownItemError names an item that is not on the menu.
type UnknownItemError struct{ ID string }

func (e UnknownItemError) Error() string {
	return fmt.Sprintf("nothing called %q on the menu", e.ID)
}

// Tab is what has been chalked on the bar in front of one person: a count of
// marks per menu item.
type Tab struct {
	Name  string         `json:"name"`
	Marks map[string]int `json:"marks"`
}

// New opens an empty tab.
func New(name string) *Tab {
	return &Tab{Name: name, Marks: map[string]int{}}
}

// Order chalks a round. Every item is checked before any is marked, so a
// round with one unknown item marks nothing.
func (t *Tab) Order(m *Menu, ids ...string) error {
	if len(ids) == 0 {
		return ErrEmptyRound
	}
	for _, id := range ids {
		if _, ok := m.Find(id); !ok {
			return UnknownItemError{ID: id}
		}
	}
	for _, id := range ids {
		t.Marks[id]++
	}
	return nil
}

// Penultima orders one more of the drink with the most marks. Nobody in a
// tabanco orders the last one, only the one before it. A tie goes to the
// drink higher up the board.
func (t *Tab) Penultima(m *Menu) (Item, error) {
	var best Item
	most := 0
	for _, it := range m.Items() {
		if it.Kind == "drink" && t.Marks[it.ID] > most {
			best, most = it, t.Marks[it.ID]
		}
	}
	if most == 0 {
		return Item{}, ErrNothingToRepeat
	}
	t.Marks[best.ID]++
	return best, nil
}

// Line is one item on a bill.
type Line struct {
	Item  string `json:"item"`
	Name  string `json:"name"`
	Count int    `json:"count"`
	Cents int    `json:"cents"`
}

// Bill is a tab priced against a menu, in board order.
type Bill struct {
	Name  string `json:"name"`
	Lines []Line `json:"lines"`
	Total int    `json:"total"`
}

// Bill prices the tab.
func (t *Tab) Bill(m *Menu) Bill {
	b := Bill{Name: t.Name, Lines: []Line{}}
	for _, it := range m.Items() {
		n := t.Marks[it.ID]
		if n == 0 {
			continue
		}
		b.Lines = append(b.Lines, Line{Item: it.ID, Name: it.Name, Count: n, Cents: n * it.Price})
		b.Total += n * it.Price
	}
	return b
}

// Euros formats cents the Spanish way, "16,50 €", by hand: Go's standard
// library has no locale-aware number formatting.
func Euros(cents int) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

// tally draws n as chalk marks, in gates of five.
func tally(n int) string {
	var gates []string
	for ; n >= 5; n -= 5 {
		gates = append(gates, "||||/")
	}
	if n > 0 {
		gates = append(gates, strings.Repeat("|", n))
	}
	return strings.Join(gates, " ")
}

// Chalk draws the bill as it would look on the bar.
func Chalk(b Bill) string {
	var s strings.Builder
	s.WriteString(b.Name + "\n")
	if len(b.Lines) == 0 {
		s.WriteString("  (nothing chalked yet)\n")
		return s.String()
	}
	for _, l := range b.Lines {
		fmt.Fprintf(&s, "  %-14s %-16s %3d %10s\n", l.Name, tally(l.Count), l.Count, Euros(l.Cents))
	}
	s.WriteString("  " + strings.Repeat("-", 46) + "\n")
	fmt.Fprintf(&s, "  %-35s %10s\n", "total", Euros(b.Total))
	return s.String()
}
