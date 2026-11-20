package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func prepareCmd(args []string) int {
	fs := flag.NewFlagSet("prepare", flag.ContinueOnError)
	dir := fs.String("dir", "", "")
	manifestPath := fs.String("manifest", "", "")
	if err := fs.Parse(args); err != nil || *dir == "" || fs.NArg() != 0 {
		usage()
		return 1
	}
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 64<<10))
	if err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	}
	var sp prepareSpec
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&sp); err != nil {
		fmt.Fprintln(os.Stderr, "erro: especificação inválida:", err)
		return 1
	}
	var present map[string]json.RawMessage
	json.Unmarshal(raw, &present)
	for _, k := range []string{"title", "url", "version", "authors", "bar", "footer", "og_image", "og_authors"} {
		if _, ok := present[k]; !ok {
			fmt.Fprintln(os.Stderr, "erro: especificação sem o campo", k)
			return 1
		}
	}
	if st, err := os.Stat(*dir); err != nil || !st.IsDir() {
		fmt.Fprintln(os.Stderr, "erro: pasta inexistente:", *dir)
		return 1
	}
	if err := prepare(*dir, *manifestPath, sp); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 4
	}
	return 0
}

func prepare(dir, manifestPath string, sp prepareSpec) error {
	if err := injectChrome(dir, sp); err != nil {
		return fmt.Errorf("barra e footer: %w", err)
	}
	if err := injectOG(dir, sp); err != nil {
		return fmt.Errorf("open graph: %w", err)
	}
	if manifestPath == "" {
		return nil
	}
	total, err := totalBytes(dir)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var m manifest
	if err := d.Decode(&m); err != nil || m == nil {
		return fmt.Errorf("manifest inválido %s: %v", filepath.Base(manifestPath), err)
	}
	m["bytes"] = total
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return err
	}
	return writeAtomic(manifestPath, buf.Bytes())
}
