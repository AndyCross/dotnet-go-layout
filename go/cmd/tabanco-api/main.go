// Command tabanco-api serves the bar over HTTP, with tabs kept in memory.
package main

import (
	"cmp"
	"log"
	"net/http"
	"os"

	"github.com/AndyCross/dotnet-go-layout/go/internal/api"
	"github.com/AndyCross/dotnet-go-layout/go/internal/store"
	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

func main() {
	addr := "localhost:" + cmp.Or(os.Getenv("PORT"), "8080")
	log.Printf("tabanco open on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, api.New(tab.DefaultMenu(), store.NewMemory())))
}
