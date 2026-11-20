package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testTeam   = "team.example"
	testAud    = "aud-123"
	testOrigin = "https://admin.example.com"
)

type keypair struct {
	key *rsa.PrivateKey
	kid string
}

func newKeypair(t *testing.T, kid string) keypair {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	return keypair{k, kid}
}

func enc(v any) string {
	b, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(b)
}

func (kp keypair) sign(t *testing.T, header, claims map[string]any) string {
	t.Helper()
	signing := enc(header) + "." + enc(claims)
	sum := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, kp.key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func (kp keypair) token(t *testing.T, mut func(h, c map[string]any)) string {
	h := map[string]any{"alg": "RS256", "kid": kp.kid}
	c := map[string]any{"iss": "https://" + testTeam, "aud": []string{testAud}, "exp": time.Now().Add(time.Hour).Unix(), "nbf": time.Now().Add(-time.Minute).Unix(), "email": "g@example.com"}
	if mut != nil {
		mut(h, c)
	}
	return kp.sign(t, h, c)
}

func jwks(kps ...keypair) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var keys []map[string]string
		for _, kp := range kps {
			keys = append(keys, map[string]string{
				"kty": "RSA", "kid": kp.kid, "alg": "RS256",
				"n": base64.RawURLEncoding.EncodeToString(kp.key.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(kp.key.E)).Bytes()),
			})
		}
		json.NewEncoder(w).Encode(map[string]any{"keys": keys})
	})
}

func testVerifier(t *testing.T, kps ...keypair) *accessVerifier {
	ts := httptest.NewServer(jwks(kps...))
	t.Cleanup(ts.Close)
	v := newAccessVerifier(testTeam, testAud)
	v.certsURL = ts.URL
	return v
}

func TestVerifier(t *testing.T) {
	good := newKeypair(t, "k1")
	other := newKeypair(t, "k2")
	v := testVerifier(t, good)
	cases := map[string]struct {
		token string
		ok    bool
	}{
		"válido":         {good.token(t, nil), true},
		"aud string":     {good.token(t, func(h, c map[string]any) { c["aud"] = testAud }), true},
		"sem token":      {"", false},
		"lixo":           {"a.b.c", false},
		"outra chave":    {other.sign(t, map[string]any{"alg": "RS256", "kid": "k1"}, map[string]any{"iss": "https://" + testTeam, "aud": testAud, "exp": time.Now().Add(time.Hour).Unix()}), false},
		"kid estranho":   {other.token(t, nil), false},
		"alg none":       {good.token(t, func(h, c map[string]any) { h["alg"] = "none" }), false},
		"alg HS256":      {good.token(t, func(h, c map[string]any) { h["alg"] = "HS256" }), false},
		"aud errado":     {good.token(t, func(h, c map[string]any) { c["aud"] = []string{"outra-app"} }), false},
		"issuer errado":  {good.token(t, func(h, c map[string]any) { c["iss"] = "https://evil.example" }), false},
		"expirado":       {good.token(t, func(h, c map[string]any) { c["exp"] = time.Now().Add(-5 * time.Minute).Unix() }), false},
		"sem exp":        {good.token(t, func(h, c map[string]any) { delete(c, "exp") }), false},
		"nbf no futuro":  {good.token(t, func(h, c map[string]any) { c["nbf"] = time.Now().Add(10 * time.Minute).Unix() }), false},
		"claims trocada": {tamper(good.token(t, nil)), false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := v.verify(tc.token)
			if (err == nil) != tc.ok {
				t.Fatalf("ok=%v err=%v", tc.ok, err)
			}
		})
	}
}

func tamper(tok string) string {
	p := strings.Split(tok, ".")
	p[1] = enc(map[string]any{"iss": "https://" + testTeam, "aud": testAud, "exp": time.Now().Add(time.Hour).Unix(), "email": "intruso@example.com"})
	return strings.Join(p, ".")
}

func TestVerifierKeyRotation(t *testing.T) {
	k1 := newKeypair(t, "k1")
	k2 := newKeypair(t, "k2")
	current := []keypair{k1}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { jwks(current...).ServeHTTP(w, r) }))
	defer ts.Close()
	now := time.Now()
	v := newAccessVerifier(testTeam, testAud)
	v.certsURL = ts.URL
	v.now = func() time.Time { return now }
	if _, err := v.verify(k1.token(t, nil)); err != nil {
		t.Fatal(err)
	}
	current = []keypair{k1, k2}
	if _, err := v.verify(k2.token(t, nil)); err == nil {
		t.Fatal("aceitou kid novo antes do intervalo de nova busca")
	}
	now = now.Add(2 * time.Minute)
	if _, err := v.verify(k2.token(t, nil)); err != nil {
		t.Fatalf("não buscou a chave nova: %v", err)
	}
}

func testServer(t *testing.T) (*server, keypair, string) {
	kp := newKeypair(t, "k1")
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "dumps-meta"), 0o755)
	os.MkdirAll(filepath.Join(root, "dumps", "abc"), 0o755)
	os.MkdirAll(filepath.Join(root, "dumps", "_", "thumbs"), 0o755)
	os.WriteFile(filepath.Join(root, "dumps", "abc", "index.html"), []byte("oi"), 0o644)
	os.WriteFile(filepath.Join(root, "dumps", "_", "thumbs", "abc.jpg"), []byte("jpg"), 0o644)
	src, _ := os.ReadFile("testdata/one/meta/abc.json")
	os.WriteFile(filepath.Join(root, "dumps-meta", "abc.json"), src, 0o644)
	s := &server{root: root, baseURL: testBase, origin: testOrigin, auth: accessAuth{testVerifier(t, kp)}, logf: t.Logf}
	return s, kp, root
}

func do(s *server, method, path, token, origin, ctype, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		r.Header.Set("Cf-Access-Jwt-Assertion", token)
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if ctype != "" {
		r.Header.Set("Content-Type", ctype)
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	return w
}

func TestServerGuards(t *testing.T) {
	s, kp, _ := testServer(t)
	tok := kp.token(t, nil)
	body := `{"id":"abc"}`
	cases := []struct {
		name                         string
		method, path, tok, org, ctyp string
		body                         string
		want                         int
	}{
		{"sem jwt", "POST", "/_dumps/api/archive", "", testOrigin, "application/json", body, 403},
		{"jwt inválido", "POST", "/_dumps/api/archive", "x.y.z", testOrigin, "application/json", body, 403},
		{"get", "GET", "/_dumps/api/archive", tok, testOrigin, "", "", 405},
		{"sem origin", "POST", "/_dumps/api/archive", tok, "", "application/json", body, 403},
		{"origin errado", "POST", "/_dumps/api/archive", tok, "https://evil.example", "application/json", body, 403},
		{"form", "POST", "/_dumps/api/archive", tok, testOrigin, "application/x-www-form-urlencoded", body, 415},
		{"id inválido", "POST", "/_dumps/api/archive", tok, testOrigin, "application/json", `{"id":"../x"}`, 400},
		{"id underscore", "POST", "/_dumps/api/delete", tok, testOrigin, "application/json", `{"id":"_"}`, 400},
		{"campo extra", "POST", "/_dumps/api/archive", tok, testOrigin, "application/json", `{"id":"abc","x":1}`, 400},
		{"ação desconhecida", "POST", "/_dumps/api/nuke", tok, testOrigin, "application/json", body, 404},
		{"fora da api", "POST", "/outra", tok, testOrigin, "application/json", body, 404},
		{"inexistente", "POST", "/_dumps/api/archive", tok, testOrigin, "application/json", `{"id":"zzz"}`, 404},
		{"título vazio", "POST", "/_dumps/api/title", tok, testOrigin, "application/json", `{"id":"abc","title":"   "}`, 400},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if w := do(s, c.method, c.path, c.tok, c.org, c.ctyp, c.body); w.Code != c.want {
				t.Fatalf("status %d, quer %d: %s", w.Code, c.want, w.Body.String())
			}
		})
	}
	if w := do(s, "GET", "/healthz", "", "", "", ""); w.Code != 200 {
		t.Fatalf("healthz %d", w.Code)
	}
}

func TestServerActions(t *testing.T) {
	s, kp, root := testServer(t)
	tok := kp.token(t, nil)
	post := func(action, body string) {
		t.Helper()
		if w := do(s, "POST", "/_dumps/api/"+action, tok, testOrigin, "application/json; charset=utf-8", body); w.Code != 200 {
			t.Fatalf("%s: %d %s", action, w.Code, w.Body.String())
		}
	}
	read := func(rel string) string {
		b, _ := os.ReadFile(filepath.Join(root, rel))
		return string(b)
	}
	post("archive", `{"id":"abc"}`)
	if !strings.Contains(read("dumps-meta/abc.json"), `"archived": true`) || strings.Contains(read("dumps/index.html"), `href="/abc/"`) {
		t.Fatal("archive não tirou da home")
	}
	if !strings.Contains(read("admin/index.html"), `data-act="unarchive"`) {
		t.Fatal("listagem sem desarquivar")
	}
	post("unarchive", `{"id":"abc"}`)
	if !strings.Contains(read("dumps/index.html"), `href="/abc/"`) {
		t.Fatal("unarchive não voltou para a home")
	}
	post("title", `{"id":"abc","title":"  Novo   <título> & cia "}`)
	m := read("dumps-meta/abc.json")
	if !strings.Contains(m, `"title": "Novo <título> & cia"`) || !strings.Contains(m, `"title_locked": true`) || !strings.Contains(m, `"bytes": 2048`) {
		t.Fatalf("manifest depois do título: %s", m)
	}
	if !strings.Contains(read("dumps/index.html"), "Novo &lt;título&gt; &amp; cia") {
		t.Fatal("home sem o título novo")
	}
	post("delete", `{"id":"abc"}`)
	for _, rel := range []string{"dumps/abc", "dumps-meta/abc.json", "dumps/_/thumbs/abc.jpg"} {
		if _, err := os.Stat(filepath.Join(root, rel)); !os.IsNotExist(err) {
			t.Fatalf("%s sobrou", rel)
		}
	}
	if strings.Contains(read("admin/index.html"), "abc") {
		t.Fatal("listagem ainda mostra o apagado")
	}
}

func TestTitleRollsBackWhenRenderFails(t *testing.T) {
	s, kp, root := testServer(t)
	os.WriteFile(filepath.Join(root, "dumps-meta", "zz.json"), []byte("{quebrado"), 0o644)
	before, _ := os.ReadFile(filepath.Join(root, "dumps-meta", "abc.json"))
	w := do(s, "POST", "/_dumps/api/title", kp.token(t, nil), testOrigin, "application/json", `{"id":"abc","title":"X"}`)
	if w.Code != 500 {
		t.Fatalf("status %d", w.Code)
	}
	after, _ := os.ReadFile(filepath.Join(root, "dumps-meta", "abc.json"))
	if string(before) != string(after) {
		t.Fatal("manifest não voltou")
	}
}

func TestNameservers(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "resolv.conf")
	os.WriteFile(p, []byte("# x\nnameserver 9.9.9.9\nnameserver lixo\nsearch lan\nnameserver 2606:4700:4700::1111\n"), 0o644)
	got := nameservers(filepath.Join(dir, "nao-existe"), p)
	if strings.Join(got, ",") != "9.9.9.9,2606:4700:4700::1111" {
		t.Fatalf("got %v", got)
	}
	if strings.Join(nameservers(filepath.Join(dir, "nao-existe")), ",") != "1.1.1.1,8.8.8.8" {
		t.Fatal("sem fallback")
	}
}

func TestTitleRejectsControlAndDeleteNeedsManifest(t *testing.T) {
	s, kp, root := testServer(t)
	tok := kp.token(t, nil)
	if w := do(s, "POST", "/_dumps/api/title", tok, testOrigin, "application/json", `{"id":"abc","title":"x\u202ey"}`); w.Code != 400 {
		t.Fatalf("bidi aceito: %d", w.Code)
	}
	os.MkdirAll(filepath.Join(root, "dumps", "orfao"), 0o755)
	if w := do(s, "POST", "/_dumps/api/delete", tok, testOrigin, "application/json", `{"id":"orfao"}`); w.Code != 404 {
		t.Fatalf("apagou pasta sem manifest pelo admin: %d", w.Code)
	}
	if err := s.remove("orfao", true); err != nil {
		t.Fatalf("CLI não removeu órfão: %v", err)
	}
}
