package tab

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestMenuMatchesContract(t *testing.T) {
	want, err := os.ReadFile("../../../contract/menu.json")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(menuJSON, want) {
		t.Fatal("internal/tab/menu.json differs from contract/menu.json: copy it again")
	}
}

func TestOrderAndBill(t *testing.T) {
	m := DefaultMenu()
	tb := New("Mari")
	if err := tb.Order(m, "fino", "fino", "chicharrones"); err != nil {
		t.Fatal(err)
	}
	b := tb.Bill(m)
	if b.Total != 800 || len(b.Lines) != 2 || b.Lines[0].Item != "fino" || b.Lines[0].Count != 2 {
		t.Fatalf("bill = %+v", b)
	}
}

func TestOrderWithUnknownItemMarksNothing(t *testing.T) {
	m := DefaultMenu()
	tb := New("Mari")
	err := tb.Order(m, "fino", "cerveza")
	var unknown UnknownItemError
	if !errors.As(err, &unknown) || unknown.ID != "cerveza" {
		t.Fatalf("err = %v", err)
	}
	if len(tb.Marks) != 0 {
		t.Fatalf("marks = %v", tb.Marks)
	}
}

func TestPenultima(t *testing.T) {
	m := DefaultMenu()
	tb := New("Paco")
	if _, err := tb.Penultima(m); !errors.Is(err, ErrNothingToRepeat) {
		t.Fatalf("err = %v", err)
	}
	// Tapas never count, and a tie goes to the drink higher on the board.
	_ = tb.Order(m, "chicharrones", "chicharrones", "chicharrones", "oloroso", "manzanilla")
	it, err := tb.Penultima(m)
	if err != nil || it.ID != "manzanilla" {
		t.Fatalf("got %v, %v", it.ID, err)
	}
}

func TestTally(t *testing.T) {
	for n, want := range map[int]string{0: "", 1: "|", 4: "||||", 5: "||||/", 7: "||||/ ||", 10: "||||/ ||||/"} {
		if got := tally(n); got != want {
			t.Errorf("tally(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestChalk(t *testing.T) {
	m := DefaultMenu()
	tb := New("Mari")
	_ = tb.Order(m, "fino", "fino", "fino", "fino", "fino", "fino", "fino", "px", "chicharrones")
	want := "" +
		"Mari\n" +
		"  Fino           ||||/ ||           7    14,00 €\n" +
		"  Pedro Ximénez  |                  1     2,50 €\n" +
		"  Chicharrones   |                  1     4,00 €\n" +
		"  ----------------------------------------------\n" +
		"  total                                  20,50 €\n"
	if got := Chalk(tb.Bill(m)); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
}
