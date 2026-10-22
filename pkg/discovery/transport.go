package discovery

import (
	"net/http"
	"time"
)

// DefaultTimeout is the timeout of the HTTP clients the services use to call
// each other.
const DefaultTimeout = 5 * time.Second

// Transport returns an http.RoundTripper that sends every request to a healthy
// instance of service, chosen per request, instead of the host in the
// request's URL. The URL's scheme, path and query are kept, so a client keeps
// its configured base URL ("http://reward:8085") and discovery only decides
// where it really goes. When the registry has nothing for the service the
// request goes to the configured address, i.e. unchanged.
//
// In static mode it returns next untouched (http.DefaultTransport if nil), so
// nothing changes for services that don't use a registry.
func (l *Locator) Transport(service string, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	if !l.Enabled() {
		return next
	}
	return &transport{loc: l, service: service, next: next}
}

// HTTPClient returns the HTTP client for calling service: DefaultTimeout and,
// when a registry is in use, a Transport that follows its healthy instances.
func (l *Locator) HTTPClient(service string) *http.Client {
	return &http.Client{Timeout: DefaultTimeout, Transport: l.Transport(service, nil)}
}

type transport struct {
	loc     *Locator
	service string
	next    http.RoundTripper
}

// RoundTrip implements http.RoundTripper. It never modifies req.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	hostPort := t.loc.Target(t.service)
	if hostPort == "" || hostPort == req.URL.Host {
		return t.next.RoundTrip(req)
	}
	redirected := req.Clone(req.Context())
	redirected.URL.Host = hostPort
	redirected.Host = hostPort
	return t.next.RoundTrip(redirected)
}
