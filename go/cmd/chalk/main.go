// Command chalk keeps the bar's tabs from the terminal, in a JSON file.
//
//	chalk [-file chalk.json] menu
//	chalk open <name>
//	chalk order <name> <item>...
//	chalk penultima <name>
//	chalk show [name]
//	chalk settle <name>
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/AndyCross/dotnet-go-layout/go/internal/store"
	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

const usage = `usage: chalk [-file chalk.json] <command>

  menu                    what the bar sells
  open <name>             open a tab
  order <name> <item>...  chalk up a round
  penultima <name>        one more of the usual
  show [name]             one tab, or every open tab
  settle <name>           price it up and wipe it off
`

func main() {
	file := flag.String("file", "chalk.json", "where the tabs are kept")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()
	if err := run(store.File{Path: *file}, tab.DefaultMenu(), flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "chalk:", err)
		os.Exit(1)
	}
}

func run(s store.Store, m *tab.Menu, args []string) error {
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	cmd, args := args[0], args[1:]
	needName := func() (string, error) {
		if len(args) == 0 {
			return "", fmt.Errorf("%s needs a name", cmd)
		}
		return args[0], nil
	}

	switch cmd {
	case "menu":
		for _, it := range m.Items() {
			fmt.Printf("  %-14s %-14s %-6s %8s\n", it.ID, it.Name, it.Kind, tab.Euros(it.Price))
		}
		return nil

	case "open":
		name, err := needName()
		if err != nil {
			return err
		}
		if _, err := s.Get(name); err == nil {
			return fmt.Errorf("%s already has a tab open", name)
		}
		return s.Put(tab.New(name))

	case "order":
		name, err := needName()
		if err != nil {
			return err
		}
		t, err := s.Get(name)
		if err != nil {
			return err
		}
		if err := t.Order(m, args[1:]...); err != nil {
			return err
		}
		fmt.Print(tab.Chalk(t.Bill(m)))
		return s.Put(t)

	case "penultima":
		name, err := needName()
		if err != nil {
			return err
		}
		t, err := s.Get(name)
		if err != nil {
			return err
		}
		it, err := t.Penultima(m)
		if err != nil {
			return err
		}
		fmt.Printf("la penúltima: %s\n", it.Name)
		fmt.Print(tab.Chalk(t.Bill(m)))
		return s.Put(t)

	case "show":
		if len(args) > 0 {
			t, err := s.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Print(tab.Chalk(t.Bill(m)))
			return nil
		}
		tabs, err := s.List()
		if err != nil {
			return err
		}
		if len(tabs) == 0 {
			fmt.Println("the bar is empty")
		}
		for _, t := range tabs {
			fmt.Print(tab.Chalk(t.Bill(m)))
		}
		return nil

	case "settle":
		name, err := needName()
		if err != nil {
			return err
		}
		t, err := s.Get(name)
		if err != nil {
			return err
		}
		fmt.Print(tab.Chalk(t.Bill(m)))
		fmt.Println("settled. ¡Salud!")
		return s.Delete(name)
	}
	return errors.New("no command called " + cmd)
}
