package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/manifoldco/promptui"
)

type Transaction struct {
	c  *http.Client
	id string
}

func NewTransaction(c *http.Client) *Transaction {
	if c == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			panic(fmt.Errorf("new cookie jar: %w", err))
		}

		c = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).Dial,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Jar: jar,
		}
	}
	return &Transaction{c: c}
}

func (t *Transaction) Do(ctx context.Context, authURL string) (code string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "tesla_exporter")

	res, err := t.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	username, err := (&promptui.Prompt{
		Label:   "Username",
		Pointer: promptui.PipeCursor,
		Validate: func(s string) error {
			if len(s) == 0 {
				return errors.New("len(s) == 0")
			}
			return nil
		},
	}).Run()
	if err != nil {
		return "", err
	}

	password, err := (&promptui.Prompt{
		Label:   "Password",
		Mask:    ' ',
		Pointer: promptui.PipeCursor,
		Validate: func(s string) error {
			if len(s) == 0 {
				return errors.New("len(s) == 0")
			}
			return nil
		},
	}).Run()
	if err != nil {
		return "", err
	}

	v := url.Values{
		"identity":   {username},
		"credential": {password},
	}

	d, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", fmt.Errorf("new document: %w", err)
	}

	d.Find("input[type=hidden]").Each(func(_ int, s *goquery.Selection) {
		name, ok := s.Attr("name")
		if !ok {
			return
		}
		value, ok := s.Attr("value")
		if !ok {
			return
		}
		v.Set(name, value)
	})

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(v.Encode()))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "tesla_exporter")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = t.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusFound:
		return codeFromResponse(res)
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	t.id = v.Get("transaction_id")

	devices, err := t.listDevices(ctx)
	if err != nil {
		return "", fmt.Errorf("list devices: %w", err)
	}

	if len(devices) != 0 {
		if err := t.verify(ctx, devices); err != nil {
			return "", fmt.Errorf("verify: %w", err)
		}
	}

	cr := t.c.CheckRedirect
	t.c.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer func() { t.c.CheckRedirect = cr }()

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, authURL, strings.NewReader(url.Values{
		"transaction_id": {t.id},
	}.Encode()))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "tesla_exporter")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err = t.c.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusFound {
		return "", fmt.Errorf("unexpected status code %d", res.StatusCode)
	}
	return codeFromResponse(res)
}

type Device struct {
	DispatchRequired bool      `json:"dispatchRequired"`
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	FactorType       string    `json:"factorType"`
	FactorProvider   string    `json:"factorProvider"`
	SecurityLevel    int       `json:"securityLevel"`
	Activated        time.Time `json:"activatedAt"`
	Updated          time.Time `json:"updatedAt"`
}

func (d Device) String() string { return d.Name }

func (t *Transaction) listDevices(ctx context.Context) ([]Device, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, (&url.URL{
		Scheme:   "https",
		Host:     "auth.tesla.com",
		Path:     "oauth2/v3/authorize/mfa/factors",
		RawQuery: url.Values{"transaction_id": {t.id}}.Encode(),
	}).String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "tesla_exporter")

	res, err := t.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	var out struct {
		Data []Device `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return out.Data, nil
}

func (t *Transaction) verify(ctx context.Context, devices []Device) error {
	var i int
	if len(devices) > 1 {
		var err error
		i, _, err = (&promptui.Select{
			Label:   "Device",
			Items:   devices,
			Pointer: promptui.PipeCursor,
		}).Run()
		if err != nil {
			return fmt.Errorf("select device: %w", err)
		}
	}
	d := devices[i]

	passcode, err := (&promptui.Prompt{
		Label:   "Passcode",
		Pointer: promptui.PipeCursor,
		Validate: func(s string) error {
			if len(s) != 6 {
				return errors.New("len(s) != 6")
			}
			return nil
		},
	}).Run()
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]string{
		"transaction_id": t.id,
		"factor_id":      d.ID,
		"passcode":       passcode,
	}); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://auth.tesla.com/oauth2/v3/authorize/mfa/verify", &buf)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("User-Agent", "tesla_exporter")
	req.Header.Set("Content-Type", "application/json")

	res, err := t.c.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	var out struct {
		Data struct {
			Approved bool `json:"approved"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}

	if !out.Data.Approved {
		return errors.New("not approved")
	}
	return nil
}

func codeFromResponse(res *http.Response) (code string, err error) {
	u, err := res.Location()
	if err != nil {
		return "", fmt.Errorf("response location: %w", err)
	}
	return u.Query().Get("code"), nil
}
