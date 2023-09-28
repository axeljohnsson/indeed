package indeed

import (
	"context"
	"errors"
	"io"
	"net/textproto"
	urlpkg "net/url"
	"regexp"
	"strings"
	"time"
)

const whoisBaseURL = "https://www.whois.com/"

var (
	errNoServer = errors.New("no WHOIS server")
	tldRE       = regexp.MustCompile(`\.[a-z]+$`)
	updateRE    = regexp.MustCompile(`\S+Z\S*`)
)

type WHOISClient struct {
	m func(string) string
}

func NewWHOISClient() *WHOISClient {
	return &WHOISClient{
		m: func(name string) string {
			switch tldRE.FindString(name) {
			case ".io":
				return "whois.nic.io:43"
			default:
				return ""
			}
		},
	}
}

func (c *WHOISClient) Resolve(ctx context.Context, name string) (*Domain, error) {
	addr := c.m(name)
	if addr == "" {
		return nil, errNoServer
	}

	conn, err := textproto.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	id, err := conn.Cmd(name)
	if err != nil {
		return nil, err
	}

	conn.StartResponse(id)
	defer conn.EndResponse(id)

	return c.unmarshal(conn)
}

func (c *WHOISClient) unmarshal(conn *textproto.Conn) (*Domain, error) {
	domain := &Domain{
		Events: make([]Event, 0),
	}

	for {
		line, err := conn.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		before, after, found := c.cutTrim(line)
		if !found {
			break
		}

		switch before {
		case "Domain Name":
			domain.Name = after
			link, err := urlpkg.JoinPath(whoisBaseURL, "whois", after)
			if err != nil {
				return nil, err
			}
			domain.Link = link
		case "Creation Date":
			if err := c.unmarshalEvent(after, domain, "registration"); err != nil {
				return nil, err
			}
		case "Registry Expiry Date":
			if err := c.unmarshalEvent(after, domain, "expiration"); err != nil {
				return nil, err
			}
		case "Updated Date":
			if err := c.unmarshalEvent(after, domain, "last changed"); err != nil {
				return nil, err
			}
		}

		if strings.HasPrefix(before, ">>>") {
			data := updateRE.FindString(after)
			action := "last update of WHOIS database"
			if err := c.unmarshalEvent(data, domain, action); err != nil {
				return nil, err
			}
			break
		}
	}

	if len(domain.Events) == 0 {
		return nil, nil
	}

	return domain, nil
}

func (c *WHOISClient) unmarshalEvent(data string, domain *Domain, action string) error {
	t, err := time.Parse(time.RFC3339, data)
	if err != nil {
		return err
	}
	domain.Events = append(domain.Events, Event{
		Action: action,
		Date:   t,
	})
	return nil
}

func (c *WHOISClient) cutTrim(line string) (string, string, bool) {
	b, a, ok := strings.Cut(line, ":")
	b = strings.TrimSpace(b)
	a = strings.TrimSpace(a)
	return b, a, ok
}
