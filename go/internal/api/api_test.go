package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AndyCross/dotnet-go-layout/go/internal/store"
	"github.com/AndyCross/dotnet-go-layout/go/internal/tab"
)

// contractDir is shared with the .NET tests. Tests run in their package's
// directory, so a relative path reaches it.
const contractDir = "../../../contract"

type step struct {
	Say      string          `json:"say"`
	Method   string          `json:"method"`
	Path     string          `json:"path"`
	Body     json.RawMessage `json:"body"`
	Raw      string          `json:"raw"`
	Status   int             `json:"status"`
	Location string          `json:"location"`
	JSON     json.RawMessage `json:"json"`
	JSONFile string          `json:"jsonFile"`
	Text     *string         `json:"text"`
}

func TestContract(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(contractDir, "http.json"))
	if err != nil {
		t.Fatal(err)
	}
	var contract struct{ Steps []step }
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(New(tab.DefaultMenu(), store.NewMemory()))
	defer srv.Close()

	for i, s := range contract.Steps {
		var body io.Reader
		switch {
		case s.Body != nil:
			body = bytes.NewReader(s.Body)
		case s.Raw != "":
			body = bytes.NewBufferString(s.Raw)
		}
		req, _ := http.NewRequest(s.Method, srv.URL+s.Path, body)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("step %d (%s): %v", i, s.Say, err)
		}
		got, _ := io.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode != s.Status {
			t.Fatalf("step %d (%s): status %d, want %d\n%s", i, s.Say, res.StatusCode, s.Status, got)
		}
		if s.Location != "" && res.Header.Get("Location") != s.Location {
			t.Errorf("step %d (%s): Location %q, want %q", i, s.Say, res.Header.Get("Location"), s.Location)
		}
		want := []byte(s.JSON)
		if s.JSONFile != "" {
			if want, err = os.ReadFile(filepath.Join(contractDir, s.JSONFile)); err != nil {
				t.Fatal(err)
			}
		}
		switch {
		case s.Text != nil:
			if string(got) != *s.Text {
				t.Errorf("step %d (%s): got\n%s\nwant\n%s", i, s.Say, got, *s.Text)
			}
		case want != nil:
			if !sameJSON(t, got, want) {
				t.Errorf("step %d (%s): got\n%s\nwant\n%s", i, s.Say, got, want)
			}
		}
	}
}

func sameJSON(t *testing.T, a, b []byte) bool {
	var x, y any
	if err := json.Unmarshal(a, &x); err != nil {
		t.Fatalf("response is not JSON: %s", a)
	}
	if err := json.Unmarshal(b, &y); err != nil {
		t.Fatal(err)
	}
	return reflect.DeepEqual(x, y)
}
