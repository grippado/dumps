package main

import (
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func hasHiddenSegment(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if strings.HasPrefix(seg, ".") {
			return true
		}
	}
	return false
}

func (s *server) serveStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		reply(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	if strings.ContainsAny(r.URL.Path, "\\\x00") {
		http.NotFound(w, r)
		return
	}
	p := path.Clean("/" + r.URL.Path)
	if r.URL.Path != "/" && strings.HasSuffix(r.URL.Path, "/") && p != "/" {
		p += "/"
	}
	if strings.HasPrefix(p, "/_dumps") {
		http.NotFound(w, r)
		return
	}
	if hasHiddenSegment(p) {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(p, "/index.html") {
		localRedirect(w, r, strings.TrimSuffix(p, "index.html"))
		return
	}
	if strings.HasSuffix(p, ".html") {
		localRedirect(w, r, strings.TrimSuffix(p, ".html"))
		return
	}
	docroot := filepath.Join(s.root, "dumps")
	full := filepath.Join(docroot, filepath.FromSlash(p))
	if !strings.HasPrefix(full, docroot) {
		http.NotFound(w, r)
		return
	}
	if strings.HasSuffix(p, "/") {
		s.sendFile(w, r, filepath.Join(full, "index.html"))
		return
	}
	if st, err := os.Stat(full); err == nil && !st.IsDir() {
		s.sendFile(w, r, full)
		return
	}
	if st, err := os.Lstat(full + ".html"); err == nil && st.Mode().IsRegular() {
		s.sendFile(w, r, full+".html")
		return
	}
	if st, err := os.Stat(filepath.Join(full, "index.html")); err == nil && !st.IsDir() {
		localRedirect(w, r, p+"/")
		return
	}
	http.NotFound(w, r)
}

func localRedirect(w http.ResponseWriter, r *http.Request, p string) {
	if !strings.HasPrefix(p, "/") || strings.HasPrefix(p, "//") {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, (&url.URL{Path: p}).String(), http.StatusPermanentRedirect)
}

func (s *server) serveListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		reply(w, http.StatusMethodNotAllowed, "use GET")
		return
	}
	if r.URL.Path != "/_dumps/" {
		http.Redirect(w, r, "/_dumps/", http.StatusFound)
		return
	}
	site, err := loadSite(s.root, s.baseURL)
	if err != nil {
		reply(w, http.StatusInternalServerError, "site.json inválido")
		return
	}
	s.sendFile(w, r, filepath.Join(s.root, site.ListingDir, "index.html"))
}

func (s *server) sendFile(w http.ResponseWriter, r *http.Request, name string) {
	st, err := os.Lstat(name)
	if err != nil || !st.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}
	f, err := os.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	http.ServeContent(w, r, filepath.Base(name), st.ModTime(), f)
}
