package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type goldenParams struct {
	Title, URL, Version, Authors, Bar, Footer, Image, Names string
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return os.WriteFile(target, b, 0o644)
	})
}

func TestInjectGolden(t *testing.T) {
	cases, _ := filepath.Glob("testdata/inject/*/params.json")
	if len(cases) == 0 {
		t.Fatal("sem casos")
	}
	for _, pf := range cases {
		dir := filepath.Dir(pf)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			raw, _ := os.ReadFile(pf)
			var gp goldenParams
			if err := json.Unmarshal(raw, &gp); err != nil {
				t.Fatal(err)
			}
			sp := prepareSpec{Title: gp.Title, URL: gp.URL, Version: gp.Version, Authors: gp.Authors,
				Bar: gp.Bar == "true", Footer: gp.Footer == "true", OGImage: gp.Image, OGAuthor: gp.Names}
			work := t.TempDir()
			copyTree(t, filepath.Join(dir, "in"), work)
			if err := injectChrome(work, sp); err != nil {
				t.Fatal(err)
			}
			if err := injectOG(work, sp); err != nil {
				t.Fatal(err)
			}
			want := filepath.Join(dir, "out")
			filepath.WalkDir(want, func(p string, d fs.DirEntry, err error) error {
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(want, p)
				w, _ := os.ReadFile(p)
				g, err := os.ReadFile(filepath.Join(work, rel))
				if err != nil {
					t.Errorf("%s: %v", rel, err)
					return nil
				}
				if !bytes.Equal(w, g) {
					i := 0
					for i < len(w) && i < len(g) && w[i] == g[i] {
						i++
					}
					lo := max(0, i-60)
					t.Errorf("%s difere no byte %d\nquer: %q\ntem:  %q", rel, i, w[lo:min(len(w), i+60)], g[lo:min(len(g), i+60)])
				}
				return nil
			})
		})
	}
}

func TestInjectReadOnlyAndBytes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "index.html")
	os.WriteFile(p, []byte("<html><head><title>x</title></head><body>y</body></html>"), 0o444)
	sp := prepareSpec{Title: "x", URL: "https://dumps.example.com/x/", Version: "v", Authors: "[]", Bar: true, Footer: true}
	if err := injectChrome(dir, sp); err != nil {
		t.Fatal(err)
	}
	if err := injectOG(dir, sp); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "<dumps-bar") || !strings.Contains(string(b), `og:url" content="https://dumps.example.com/x/"`) {
		t.Fatalf("não injetou: %s", b)
	}
	n, err := totalBytes(dir)
	if err != nil || n != int64(len(b)) {
		t.Fatalf("bytes %d %v, quer %d", n, err, len(b))
	}
}

func TestPrepareUpdatesManifestBytes(t *testing.T) {
	dir := t.TempDir()
	content := filepath.Join(dir, "c")
	os.MkdirAll(content, 0o755)
	os.WriteFile(filepath.Join(content, "index.html"), []byte("<html><head></head><body>x</body></html>"), 0o644)
	man := filepath.Join(dir, "m.json")
	os.WriteFile(man, []byte(`{"id":"x","bytes":1,"title":"<a> & b"}`), 0o644)
	if err := prepare(content, man, prepareSpec{Title: "x", URL: "https://dumps.example.com/x/", Version: "v", Authors: "[]", Bar: true}); err != nil {
		t.Fatal(err)
	}
	n, _ := totalBytes(content)
	m, _ := os.ReadFile(man)
	if !strings.Contains(string(m), `"bytes": `+strconv.FormatInt(n, 10)) || !strings.Contains(string(m), `"title": "<a> & b"`) {
		t.Fatalf("manifest: %s", m)
	}
}

func TestPyDecodeAndNoncharacters(t *testing.T) {
	cases := map[string]string{
		"a\xe9\xa7b":     "a\uFFFDb",
		"a\xe2\x82b":     "a�b",
		"a\xf0\x9f\x98b": "a�b",
		"a\xed\xa0\x80b": "a���b",
		"ok ção":         "ok ção",
	}
	for in, want := range cases {
		if got := pyDecode([]byte(in)); got != want {
			t.Errorf("pyDecode(%q) = %q, quer %q", in, got, want)
		}
	}
	if got := cleanText("&#xFDD0;x&#xFFFF;&#x1FFFE;y &#65;"); got != "xy A" {
		t.Errorf("cleanText = %q", got)
	}
}
