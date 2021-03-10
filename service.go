package tesla

import (
	"context"
	"net/http"
	"net/url"
)

type Service struct {
	c       *http.Client
	baseURL *url.URL
}

// New creates a new Tesla service client.
func New(ctx context.Context, opts ...Option) (*Service, error) {
	s := &Service{
		baseURL: &url.URL{
			Scheme: "https",
			Host:   "owner-api.teslamotors.com",
			Path:   "api/1",
		},
	}
	for _, opt := range opts {
		if err := opt(ctx, s); err != nil {
			return nil, err
		}
	}
	s.c.Transport = &Transport{RoundTripper: s.c.Transport}
	return s, nil
}
