package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

type authenticator interface {
	authenticate(r *http.Request) (string, error)
}

type accessAuth struct{ v tokenVerifier }

func (a accessAuth) authenticate(r *http.Request) (string, error) {
	c, err := a.v.verify(r.Header.Get("Cf-Access-Jwt-Assertion"))
	if err != nil {
		return "", err
	}
	return c.Email, nil
}

type tailscaleAuth struct{ allow map[string]bool }

func fromLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (a tailscaleAuth) authenticate(r *http.Request) (string, error) {
	if !fromLoopback(r) {
		return "", errors.New("tailscale: pedido fora do loopback")
	}
	login := strings.TrimSpace(r.Header.Get("Tailscale-User-Login"))
	if login == "" {
		return "", errors.New("tailscale: sem Tailscale-User-Login")
	}
	if len(a.allow) > 0 && !a.allow[strings.ToLower(login)] {
		return "", errors.New("tailscale: login fora da lista")
	}
	return login, nil
}

type noAuth struct{}

func forwarded(r *http.Request) bool {
	for _, h := range []string{"X-Forwarded-For", "Forwarded", "Cf-Connecting-Ip", "X-Real-Ip", "Cf-Access-Jwt-Assertion"} {
		if r.Header.Get(h) != "" {
			return true
		}
	}
	return false
}

func (noAuth) authenticate(r *http.Request) (string, error) {
	if !fromLoopback(r) || forwarded(r) {
		return "", errors.New("sem login só aceita pedido direto do loopback")
	}
	return "local", nil
}

type listFlag []string

func (l *listFlag) String() string     { return strings.Join(*l, ",") }
func (l *listFlag) Set(v string) error { *l = append(*l, v); return nil }
