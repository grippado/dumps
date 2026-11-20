package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var version = "dev"

const lockWait = 60 * time.Second

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 1
	}
	switch args[0] {
	case "version":
		fmt.Println(version)
		return 0
	case "capabilities":
		fmt.Println("render serve apply prepare activate")
		return 0
	case "render":
		return renderCmd(args[1:])
	case "serve":
		return serveCmd(args[1:])
	case "apply":
		return applyCmd(args[1:])
	case "prepare":
		return prepareCmd(args[1:])
	case "activate":
		return activateCmd(args[1:])
	default:
		usage()
		return 1
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "uso: dumpsd render --base-url <url> [--ga] [--no-lock] [--root <dir>]")
	fmt.Fprintln(os.Stderr, "     dumpsd apply --base-url <url> [--root <dir>] <archive|unarchive|title|delete> <id>  (title lê o título do stdin)")
	fmt.Fprintln(os.Stderr, "     dumpsd serve --base-url <url> --team <host> --aud <tag> --origin <url> [--listen <addr>] [--root <dir>]")
	fmt.Fprintln(os.Stderr, "     dumpsd prepare --dir <pasta> [--manifest <arquivo>]  (especificação JSON pelo stdin)")
	fmt.Fprintln(os.Stderr, "     dumpsd activate --base-url <url> [--ga] [--root <dir>] <id>")
	fmt.Fprintln(os.Stderr, "     dumpsd version")
}

func renderCmd(args []string) int {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	baseURL := fs.String("base-url", "", "")
	ga := fs.Bool("ga", false, "")
	noLock := fs.Bool("no-lock", false, "")
	root := fs.String("root", "", "")
	if err := fs.Parse(args); err != nil || *baseURL == "" || fs.NArg() != 0 {
		usage()
		return 1
	}
	if r, err := resolveRoot(*root); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	} else {
		*root = r
	}
	run := render
	if *noLock {
		run = func(root, baseURL string, ga bool) error {
			signal.Ignore(syscall.SIGHUP)
			return renderLocked(root, baseURL, ga)
		}
	}
	if err := run(*root, *baseURL, *ga); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 4
	}
	return 0
}

func render(root, baseURL string, ga bool) error {
	signal.Ignore(syscall.SIGHUP)
	unlock, err := lockRoot(root)
	if err != nil {
		return err
	}
	defer unlock()
	return renderLocked(root, baseURL, ga)
}

func lockRoot(root string) (func(), error) {
	metaDir := filepath.Join(root, "dumps-meta")
	if err := os.MkdirAll(metaDir, 0o755); err != nil {
		return nil, err
	}
	return lock(filepath.Join(metaDir, ".render.lock"))
}

func renderLocked(root, baseURL string, ga bool) error {
	metaDir := filepath.Join(root, "dumps-meta")
	items, err := loadManifests(metaDir)
	if err != nil {
		return err
	}
	ver := ""
	bar := filepath.Join(root, "dumps", "_", "dumps-bar.js")
	if ga || barHasGA(bar) {
		ver = barVersion(bar)
	}
	site, err := loadSite(root, baseURL)
	if err != nil {
		return err
	}
	listing := renderListing(items, baseURL)
	home := renderHome(items, baseURL, ver, site)

	listingPath := filepath.Join(root, site.ListingDir, "index.html")
	homePath := filepath.Join(root, "dumps", "index.html")
	cleanStale(filepath.Dir(listingPath))
	cleanStale(filepath.Dir(homePath))
	lt, err := stage(listingPath, listing)
	if err != nil {
		return err
	}
	ht, err := stage(homePath, home)
	if err != nil {
		os.Remove(lt)
		return err
	}
	if err := os.Rename(lt, listingPath); err != nil {
		os.Remove(lt)
		os.Remove(ht)
		return err
	}
	if err := os.Rename(ht, homePath); err != nil {
		os.Remove(ht)
		return err
	}
	return nil
}

func lock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(lockWait)
	for {
		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, nil
		}
		if errors.Is(err, syscall.EINTR) {
			continue
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) || time.Now().After(deadline) {
			f.Close()
			return nil, fmt.Errorf("lock %s: %w", path, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func writeAtomic(dest string, content []byte) error {
	tmp, err := stage(dest, content)
	if err != nil {
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

func stage(dest string, content []byte) (string, error) {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, ".index.html.*.tmp")
	if err != nil {
		return "", err
	}
	name := f.Name()
	if _, err := f.Write(content); err != nil {
		f.Close()
		os.Remove(name)
		return "", err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(name)
		return "", err
	}
	if err := f.Close(); err != nil {
		os.Remove(name)
		return "", err
	}
	if err := os.Chmod(name, 0o644); err != nil {
		os.Remove(name)
		return "", err
	}
	return name, nil
}

func cleanStale(dir string) {
	matches, _ := filepath.Glob(filepath.Join(dir, ".index.html.*.tmp"))
	for _, m := range matches {
		if st, err := os.Stat(m); err == nil && time.Since(st.ModTime()) > time.Hour {
			os.Remove(m)
		}
	}
}
