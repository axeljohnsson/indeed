package indeed

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func whoisServer(ch chan<- net.Addr) error {
	l, err := net.Listen("tcp", ":")
	if err != nil {
		return err
	}
	defer l.Close()

	ch <- l.Addr()

	conn, err := l.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	r := textproto.NewReader(bufio.NewReader(conn))
	line, err := r.ReadLine()
	if err != nil {
		return err
	}

	var name string
	switch line {
	case "example.com":
		name = "testdata/whois-example-com.txt"
	case "example.net":
		name = "testdata/whois-example-net.txt"
	case "example.org":
		name = "testdata/whois-example-org.txt"
	default:
		_, err := fmt.Fprint(conn, "Domain not found.\r\n")
		return err
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(conn, f); err != nil {
		return err
	}

	return nil
}

func TestWHOIS(t *testing.T) {
	tests := []struct {
		description string
		name        string
		want        *Domain
	}{
		{
			"ok",
			"example.com",
			&Domain{
				Name: "EXAMPLE.COM",
				Link: "https://www.whois.com/whois/EXAMPLE.COM",
				Events: []Event{
					{
						Action: "last changed",
						Date:   time.Date(2023, 8, 14, 7, 1, 38, 0, time.UTC),
					},
					{
						Action: "registration",
						Date:   time.Date(1995, 8, 14, 4, 0, 0, 0, time.UTC),
					},
					{
						Action: "expiration",
						Date:   time.Date(2024, 8, 13, 4, 0, 0, 0, time.UTC),
					},
					{
						Action: "last update of WHOIS database",
						Date:   time.Date(2023, 9, 6, 11, 4, 43, 0, time.UTC),
					},
				},
			},
		},
		{
			"not found",
			"404.com",
			nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			g, ctx := errgroup.WithContext(context.Background())
			ch := make(chan net.Addr)

			g.Go(func() error {
				return whoisServer(ch)
			})

			g.Go(func() error {
				addr := <-ch

				c := &WHOISClient{
					m: func(string) string {
						return addr.String()
					},
				}
				got, err := c.Resolve(ctx, tc.name)
				if err != nil {
					return err
				}

				if !reflect.DeepEqual(got, tc.want) {
					return fmt.Errorf("got: %v; want: %v", got, tc.want)
				}

				return nil
			})

			if err := g.Wait(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
