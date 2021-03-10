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
func New(ctx context.Context, opts ...Option) (*Service, error) {
	s := &Service{
		baseURL: &url.URL{
			Scheme: "https",
			Host:   "owner-api.teslamotors.com",
			Path:   "api/1",
		},
	}
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		c.Transport = &Transport{RoundTripper: c.Transport}
	}
	for _, opt := range opts {
		if err := opt(ctx, s); err != nil {
			return nil, err
		}
	}
	return s, nil
}
