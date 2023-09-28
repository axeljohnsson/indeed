package indeed

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"
)

type Resolver interface {
	Resolve(ctx context.Context, name string) (*Domain, error)
}

type Domain struct {
	Name   string
	Link   string
	Events []Event
}

type Event struct {
	Action string
	Actor  string
	Date   time.Time
}

func ResolveDomains(ctx context.Context, resolver Resolver, names []string) ([]Domain, error) {
	g, ctx := errgroup.WithContext(ctx)
	ch := make(chan *Domain)

	for _, name := range names {
		name := name
		g.Go(func() error {
			domain, err := resolver.Resolve(ctx, name)
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

	domains := make([]Domain, 0, len(names))
	for domain := range ch {
		domains = append(domains, *domain)
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return domains, nil
}

type multiResolver struct {
	rr []Resolver
}

func MultiResolver(resolvers []Resolver) Resolver {
	return &multiResolver{rr: resolvers}
}

func (r *multiResolver) Resolve(ctx context.Context, name string) (*Domain, error) {
	g, ctx := errgroup.WithContext(ctx)
	domains := make([]*Domain, len(r.rr))

	for i := range r.rr {
		i := i
		g.Go(func() error {
			var err error
			domains[i], err = r.rr[i].Resolve(ctx, name)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domain != nil {
			return domain, nil
		}
	}

	return nil, nil
}

type tryResolver struct {
	r   Resolver
	err error
}

func TryResolver(resolver Resolver, err error) Resolver {
	return &tryResolver{
		r:   resolver,
		err: err,
	}
}

func (r *tryResolver) Resolve(ctx context.Context, name string) (*Domain, error) {
	domain, err := r.r.Resolve(ctx, name)
	if err != nil {
		if errors.Is(err, r.err) {
			return nil, nil
		}
		return nil, err
	}
	return domain, nil
}
