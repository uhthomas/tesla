package tesla

import (
	"context"
	"net/http"
	"net/url"
)

type Service struct {
	c         *http.Client
	baseURL   *url.URL
	userAgent string
}

// New creates a new Tesla service client.
func New(ctx context.Context, opts ...Option) (*Service, error) {
	s := &Service{
		baseURL: &url.URL{
			Scheme: "https",
			Host:   "owner-api.teslamotors.com",
			Path:   "api/1",
		},
		userAgent: "uhthomas/tesla",
	}
	for _, opt := range opts {
		if err := opt(ctx, s); err != nil {
			return nil, err
		}
	}
	return s, nil
}
