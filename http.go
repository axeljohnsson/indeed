package indeed

import (
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const paramQ = "q"

type FeedHandler struct {
	rdap *RDAPClient
}

func NewFeedHandler(client *RDAPClient) *FeedHandler {
	return &FeedHandler{client}
}

func (h *FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	names, err := h.names(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	domains, err := h.rdap.LookupDomains(r.Context(), names)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	feed, err := h.convert(names, domains)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/xml")

	if xml.NewEncoder(w).Encode(feed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *FeedHandler) names(params url.Values) ([]string, error) {
	if !params.Has(paramQ) {
		return nil, fmt.Errorf("missing query parameter %q", paramQ)
	}

	names := params[paramQ]
	for i, name := range names {
		names[i] = strings.ToLower(name)
	}

	sort.Strings(names)

	return names, nil
}

func (h *FeedHandler) convert(names []string, domains []RDAPDomain) (*RSSFeed, error) {
	items := make([]RSSItem, 0)
	for _, domain := range domains {
		for _, event := range domain.Events {
			link, err := url.JoinPath(h.rdap.BaseURL, "domain", domain.Name)
			if err != nil {
				return nil, err
			}
			items = append(items, RSSItem{
				Link:        link,
				Description: fmt.Sprintf("%s: %s", strings.ToLower(domain.Name), event.Action),
				Author:      event.Actor,
				PubDate:     RSSTime{event.Date},
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PubDate.After(items[j].PubDate.Time)
	})

	var link url.URL
	link.Path = "/feed"
	link.RawQuery = url.Values{paramQ: names}.Encode()

	return &RSSFeed{
		Version:     "2.0",
		Title:       "Domain Events",
		Link:        link.String(),
		Description: fmt.Sprintf("Domain events for: %s.", strings.Join(names, ", ")),
		Items:       items,
	}, nil
}

func LogHandler(h http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Now().Sub(start)

		req := slog.Group("req", slog.String("method", r.Method), slog.String("url", r.URL.String()))
		logger.InfoContext(r.Context(), "processed", req, slog.Duration("duration", duration))
	})
}
