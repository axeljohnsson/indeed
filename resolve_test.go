package indeed

import (
	"context"
	"errors"
	"testing"
)

func TestMultiResolver(t *testing.T) {
	one := &Domain{
		Name: "EXAMPLE.COM",
	}
	two := &Domain{
		Name: "EXAMPLE.COM",
	}
	tests := []struct {
		description string
		name        string
		r           Resolver
		want        *Domain
	}{
		{
			"multiple",
			"example.com",
			MultiResolver([]Resolver{
				mapResolver{
					"example.com": one,
				},
				mapResolver{
					"example.com": two,
				},
			}),
			one,
		},
		{
			"fallback",
			"example.com",
			MultiResolver([]Resolver{
				mapResolver{},
				mapResolver{
					"example.com": two,
				},
			}),
			two,
		},
		{
			"not found",
			"404.com",
			MultiResolver([]Resolver{
				mapResolver{},
			}),
			nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got, err := tc.r.Resolve(context.Background(), tc.name)
			if err != nil {
				t.Fatal(err)
			}

			if got != tc.want {
				t.Fatalf("got: %v; want: %v", got, tc.want)
			}
		})
	}
}

func TestTryResolver(t *testing.T) {
	one := errors.New("one")
	two := errors.New("two")
	tests := []struct {
		description string
		name        string
		r           Resolver
		want        error
	}{
		{
			"ok",
			"example.com",
			TryResolver(mapResolver{"example.com": &Domain{Name: "EXAMPLE.COM"}}, one),
			nil,
		},
		{
			"error match",
			"example.com",
			TryResolver(&errResolver{one}, one),
			nil,
		},
		{
			"error miss",
			"example.com",
			TryResolver(&errResolver{two}, one),
			two,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			_, got := tc.r.Resolve(context.Background(), tc.name)
			if got != tc.want {
				t.Fatalf("got: %v; want: %v", got, tc.want)
			}
		})
	}
}

type mapResolver map[string]*Domain

func (r mapResolver) Resolve(ctx context.Context, name string) (*Domain, error) {
	domain, ok := r[name]
	if !ok {
		return nil, nil
	}
	return domain, nil
}

type errResolver struct {
	err error
}

func (e *errResolver) Resolve(ctx context.Context, name string) (*Domain, error) {
	return nil, e.err
}
