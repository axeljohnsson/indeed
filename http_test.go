package indeed

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	urlpkg "net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPBody(t *testing.T) {
	v := urlpkg.Values{}
	v.Set(paramQ, "example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.RawQuery = v.Encode()

	res := testLookup(req)
	defer res.Body.Close()

	var got RSSFeed
	if err := xml.NewDecoder(res.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}

	link, err := urlpkg.JoinPath(RDAPBaseURL, "domain", strings.ToUpper(v.Get(paramQ)))
	if err != nil {
		t.Fatal(err)
	}

	want := RSSFeed{
		XMLName:     xml.Name{Local: "rss"},
		Version:     "2.0",
		Title:       "Domain Events",
		Link:        "/feed?q=example.com",
		Description: "Domain events for: example.com.",
		Items: []RSSItem{
			{
				Link:        link,
				Description: "example.com: expiration",
				GUID:        "0e7b8746deb1b3df50b53bd3fa1df6f795e130088f3dbee4fbcd559b99ea7e46",
				PubDate:     RSSTime{time.Date(2024, 8, 13, 4, 0, 0, 0, time.UTC)},
			},
			{
				Link:        link,
				Description: "example.com: last changed",
				GUID:        "264aaecf302ed10f175731ded269a76e2ac202212ac70cf6e73977e6ba033f5b",
				PubDate:     RSSTime{time.Date(2023, 8, 14, 7, 1, 0, 0, time.UTC)},
			},
			{
				Link:        link,
				Description: "example.com: registration",
				GUID:        "8c0e7bcead41a573c598c2ab9ae7e95fde486b0d7307b115a1da9b6d6fbb8c4a",
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
		params urlpkg.Values
		want   int
	}{
		{
			"ok",
			http.MethodGet,
			urlpkg.Values{
				paramQ: []string{"example.com"},
			},
			http.StatusOK,
		},
		{
			"invalid method",
			http.MethodPost,
			urlpkg.Values{
				paramQ: []string{"example.com"},
			},
			http.StatusMethodNotAllowed,
		},
		{
			"missing query parameter",
			http.MethodGet,
			urlpkg.Values{},
			http.StatusBadRequest,
		},
		{
			"not found",
			http.MethodGet,
			urlpkg.Values{
				paramQ: []string{"404.com"},
			},
			http.StatusNotFound,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			req.URL.RawQuery = tc.params.Encode()

			res := testLookup(req)
			defer res.Body.Close()

			if res.StatusCode != tc.want {
				t.Fatalf("got: %d; want: %d", res.StatusCode, tc.want)
			}
		})
	}
}

func TestMSM(t *testing.T) {
	tests := []struct {
		name   string
		params urlpkg.Values
		want   int
		err    error
	}{
		{
			"default",
			urlpkg.Values{
				paramQ: []string{"example.com"},
			},
			1,
			nil,
		},
		{
			"msm",
			urlpkg.Values{
				paramMSM: []string{"0"},
			},
			0,
			nil,
		},
		{
			"and operator",
			urlpkg.Values{
				paramQ:  []string{"example.com", "example.net"},
				paramOp: []string{"and"},
			},
			2,
			nil,
		},
		{
			"or operator",
			urlpkg.Values{
				paramQ:  []string{"example.com", "example.net"},
				paramOp: []string{"or"},
			},
			1,
			nil,
		},
		{
			"invalid msm",
			urlpkg.Values{
				paramMSM: []string{"bad"},
			},
			0,
			errBadParam,
		},
		{
			"invalid operator",
			urlpkg.Values{
				paramOp: []string{"bad"},
			},
			0,
			errBadParam,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := FeedHandler{}
			got, err := h.msm(tc.params)
			if got != tc.want {
				t.Fatalf("got: %d; want: %d", got, tc.want)
			}
			if !errors.Is(err, tc.err) {
				t.Fatalf("got: %v; want: %v", err, tc.err)
			}
		})
	}
}

func testLookup(r *http.Request) *http.Response {
	server := httptest.NewServer(http.HandlerFunc(rdapHandler))
	defer server.Close()

	h := &FeedHandler{NewRDAPClient(server.URL)}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	return w.Result()
}
