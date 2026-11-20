package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	exitNotFound = 6
	exitWarning  = 5
)

func applyCmd(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	baseURL := fs.String("base-url", "", "")
	root := fs.String("root", "", "")
	if err := fs.Parse(args); err != nil || *baseURL == "" || fs.NArg() != 2 {
		usage()
		return 1
	}
	action, id := fs.Arg(0), fs.Arg(1)
	if !idRe.MatchString(id) {
		fmt.Fprintln(os.Stderr, "erro: id inválido:", id)
		return 1
	}
	if r, err := resolveRoot(*root); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	} else {
		*root = r
	}
	signal.Ignore(syscall.SIGHUP)
	s := &server{root: *root, baseURL: *baseURL}
	var err error
	switch action {
	case "archive":
		err = s.setField(id, func(m manifest) { m["archived"] = true })
	case "unarchive":
		err = s.setField(id, func(m manifest) { m["archived"] = false })
	case "title":
		raw, rerr := io.ReadAll(io.LimitReader(os.Stdin, maxBodyBytes+1))
		if rerr != nil || len(raw) > maxBodyBytes {
			fmt.Fprintln(os.Stderr, "erro: título ilegível ou grande demais")
			return 1
		}
		title, terr := cleanTitle(string(raw))
		if terr != nil {
			fmt.Fprintln(os.Stderr, "erro:", terr)
			return 1
		}
		err = s.setField(id, func(m manifest) { m["title"] = title; m["title_locked"] = true })
	case "delete":
		err = s.remove(id, true)
	default:
		usage()
		return 1
	}
	var ae apiError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &ae) && ae.status == http.StatusNotFound:
		fmt.Fprintln(os.Stderr, "erro:", ae.msg)
		return exitNotFound
	case errors.As(err, &ae) && ae.status == http.StatusOK:
		fmt.Fprintln(os.Stderr, "aviso:", ae.msg)
		return exitWarning
	default:
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 4
	}
}
