package indeed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/sync/errgroup"
)

type RDAPDomain struct {
	Name   string      `json:"ldhName"`
	Events []RDAPEvent `json:"events"`
}

type RDAPEvent struct {
	Action string    `json:"eventAction"`
	Actor  string    `json:"eventActor"`
	Date   time.Time `json:"eventDate"`
}

type RDAPClient struct {
	BaseURL string
	client  *http.Client
}

func NewRDAPClient(baseURL string) *RDAPClient {
	return &RDAPClient{
		BaseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *RDAPClient) LookupDomains(ctx context.Context, names []string) ([]RDAPDomain, error) {
	g, ctx := errgroup.WithContext(ctx)
	ch := make(chan *RDAPDomain)

	for _, name := range names {
		name := name
		g.Go(func() error {
			domain, err := c.LookupDomain(ctx, name)
			if err != nil {
				return err
			}

			if domain != nil {
				select {
				case ch <- domain:
				case <-ctx.Done():
					return nil
				}
			}

			return nil
		})
	}
	go func() {
		g.Wait()
		close(ch)
	}()

	domains := make([]RDAPDomain, 0, len(names))
	for domain := range ch {
		domains = append(domains, *domain)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return domains, nil
}

func (c *RDAPClient) LookupDomain(ctx context.Context, name string) (*RDAPDomain, error) {
	url, err := url.JoinPath(c.BaseURL, "domain", name)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected response status %q", res.Status)
	}

	domain := new(RDAPDomain)
	if err := json.NewDecoder(res.Body).Decode(domain); err != nil {
		return nil, err
	}

	return domain, nil
}
