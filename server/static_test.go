package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func localServer(t *testing.T) (*server, string) {
	pub, _, root := localServers(t)
	return pub, root
}

func localServers(t *testing.T) (*server, *server, string) {
	root := t.TempDir()
	for _, d := range []string{"dumps/abc/sub", "dumps-meta", "admin"} {
		os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	files := map[string]string{
		"dumps/index.html":      "home",
		"dumps/abc/index.html":  "abc",
		"dumps/abc/page.html":   "page",
		"dumps/abc/sub/x.css":   "css",
		"dumps/abc/.secret":     "nao",
		"dumps/abc/sub.html":    "subpage",
		"admin/index.html":      "listagem",
		"dumps-meta/abc.json":   `{"id":"abc"}`,
		"dumps-meta/.site.json": `{"listing_dir":"admin"}`,
	}
	for name, body := range files {
		os.WriteFile(filepath.Join(root, name), []byte(body), 0o644)
	}
	mk := func(role string) *server {
		return &server{root: root, baseURL: "http://localhost:8090", origin: "http://localhost:8091", auth: noAuth{}, role: role, logf: t.Logf}
	}
	return mk(rolePublic), mk(roleAdmin), root
}

func get(s *server, path, remote string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", path, nil)
	if remote != "" {
		r.RemoteAddr = remote
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestStaticCleanURLs(t *testing.T) {
	s, _ := localServer(t)
	cases := []struct {
		path, body, loc string
		code            int
	}{
		{"/", "home", "", 200},
		{"/abc/", "abc", "", 200},
		{"/abc", "", "/abc/", 308},
		{"/abc/page", "page", "", 200},
		{"/abc/page.html", "", "/abc/page", 308},
		{"/abc/index.html", "", "/abc/", 308},
		{"/abc/sub/x.css", "css", "", 200},
		{"/abc/.secret", "", "", 404},
		{"/../dumps-meta/abc.json", "", "", 404},
		{"/dumps-meta/abc.json", "", "", 404},
		{"/_dumps/", "", "", 404},
		{"/_dumps/api/archive", "", "", 404},
		{"/%5Cevil.com.html", "", "", 404},
		{"/%5C%5Cevil.com/index.html", "", "", 404},
		{"/nada", "", "", 404},
		{"/abc/sub", "subpage", "", 200},
	}
	for _, c := range cases {
		w := get(s, c.path, "127.0.0.1:5555")
		if w.Code != c.code || (c.body != "" && w.Body.String() != c.body) || (c.loc != "" && w.Header().Get("Location") != c.loc) {
			t.Errorf("%s: %d %q loc=%q", c.path, w.Code, w.Body.String(), w.Header().Get("Location"))
		}
	}
}

func TestAdminOrigin(t *testing.T) {
	_, adm, _ := localServers(t)
	if w := get(adm, "/_dumps/", "127.0.0.1:1"); w.Code != 200 || w.Body.String() != "listagem" {
		t.Fatalf("admin sem listagem: %d", w.Code)
	}
	if w := get(adm, "/", "127.0.0.1:1"); w.Code != 302 {
		t.Fatalf("admin / : %d", w.Code)
	}
	if w := get(adm, "/abc/", "127.0.0.1:1"); w.Code != 404 {
		t.Fatalf("admin serviu dump: %d", w.Code)
	}
}

func TestAuthModes(t *testing.T) {
	_, s, _ := localServers(t)
	post := func(remote string, hdr map[string]string) int {
		r := httptest.NewRequest("POST", "/_dumps/api/archive", strings.NewReader(`{"id":"abc"}`))
		r.RemoteAddr = remote
		r.Header.Set("Origin", "http://localhost:8091")
		r.Header.Set("Content-Type", "application/json")
		for k, v := range hdr {
			r.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w.Code
	}
	if c := post("10.0.0.5:1234", nil); c != http.StatusForbidden {
		t.Fatalf("none aceitou fora do loopback: %d", c)
	}
	if c := post("127.0.0.1:1234", nil); c != 200 {
		t.Fatalf("none recusou loopback: %d", c)
	}
	if c := post("127.0.0.1:1234", map[string]string{"X-Forwarded-For": "203.0.113.9"}); c != http.StatusForbidden {
		t.Fatalf("none aceitou pedido repassado por proxy: %d", c)
	}
	s.auth = tailscaleAuth{allow: map[string]bool{"ada@example.com": true}}
	if c := post("127.0.0.1:1234", nil); c != http.StatusForbidden {
		t.Fatalf("tailscale sem header: %d", c)
	}
	if c := post("127.0.0.1:1234", map[string]string{"Tailscale-User-Login": "eve@example.com"}); c != http.StatusForbidden {
		t.Fatalf("tailscale aceitou login fora da lista: %d", c)
	}
	if c := post("10.0.0.5:1234", map[string]string{"Tailscale-User-Login": "ada@example.com"}); c != http.StatusForbidden {
		t.Fatalf("tailscale aceitou fora do loopback: %d", c)
	}
	if c := post("127.0.0.1:1234", map[string]string{"Tailscale-User-Login": "Ada@Example.com"}); c != 200 {
		t.Fatalf("tailscale recusou login da lista: %d", c)
	}
}

func TestListingDirAndRoot(t *testing.T) {
	for _, bad := range []string{"dumps", "Dumps/x", "dumps-staging", "dumps-meta", "../x", "/abs", ".hidden", "bin"} {
		root := t.TempDir()
		os.MkdirAll(filepath.Join(root, "dumps-meta"), 0o755)
		os.WriteFile(sitePath(root), []byte(`{"listing_dir":"`+bad+`"}`), 0o644)
		if _, err := loadSite(root, "http://x"); err == nil {
			t.Errorf("listing_dir %q aceito", bad)
		}
	}
	dir := t.TempDir()
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	os.Chdir(dir)
	if _, err := resolveRoot(""); err == nil {
		t.Fatal("resolveRoot aceitou diretório sem dumps-meta")
	}
	os.Mkdir("dumps-meta", 0o755)
	if _, err := resolveRoot(""); err != nil {
		t.Fatal(err)
	}
}
