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
		return &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}, nil
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
		var auth *proxy.Auth
		if u.User != nil {
			auth = &proxy.Auth{
				User:     u.User.Username(),
				Password: "",
			}
			if password, ok := u.User.Password(); ok {
				auth.Password = password
			}
		}
		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		})
		if err != nil {
			return nil, err
		}
		transport.Proxy = nil
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			type dialResult struct {
				conn net.Conn
				err  error
			}
			ch := make(chan dialResult, 1)
			go func() {
				conn, err := dialer.Dial(network, addr)
				ch <- dialResult{conn: conn, err: err}
			}()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case r := <-ch:
				return r.conn, r.err
			}
		}
		return transport, nil
	default:
		return nil, errors.New("unsupported proxy scheme: " + strings.TrimSpace(u.Scheme))
	}
}
