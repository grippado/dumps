package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

var fallbackNameservers = []string{"1.1.1.1", "8.8.8.8"}

func nameservers(paths ...string) []string {
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		var out []string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			fields := strings.Fields(sc.Text())
			if len(fields) >= 2 && fields[0] == "nameserver" && net.ParseIP(fields[1]) != nil {
				out = append(out, fields[1])
			}
		}
		f.Close()
		if len(out) > 0 {
			return out
		}
	}
	return fallbackNameservers
}

func resolvConfPaths() []string {
	var paths []string
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		paths = append(paths, filepath.Join(prefix, "etc", "resolv.conf"))
	}
	return append(paths, "/data/data/com.termux/files/usr/etc/resolv.conf", "/etc/resolv.conf")
}

func termuxRoots(paths ...string) *x509.CertPool {
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM(data) {
			return pool
		}
	}
	return nil
}

func certPaths() []string {
	var paths []string
	if prefix := os.Getenv("PREFIX"); prefix != "" {
		paths = append(paths, filepath.Join(prefix, "etc", "tls", "cert.pem"))
	}
	return append(paths, "/data/data/com.termux/files/usr/etc/tls/cert.pem")
}

func newHTTPClient() *http.Client {
	servers := nameservers(resolvConfPaths()...)
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	var next atomic.Uint32
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			ns := servers[int(next.Add(1)-1)%len(servers)]
			return dialer.DialContext(ctx, network, net.JoinHostPort(ns, "53"))
		},
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{Timeout: 5 * time.Second, Resolver: resolver}).DialContext
	if roots := termuxRoots(certPaths()...); roots != nil {
		transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	}
	return &http.Client{Timeout: 10 * time.Second, Transport: transport}
}
