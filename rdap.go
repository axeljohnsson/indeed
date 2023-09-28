package indeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	urlpkg "net/url"
	"time"
)

const RDAPBaseURL = "https://rdap.org/"

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

func (c *RDAPClient) Resolve(ctx context.Context, name string) (*Domain, error) {
	url, err := urlpkg.JoinPath(c.BaseURL, "domain", name)
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

	return c.unmarshal(res.Body)
}

func (c *RDAPClient) unmarshal(r io.Reader) (*Domain, error) {
	var body struct {
		Name   string `json:"ldhName"`
		Events []struct {
			Action string    `json:"eventAction"`
			Actor  string    `json:"eventActor"`
			Date   time.Time `json:"eventDate"`
		} `json:"events"`
	}
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return nil, err
	}

	domain := Domain{
		Name:   body.Name,
		Events: make([]Event, len(body.Events)),
	}

	for i, event := range body.Events {
		domain.Events[i] = Event{
			Action: event.Action,
			Actor:  event.Actor,
			Date:   event.Date,
		}
	}

	link, err := urlpkg.JoinPath(RDAPBaseURL, "domain", domain.Name)
	if err != nil {
		return nil, err
	}
	domain.Link = link

	return &domain, nil
}
