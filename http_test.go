package indeed

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPBody(t *testing.T) {
	v := url.Values{}
	v.Set(paramQ, "example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.RawQuery = v.Encode()

	res, serverURL := testLookup(req)
	defer res.Body.Close()

	var got RSSFeed
	if err := xml.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	link, err := url.JoinPath(serverURL, "domain", strings.ToUpper(v.Get(paramQ)))
	if err != nil {
		t.Fatal(err)
	}

	want := RSSFeed{
		XMLName:     xml.Name{Local: "rss"},
		Version:     "2.0",
		Title:       "Domain Events",
		Link:        "/?q=example.com",
		Description: "Domain events for: example.com",
		Items: []RSSItem{
			{
				Link:        link,
				Description: "example.com: expiration",
				PubDate:     RSSTime{time.Date(2024, 8, 13, 4, 0, 0, 0, time.UTC)},
			},
			{
				Link:        link,
				Description: "example.com: last update of RDAP database",
				PubDate:     RSSTime{time.Date(2023, 8, 19, 8, 16, 0, 0, time.UTC)},
			},
			{
				Link:        link,
				Description: "example.com: last changed",
				PubDate:     RSSTime{time.Date(2023, 8, 14, 7, 1, 0, 0, time.UTC)},
			},
			{
				Link:        link,
				Description: "example.com: registration",
				PubDate:     RSSTime{time.Date(1995, 8, 14, 4, 0, 0, 0, time.UTC)},
			},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func TestHTTPStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		method string
		q      []string
		want   int
	}{
		{
			"ok",
			http.MethodGet,
			[]string{"example.com"},
			http.StatusOK,
		},
		{
			"invalid method",
			http.MethodPost,
			[]string{"example.com"},
			http.StatusMethodNotAllowed,
		},
		{
			"missing query parameter",
			http.MethodGet,
			[]string{},
			http.StatusBadRequest,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			v := url.Values{}
			for _, name := range tc.q {
				v.Add(paramQ, name)
			}

			req := httptest.NewRequest(tc.method, "/", nil)
			req.URL.RawQuery = v.Encode()

			res, _ := testLookup(req)
			defer res.Body.Close()

			if res.StatusCode != tc.want {
				t.Fatalf("got: %d; want: %d", res.StatusCode, tc.want)
			}
		})
	}
}

func testLookup(r *http.Request) (*http.Response, string) {
	server := httptest.NewServer(http.HandlerFunc(rdapHandler))
	defer server.Close()

	h := NewHandler(NewRDAPClient(server.URL))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	return w.Result(), server.URL
}
