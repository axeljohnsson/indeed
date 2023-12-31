package indeed

import (
	"crypto/sha256"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	urlpkg "net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	paramQ   = "q"
	paramMSM = "msm"
	paramOp  = "op"
)

var (
	errBadParam = errors.New("bad value")
	errNoParam  = errors.New("no value")
)

var updateActionRE = regexp.MustCompile("last update of (RDAP|WHOIS) database")

type FeedHandler struct {
	r Resolver
}

func NewFeedHandler(rdap *RDAPClient, whois *WHOISClient) *FeedHandler {
	r := MultiResolver([]Resolver{
		rdap,
		TryResolver(whois, errNoServer),
	})
	return &FeedHandler{r}
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

	msm, err := h.msm(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	domains, err := ResolveDomains(r.Context(), h.r, names)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var match int
	for _, name := range names {
		for _, domain := range domains {
			if strings.ToLower(domain.Name) == name {
				match++
				break
			}
		}
	}
	if match < msm {
		http.Error(w, "domain(s) not found", http.StatusNotFound)
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

func (h *FeedHandler) names(params urlpkg.Values) ([]string, error) {
	if !params.Has(paramQ) {
		return nil, &paramError{
			name: paramQ,
			err:  errNoParam,
		}
	}

	names := params[paramQ]
	for i, name := range names {
		names[i] = strings.ToLower(name)
	}

	sort.Strings(names)

	return names, nil
}

func (h *FeedHandler) msm(params urlpkg.Values) (int, error) {
	if params.Has(paramMSM) {
		value := params.Get(paramMSM)
		msm, err := strconv.Atoi(value)
		if err != nil {
			return 0, &paramError{
				name:  paramMSM,
				value: value,
				err:   errBadParam,
			}
		}
		return msm, nil
	}

	if params.Has(paramOp) {
		switch value := params.Get(paramOp); value {
		case "and":
			return len(params[paramQ]), nil
		case "or":
			return 1, nil
		default:
			return 0, &paramError{
				name:  paramOp,
				value: value,
				err:   errBadParam,
			}
		}
	}

	return len(params[paramQ]), nil
}

func (h *FeedHandler) convert(names []string, domains []Domain) (*RSSFeed, error) {
	items := make([]RSSItem, 0)
	for _, domain := range domains {
		for _, event := range domain.Events {
			if updateActionRE.MatchString(event.Action) {
				continue
			}
			items = append(items, RSSItem{
				Link:        domain.Link,
				Description: fmt.Sprintf("%s: %s", strings.ToLower(domain.Name), event.Action),
				Author:      event.Actor,
				GUID:        h.itemGUID(&domain, &event),
				PubDate:     RSSTime{event.Date},
			})
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].PubDate.After(items[j].PubDate.Time)
	})

	var link urlpkg.URL
	link.Path = "/feed"
	link.RawQuery = urlpkg.Values{paramQ: names}.Encode()

	return &RSSFeed{
		Version:     "2.0",
		Title:       "Domain Events",
		Link:        link.String(),
		Description: fmt.Sprintf("Domain events for: %s.", strings.Join(names, ", ")),
		Items:       items,
	}, nil
}

func (h *FeedHandler) itemGUID(domain *Domain, event *Event) string {
	w := sha256.New()
	w.Write([]byte(strings.ToLower(domain.Name)))
	w.Write([]byte(event.Action))
	w.Write([]byte(event.Date.Format(time.RFC3339)))
	return fmt.Sprintf("%x", w.Sum(nil))
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

type paramError struct {
	name  string
	value string
	err   error
}

func (e *paramError) Error() string {
	msg := fmt.Sprintf("parameter %q: %v", e.name, e.err)
	if e.value != "" {
		msg = fmt.Sprintf("%s %q", msg, e.value)
	}
	return msg
}

func (e *paramError) Unwrap() error {
	return e.err
}
