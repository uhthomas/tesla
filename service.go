package tesla

import "net/http"

type Service struct {
	c *http.Client
}

// New creates a new Tesla service client. The provided HTTP client must be
// using an OAuth2 transport.
func New(c *http.Client, opts ...Option) (*Service, error) {
	s := &Service{c: c}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	return s, nil
}
