package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	st, err := in.Stat()
	if err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, st.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func activate(root, baseURL string, ga bool, id string) error {
	if !idRe.MatchString(id) {
		return fmt.Errorf("id inválido: %s", id)
	}
	unlock, err := lockRoot(root)
	if err != nil {
		return err
	}
	defer unlock()

	j := func(parts ...string) string { return filepath.Join(append([]string{root}, parts...)...) }
	st := j("dumps-staging", id)
	content := j("dumps", id)
	meta := j("dumps-meta", id+".json")
	metaTmp := j("dumps-meta", "."+id+".json.tmp")
	bar := j("dumps", "_", "dumps-bar.js")
	thumb := j("dumps", "_", "thumbs", id+".jpg")
	barTmp, thumbTmp := st+".bar.tmp", st+".thumb.tmp"
	oldDir, oldMeta, oldBar, oldThumb := st+".old", st+".json.old", st+".bar.old", st+".thumb.old"

	if !exists(st) || !exists(metaTmp) {
		return errors.New("staging incompleto: falta o conteúdo ou o manifest")
	}
	for _, p := range []string{oldDir, oldMeta, oldBar, oldThumb} {
		os.RemoveAll(p)
	}
	if err := os.MkdirAll(filepath.Dir(thumb), 0o755); err != nil {
		return err
	}
	backup := func() error {
		if exists(thumb) {
			if err := copyFile(thumb, oldThumb); err != nil {
				return err
			}
		}
		if exists(bar) {
			if err := copyFile(bar, oldBar); err != nil {
				return err
			}
		}
		if exists(meta) {
			if err := copyFile(meta, oldMeta); err != nil {
				return err
			}
		}
		if exists(content) {
			if err := os.Rename(content, oldDir); err != nil {
				return err
			}
		}
		return nil
	}
	cleanupBackups := func() {
		for _, p := range []string{oldDir, oldMeta, oldBar, oldThumb} {
			os.RemoveAll(p)
		}
	}
	if err := backup(); err != nil {
		if exists(oldDir) && !exists(content) {
			os.Rename(oldDir, content)
		}
		cleanupBackups()
		for _, p := range []string{st, barTmp, thumbTmp, metaTmp} {
			os.RemoveAll(p)
		}
		return fmt.Errorf("backup: %w", err)
	}
	swap := func() error {
		if err := os.Rename(st, content); err != nil {
			return err
		}
		if err := os.Rename(metaTmp, meta); err != nil {
			return err
		}
		if exists(barTmp) {
			if err := os.Rename(barTmp, bar); err != nil {
				return err
			}
		}
		if exists(thumbTmp) {
			if err := os.Rename(thumbTmp, thumb); err != nil {
				return err
			}
		}
		return renderLocked(root, baseURL, ga)
	}
	failure := swap()
	if failure == nil {
		cleanupBackups()
		return nil
	}
	if exists(oldBar) {
		os.Rename(oldBar, bar)
	}
	if exists(oldThumb) {
		os.Rename(oldThumb, thumb)
	} else {
		os.Remove(thumb)
	}
	os.RemoveAll(content)
	if exists(oldDir) {
		os.Rename(oldDir, content)
	}
	if exists(oldMeta) {
		os.Rename(oldMeta, meta)
	} else {
		os.Remove(meta)
	}
	for _, p := range []string{st, barTmp, thumbTmp, metaTmp} {
		os.RemoveAll(p)
	}
	renderLocked(root, baseURL, ga)
	return fmt.Errorf("ativação falhou, versão anterior restaurada: %w", failure)
}

func activateCmd(args []string) int {
	fs := flag.NewFlagSet("activate", flag.ContinueOnError)
	baseURL := fs.String("base-url", "", "")
	ga := fs.Bool("ga", false, "")
	root := fs.String("root", "", "")
	if err := fs.Parse(args); err != nil || *baseURL == "" || fs.NArg() != 1 {
		usage()
		return 1
	}
	if r, err := resolveRoot(*root); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	} else {
		*root = r
	}
	signal.Ignore(syscall.SIGHUP)
	if err := activate(*root, *baseURL, *ga, fs.Arg(0)); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 4
	}
	return 0
}
