package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func stageDump(t *testing.T, root, id, body string) {
	t.Helper()
	st := filepath.Join(root, "dumps-staging", id)
	os.MkdirAll(st, 0o755)
	os.WriteFile(filepath.Join(st, "index.html"), []byte(body), 0o644)
	os.WriteFile(st+".bar.tmp", []byte("bar-"+body), 0o644)
	os.WriteFile(st+".thumb.tmp", []byte("jpg-"+body), 0o644)
	os.MkdirAll(filepath.Join(root, "dumps-meta"), 0o755)
	os.WriteFile(filepath.Join(root, "dumps-meta", "."+id+".json.tmp"), []byte(`{"id":"`+id+`","title":"`+body+`","updated":"2026-11-19T20:00:00-0300","url":"https://x/`+id+`/"}`), 0o644)
}

func TestActivateAndRollback(t *testing.T) {
	root := t.TempDir()
	read := func(rel string) string { b, _ := os.ReadFile(filepath.Join(root, rel)); return string(b) }
	stageDump(t, root, "abc", "v1")
	if err := activate(root, testBase, false, "abc"); err != nil {
		t.Fatal(err)
	}
	if read("dumps/abc/index.html") != "v1" || read("dumps/_/dumps-bar.js") != "bar-v1" || read("dumps/_/thumbs/abc.jpg") != "jpg-v1" || !strings.Contains(read("dumps/index.html"), "v1") {
		t.Fatal("v1 não ativou")
	}
	stageDump(t, root, "abc", "v2")
	os.WriteFile(filepath.Join(root, "dumps-meta", "zz.json"), []byte("{quebrado"), 0o644)
	if err := activate(root, testBase, false, "abc"); err == nil {
		t.Fatal("ativou com manifest quebrado no diretório")
	}
	if read("dumps/abc/index.html") != "v1" || read("dumps/_/dumps-bar.js") != "bar-v1" || read("dumps/_/thumbs/abc.jpg") != "jpg-v1" || !strings.Contains(read("dumps-meta/abc.json"), "v1") {
		t.Fatal("rollback não restaurou v1")
	}
	left, _ := filepath.Glob(filepath.Join(root, "dumps-staging", "*"))
	if len(left) != 0 {
		t.Fatalf("sobrou staging: %v", left)
	}
	os.Remove(filepath.Join(root, "dumps-meta", "zz.json"))
	stageDump(t, root, "new", "n1")
	os.WriteFile(filepath.Join(root, "dumps-meta", "zz.json"), []byte("{quebrado"), 0o644)
	if err := activate(root, testBase, false, "new"); err == nil {
		t.Fatal("ativou dump novo com manifest quebrado")
	}
	if _, err := os.Stat(filepath.Join(root, "dumps", "new")); !os.IsNotExist(err) {
		t.Fatal("dump novo ficou no ar depois do rollback")
	}
	if _, err := os.Stat(filepath.Join(root, "dumps", "_", "thumbs", "new.jpg")); !os.IsNotExist(err) {
		t.Fatal("miniatura do dump novo sobrou")
	}
	if err := activate(root, testBase, false, "../x"); err == nil {
		t.Fatal("aceitou id inválido")
	}
}
