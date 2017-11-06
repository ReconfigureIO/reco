package proxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxy(t *testing.T) {
	var handler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		fmt.Fprintf(w, "Hello %s", name)
	}
	fakeserver := httptest.NewServer(handler)
	addr := strings.TrimPrefix(fakeserver.URL, "http://")

	server, err := New(addr)
	if err != nil {
		t.Fatal(err)
	}

	if err := server.Start(); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(fmt.Sprintf("http://%s?name=proxy", server.Info().Listen.String()))
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected %d status code", http.StatusOK)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if expected := "Hello proxy"; string(b) != expected {
		t.Fatalf("Expected %s found %s", expected, string(b))
	}
}
