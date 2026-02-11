package channel

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

func newHTTPClient(proxyAddress string, timeout time.Duration) (*http.Client, error) {
	transport, err := newHTTPTransport(proxyAddress)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}, nil
}

func newHTTPTransport(proxyAddress string) (*http.Transport, error) {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return &http.Transport{}, nil
	}
	transport := base.Clone()
	proxyAddress = strings.TrimSpace(proxyAddress)
	if proxyAddress == "" {
		transport.Proxy = nil
		return transport, nil
	}
	if !strings.Contains(proxyAddress, "://") {
		proxyAddress = "http://" + proxyAddress
	}
	u, err := url.Parse(proxyAddress)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		transport.Proxy = http.ProxyURL(u)
		return transport, nil
	case "socks5", "socks5h":
		normalized := *u
		normalized.Scheme = "socks5"
		dialer, err := proxy.FromURL(&normalized, proxy.Direct)
		if err != nil {
			return nil, err
		}
		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
		return transport, nil
	default:
		return nil, errors.New("unsupported proxy scheme: " + strings.TrimSpace(u.Scheme))
	}
}
