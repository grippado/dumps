package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

const testBase = "https://dumps.example.com"

func TestGolden(t *testing.T) {
	cases, err := filepath.Glob("testdata/*/meta")
	if err != nil || len(cases) == 0 {
		t.Fatalf("sem casos: %v", err)
	}
	for _, metaDir := range cases {
		dir := filepath.Dir(metaDir)
		t.Run(filepath.Base(dir), func(t *testing.T) {
			items, err := loadManifests(metaDir)
			if err != nil {
				t.Fatal(err)
			}
			if os.Getenv("DUMPSD_UPDATE_GOLDEN") == "1" {
				os.WriteFile(filepath.Join(dir, "listing.html"), renderListing(items, testBase), 0o644)
			}
			check(t, filepath.Join(dir, "listing.html"), renderListing(items, testBase))
			site := testSite(t)
			if os.Getenv("DUMPSD_UPDATE_GOLDEN") == "1" {
				os.WriteFile(filepath.Join(dir, "home.html"), renderHome(items, testBase, "", site), 0o644)
				os.WriteFile(filepath.Join(dir, "home-ga.html"), renderHome(items, testBase, "abcdef0123", site), 0o644)
			}
			check(t, filepath.Join(dir, "home.html"), renderHome(items, testBase, "", site))
			check(t, filepath.Join(dir, "home-ga.html"), renderHome(items, testBase, "abcdef0123", site))
		})
	}
}

func testSite(t *testing.T) siteConfig {
	t.Helper()
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "dumps-meta"), 0o755)
	src := os.Getenv("DUMPSD_TEST_SITE")
	if src == "" {
		src = "testdata/site.json"
	}
	if b, err := os.ReadFile(src); err == nil {
		os.WriteFile(sitePath(root), b, 0o644)
	}
	sc, err := loadSite(root, testBase)
	if err != nil {
		t.Fatal(err)
	}
	return sc
}

func check(t *testing.T, golden string, got []byte) {
	t.Helper()
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		i := 0
		for i < len(want) && i < len(got) && want[i] == got[i] {
			i++
		}
		lo := max(0, i-80)
		t.Errorf("%s difere no byte %d\nquer: %q\ntem:  %q", golden, i, want[lo:min(len(want), i+80)], got[lo:min(len(got), i+80)])
	}
}

func TestRenderWritesAtomicallyAndCountsAll(t *testing.T) {
	root := t.TempDir()
	meta := filepath.Join(root, "dumps-meta")
	os.MkdirAll(meta, 0o755)
	src, _ := os.ReadFile("testdata/one/meta/abc.json")
	os.WriteFile(filepath.Join(meta, "abc.json"), src, 0o644)
	os.WriteFile(filepath.Join(meta, ".abc.json.tmp"), []byte("{quebrado"), 0o644)
	os.MkdirAll(filepath.Join(root, "dumps", "_"), 0o755)
	os.WriteFile(filepath.Join(root, "dumps", "_", "dumps-bar.js"), []byte("x"), 0o644)
	if err := render(root, testBase, true); err != nil {
		t.Fatal(err)
	}
	home, _ := os.ReadFile(filepath.Join(root, "dumps", "index.html"))
	listing, _ := os.ReadFile(filepath.Join(root, "admin", "index.html"))
	if !bytes.Contains(home, []byte(`/_/dumps-bar.js?v=2d711642b7`)) {
		t.Error("home sem a versão da barra servida")
	}
	if !bytes.Contains(listing, []byte("1 publicados")) {
		t.Error("listagem sem o dump")
	}
	st, _ := os.Stat(filepath.Join(root, "dumps", "index.html"))
	if st.Mode().Perm() != 0o644 {
		t.Errorf("permissão %v", st.Mode().Perm())
	}
	left, _ := filepath.Glob(filepath.Join(root, "dumps", ".index.html.*.tmp"))
	if len(left) != 0 {
		t.Errorf("sobrou temporário: %v", left)
	}
}

func TestInvalidManifestKeepsPages(t *testing.T) {
	root := t.TempDir()
	meta := filepath.Join(root, "dumps-meta")
	os.MkdirAll(filepath.Join(root, "dumps"), 0o755)
	os.MkdirAll(meta, 0o755)
	os.WriteFile(filepath.Join(root, "dumps", "index.html"), []byte("antiga"), 0o644)
	os.WriteFile(filepath.Join(meta, "x.json"), []byte("{quebrado"), 0o644)
	if err := render(root, testBase, false); err == nil {
		t.Fatal("render aceitou manifest inválido")
	}
	home, _ := os.ReadFile(filepath.Join(root, "dumps", "index.html"))
	if string(home) != "antiga" {
		t.Error("home trocada apesar do erro")
	}
}

func TestConcurrentRendersSerialize(t *testing.T) {
	root := t.TempDir()
	meta := filepath.Join(root, "dumps-meta")
	os.MkdirAll(meta, 0o755)
	src, _ := os.ReadFile("testdata/one/meta/abc.json")
	os.WriteFile(filepath.Join(meta, "abc.json"), src, 0o644)
	errs := make(chan error, 8)
	for range 8 {
		go func() { errs <- render(root, testBase, false) }()
	}
	for range 8 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	left, _ := filepath.Glob(filepath.Join(root, "*", ".index.html.*.tmp"))
	if len(left) != 0 {
		t.Errorf("sobrou temporário: %v", left)
	}
}
