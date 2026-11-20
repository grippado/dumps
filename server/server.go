package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	apiPrefix    = "/_dumps/api/"
	maxBodyBytes = 4 << 10
	maxTitle     = 200
)

var idRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

type tokenVerifier interface {
	verify(token string) (accessClaims, error)
}

type server struct {
	root    string
	baseURL string
	origin  string
	auth    authenticator
	role    string
	logf    func(format string, args ...any)
}

const (
	roleAPI    = ""
	rolePublic = "public"
	roleAdmin  = "admin"
)

type apiError struct {
	status int
	msg    string
}

func (e apiError) Error() string { return e.msg }

func serveCmd(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	listen := fs.String("listen", "127.0.0.1:8090", "")
	baseURL := fs.String("base-url", "", "")
	team := fs.String("team", "", "")
	aud := fs.String("aud", "", "")
	origin := fs.String("origin", "", "")
	root := fs.String("root", "", "")
	authMode := fs.String("auth", "access", "")
	static := fs.Bool("static", false, "")
	adminListen := fs.String("admin-listen", "", "")
	var allow listFlag
	fs.Var(&allow, "allow", "")
	if err := fs.Parse(args); err != nil || *baseURL == "" || *origin == "" || fs.NArg() != 0 {
		usage()
		return 1
	}
	if !strings.HasPrefix(*listen, "127.0.0.1:") && !strings.HasPrefix(*listen, "localhost:") && !strings.HasPrefix(*listen, "[::1]:") {
		fmt.Fprintln(os.Stderr, "erro: --listen precisa ser loopback (127.0.0.1, localhost ou [::1])")
		return 1
	}
	var auth authenticator
	switch *authMode {
	case "access":
		if *team == "" || *aud == "" {
			fmt.Fprintln(os.Stderr, "erro: --auth access exige --team e --aud")
			return 1
		}
		auth = accessAuth{newAccessVerifier(*team, *aud)}
	case "tailscale":
		set := map[string]bool{}
		for _, a := range allow {
			set[strings.ToLower(strings.TrimSpace(a))] = true
		}
		auth = tailscaleAuth{allow: set}
	case "none":
		auth = noAuth{}
	default:
		fmt.Fprintln(os.Stderr, "erro: --auth precisa ser access, tailscale ou none")
		return 1
	}
	if *static && *authMode != "none" {
		fmt.Fprintln(os.Stderr, "erro: --static só com --auth none (modo local); em produção, quem serve os arquivos é o proxy")
		return 1
	}
	if *static && (*adminListen == "" || *adminListen == *listen) {
		fmt.Fprintln(os.Stderr, "erro: --static exige --admin-listen numa porta diferente, para o admin ficar noutra origem que os dumps")
		return 1
	}
	if *adminListen != "" && !strings.HasPrefix(*adminListen, "127.0.0.1:") && !strings.HasPrefix(*adminListen, "localhost:") && !strings.HasPrefix(*adminListen, "[::1]:") {
		fmt.Fprintln(os.Stderr, "erro: --admin-listen precisa ser loopback")
		return 1
	}
	if r, err := resolveRoot(*root); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		return 1
	} else {
		*root = r
	}
	newSrv := func(addr string, h http.Handler) *http.Server {
		return &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 90 * time.Second, MaxHeaderBytes: 64 << 10}
	}
	var servers []*http.Server
	if *static {
		servers = append(servers,
			newSrv(*listen, &server{root: *root, baseURL: *baseURL, origin: *origin, auth: auth, role: rolePublic, logf: log.Printf}),
			newSrv(*adminListen, &server{root: *root, baseURL: *baseURL, origin: *origin, auth: auth, role: roleAdmin, logf: log.Printf}))
	} else {
		servers = append(servers, newSrv(*listen, &server{root: *root, baseURL: *baseURL, origin: *origin, auth: auth, role: roleAPI, logf: log.Printf}))
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer stop()
	errs := make(chan error, len(servers))
	for _, srv := range servers {
		log.Printf("dumpsd %s ouvindo em %s (auth %s)", version, srv.Addr, *authMode)
		go func(srv *http.Server) { errs <- srv.ListenAndServe() }(srv)
	}
	var failed error
	select {
	case <-ctx.Done():
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			failed = err
		}
	}
	shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for _, srv := range servers {
		srv.Shutdown(shut)
	}
	if failed != nil {
		fmt.Fprintln(os.Stderr, "erro:", failed)
		return 4
	}
	return 0
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if r.URL.Path == "/healthz" && r.Method == http.MethodGet {
		w.Write([]byte("ok\n"))
		return
	}
	switch s.role {
	case rolePublic:
		s.serveStatic(w, r)
		return
	case roleAdmin:
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/_dumps/", http.StatusFound)
			return
		}
		if !strings.HasPrefix(r.URL.Path, apiPrefix) {
			if strings.HasPrefix(r.URL.Path, "/_dumps") {
				s.serveListing(w, r)
				return
			}
			reply(w, http.StatusNotFound, "não encontrado")
			return
		}
	default:
		if !strings.HasPrefix(r.URL.Path, apiPrefix) {
			reply(w, http.StatusNotFound, "não encontrado")
			return
		}
	}
	user, err := s.auth.authenticate(r)
	if err != nil {
		s.logf("negado %s %s: %v", r.Method, r.URL.Path, err)
		reply(w, http.StatusForbidden, "acesso negado")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		reply(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	if r.Header.Get("Origin") != s.origin {
		s.logf("origin recusado %q para %s", r.Header.Get("Origin"), user)
		reply(w, http.StatusForbidden, "origin não aceito")
		return
	}
	ct := strings.TrimSpace(strings.SplitN(r.Header.Get("Content-Type"), ";", 2)[0])
	if !strings.EqualFold(ct, "application/json") {
		reply(w, http.StatusUnsupportedMediaType, "envie application/json")
		return
	}
	var req struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		reply(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	if !idRe.MatchString(req.ID) {
		reply(w, http.StatusBadRequest, "id inválido")
		return
	}
	action := strings.TrimPrefix(r.URL.Path, apiPrefix)
	switch action {
	case "archive":
		err = s.setField(req.ID, func(m manifest) { m["archived"] = true })
	case "unarchive":
		err = s.setField(req.ID, func(m manifest) { m["archived"] = false })
	case "title":
		title, terr := cleanTitle(req.Title)
		if terr != nil {
			reply(w, http.StatusBadRequest, terr.Error())
			return
		}
		err = s.setField(req.ID, func(m manifest) { m["title"] = title; m["title_locked"] = true })
	case "delete":
		err = s.remove(req.ID, false)
	default:
		reply(w, http.StatusNotFound, "ação desconhecida")
		return
	}
	if err != nil {
		var ae apiError
		if errors.As(err, &ae) && ae.status == http.StatusOK {
			s.logf("%s %s por %s, com aviso: %s", action, req.ID, user, ae.msg)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"ok":true}` + "\n"))
			return
		}
		if errors.As(err, &ae) {
			reply(w, ae.status, ae.msg)
			return
		}
		s.logf("falha %s %s por %s: %v", action, req.ID, user, err)
		reply(w, http.StatusInternalServerError, "falha no servidor; confira com dumps list")
		return
	}
	s.logf("%s %s por %s", action, req.ID, user)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}` + "\n"))
}

func cleanTitle(raw string) (string, error) {
	if !utf8.ValidString(raw) {
		return "", errors.New("título com UTF-8 inválido")
	}
	for _, r := range raw {
		if unicode.IsSpace(r) {
			continue
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r) {
			return "", errors.New("título com caractere de controle")
		}
	}
	title := strings.Join(strings.Fields(raw), " ")
	if title == "" || utf8.RuneCountInString(title) > maxTitle {
		return "", fmt.Errorf("título precisa ter de 1 a %d caracteres", maxTitle)
	}
	return title, nil
}

func reply(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	b, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(append(b, '\n'))
}

func (s *server) setField(id string, change func(manifest)) error {
	unlock, err := lockRoot(s.root)
	if err != nil {
		return err
	}
	defer unlock()
	path := filepath.Join(s.root, "dumps-meta", id+".json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return apiError{http.StatusNotFound, "dump não existe: " + id}
	}
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var m manifest
	if err := dec.Decode(&m); err != nil || m == nil {
		return fmt.Errorf("manifest inválido %s: %v", id, err)
	}
	change(m)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return err
	}
	if err := writeAtomic(path, buf.Bytes()); err != nil {
		return err
	}
	if err := renderLocked(s.root, s.baseURL, false); err != nil {
		if rerr := writeAtomic(path, data); rerr != nil {
			return fmt.Errorf("%v; e o manifest anterior não voltou: %v", err, rerr)
		}
		renderLocked(s.root, s.baseURL, false)
		return err
	}
	return nil
}

func (s *server) remove(id string, allowOrphan bool) error {
	unlock, err := lockRoot(s.root)
	if err != nil {
		return err
	}
	defer unlock()
	meta := filepath.Join(s.root, "dumps-meta", id+".json")
	content := filepath.Join(s.root, "dumps", id)
	_, errMeta := os.Stat(meta)
	_, errContent := os.Lstat(content)
	if errors.Is(errMeta, os.ErrNotExist) && (!allowOrphan || errors.Is(errContent, os.ErrNotExist)) {
		return apiError{http.StatusNotFound, "dump não existe: " + id}
	}
	if err := os.RemoveAll(content); err != nil {
		return err
	}
	if err := os.Remove(meta); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	os.Remove(filepath.Join(s.root, "dumps", "_", "thumbs", id+".jpg"))
	if err := renderLocked(s.root, s.baseURL, false); err != nil {
		return apiError{http.StatusOK, "apagado, mas home e listagem não foram regeneradas: " + err.Error()}
	}
	return nil
}
