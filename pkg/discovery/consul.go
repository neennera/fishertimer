package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Consul talks to a Consul agent over its HTTP API. It implements Source and
// Registrar with the standard library only, so services don't pull in the
// Consul client and its dependencies.
type Consul struct {
	base   string
	client *http.Client
}

// NewConsul returns a client for the agent at addr, either a URL
// ("http://consul:8500") or "host:port".
func NewConsul(addr string) *Consul {
	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	return &Consul{base: addr, client: &http.Client{Timeout: 3 * time.Second}}
}

type healthEntry struct {
	Node struct {
		Address string `json:"Address"`
	} `json:"Node"`
	Service struct {
		ID      string `json:"ID"`
		Service string `json:"Service"`
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	} `json:"Service"`
}

// Resolve returns the instances of service whose health checks pass.
func (c *Consul) Resolve(ctx context.Context, service string) ([]Instance, error) {
	endpoint := c.base + "/v1/health/service/" + url.PathEscape(service) + "?passing=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var entries []healthEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("consul: decode %s: %w", service, err)
	}
	out := make([]Instance, 0, len(entries))
	for _, e := range entries {
		host := e.Service.Address
		if host == "" {
			host = e.Node.Address
		}
		out = append(out, Instance{ID: e.Service.ID, Name: e.Service.Service, Host: host, Port: e.Service.Port})
	}
	return out, nil
}

type registerBody struct {
	ID      string       `json:"ID"`
	Name    string       `json:"Name"`
	Address string       `json:"Address"`
	Port    int          `json:"Port"`
	Check   registerChck `json:"Check"`
}

type registerChck struct {
	HTTP                           string `json:"HTTP,omitempty"`
	TCP                            string `json:"TCP,omitempty"`
	Interval                       string `json:"Interval"`
	Timeout                        string `json:"Timeout"`
	DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter"`
}

// Register adds (or refreshes) an instance. It is idempotent: registering the
// same ID again replaces the entry.
func (c *Consul) Register(ctx context.Context, r Registration) error {
	chk := r.Check.withDefaults()
	payload, err := json.Marshal(registerBody{
		ID:      r.ID,
		Name:    r.Name,
		Address: r.Host,
		Port:    r.Port,
		Check: registerChck{
			HTTP:                           chk.HTTP,
			TCP:                            chk.TCP,
			Interval:                       chk.Interval.String(),
			Timeout:                        chk.Timeout.String(),
			DeregisterCriticalServiceAfter: chk.DeregisterAfter.String(),
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.base+"/v1/agent/service/register", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.do(req)
	return err
}

// Deregister removes an instance.
func (c *Consul) Deregister(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.base+"/v1/agent/service/deregister/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	_, err = c.do(req)
	return err
}

func (c *Consul) do(req *http.Request) ([]byte, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("consul: %s %s: %s: %s", req.Method, req.URL.Path, resp.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}
