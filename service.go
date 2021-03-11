package tesla

import (
	"context"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

type Service struct {
	c       *http.Client
	baseURL *url.URL
}

// New creates a new Tesla service client.
func New(ctx context.Context, c *http.Client, opts ...Option) (*Service, context.Context, error) {
	s := &Service{
		baseURL: &url.URL{
			Scheme: "https",
			Host:   "owner-api.teslamotors.com",
			Path:   "api/1",
		},
	}

	c.Transport = &Transport{RoundTripper: c.Transport}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, c)

	for _, opt := range opts {
		if err := opt(ctx, s); err != nil {
			return nil, nil, err
		}
	}
	return s, ctx, nil
}
