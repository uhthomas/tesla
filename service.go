package tesla

import "net/http"

type Service struct {
	c *http.Client
}

func New(token string, opts ...Option) (*Service, error) {
	var s Service
	for _, opt := range opts {
		if err := opt(&s); err != nil {
			return nil, err
		}
	}
	return &s, nil
}
