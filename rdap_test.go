package indeed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func rdapHandler(w http.ResponseWriter, r *http.Request) {
	var name string
	switch r.URL.Path {
	case "/domain/example.com":
		name = "testdata/rdap-example-com.json"
	case "/domain/example.net":
		name = "testdata/rdap-example-net.json"
	case "/domain/example.org":
		name = "testdata/rdap-example-org.json"
	default:
		msg := fmt.Sprintf("unexpected path %q", r.URL.Path)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	f, err := os.Open(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func TestRDAPSingle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(rdapHandler))
	defer server.Close()

	client := NewRDAPClient(server.URL)

	got, err := client.LookupDomain(context.Background(), "example.com")
	if err != nil {
		t.Fatal(err)
	}

	want := &RDAPDomain{
		Name: "EXAMPLE.COM",
		Events: []RDAPEvent{
			{
				Action: "registration",
				Date:   time.Date(1995, 8, 14, 4, 0, 0, 0, time.UTC),
			},
			{
				Action: "expiration",
				Date:   time.Date(2024, 8, 13, 4, 0, 0, 0, time.UTC),
			},
			{
				Action: "last changed",
				Date:   time.Date(2023, 8, 14, 7, 1, 38, 0, time.UTC),
			},
			{
				Action: "last update of RDAP database",
				Date:   time.Date(2023, 8, 19, 8, 16, 0, 0, time.UTC),
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func TestRDAPMultiple(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(rdapHandler))
	defer server.Close()

	client := NewRDAPClient(server.URL)

	want := []string{"example.com", "example.net", "example.org"}
	domains, err := client.LookupDomains(context.Background(), want)
	if err != nil {
		t.Fatal(err)
	}

	got := make([]string, len(domains))
	for i, domain := range domains {
		got[i] = strings.ToLower(domain.Name)
	}
	sort.Strings(got)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}
