package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	clockLeeway   = 60 * time.Second
	keysMaxAge    = time.Hour
	keysMinRetry  = time.Minute
	keysColdRetry = 10 * time.Second
	keysHardLimit = 24 * time.Hour
	maxTokenBytes = 16 << 10
)

type accessVerifier struct {
	issuer   string
	aud      string
	certsURL string
	client   *http.Client
	now      func() time.Time

	mu      sync.Mutex
	keys    map[string]*rsa.PublicKey
	fetched time.Time
	tried   time.Time
}

type accessClaims struct {
	Email string `json:"email"`
}

func newAccessVerifier(team, aud string) *accessVerifier {
	issuer := "https://" + team
	return &accessVerifier{
		issuer:   issuer,
		aud:      aud,
		certsURL: issuer + "/cdn-cgi/access/certs",
		client:   newHTTPClient(),
		now:      time.Now,
	}
}

func b64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

func (v *accessVerifier) verify(token string) (accessClaims, error) {
	var out accessClaims
	if token == "" {
		return out, errors.New("sem token")
	}
	if len(token) > maxTokenBytes {
		return out, errors.New("token grande demais")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return out, errors.New("token malformado")
	}
	hb, err := b64(parts[0])
	if err != nil {
		return out, errors.New("header inválido")
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(hb, &header); err != nil {
		return out, errors.New("header inválido")
	}
	if header.Alg != "RS256" {
		return out, fmt.Errorf("alg não aceito: %q", header.Alg)
	}
	sig, err := b64(parts[2])
	if err != nil {
		return out, errors.New("assinatura inválida")
	}
	key, err := v.key(header.Kid)
	if err != nil {
		return out, err
	}
	sum := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], sig); err != nil {
		return out, errors.New("assinatura não confere")
	}
	cb, err := b64(parts[1])
	if err != nil {
		return out, errors.New("claims inválidas")
	}
	var claims struct {
		Iss   string          `json:"iss"`
		Aud   json.RawMessage `json:"aud"`
		Exp   *float64        `json:"exp"`
		Nbf   *float64        `json:"nbf"`
		Email string          `json:"email"`
	}
	if err := json.Unmarshal(cb, &claims); err != nil {
		return out, errors.New("claims inválidas")
	}
	if claims.Iss != v.issuer {
		return out, errors.New("issuer não confere")
	}
	if !audContains(claims.Aud, v.aud) {
		return out, errors.New("aud não confere")
	}
	now := v.now()
	if claims.Exp == nil || now.After(time.Unix(int64(*claims.Exp), 0).Add(clockLeeway)) {
		return out, errors.New("token expirado")
	}
	if claims.Nbf != nil && now.Add(clockLeeway).Before(time.Unix(int64(*claims.Nbf), 0)) {
		return out, errors.New("token ainda não vale")
	}
	out.Email = claims.Email
	return out, nil
}

func audContains(raw json.RawMessage, want string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return one == want
	}
	var many []string
	if json.Unmarshal(raw, &many) == nil {
		for _, a := range many {
			if a == want {
				return true
			}
		}
	}
	return false
}

func (v *accessVerifier) key(kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	now := v.now()
	if k, ok := v.keys[kid]; ok && now.Sub(v.fetched) < keysMaxAge {
		return k, nil
	}
	retry := keysMinRetry
	if v.keys == nil {
		retry = keysColdRetry
	}
	if v.tried.IsZero() || now.Sub(v.tried) >= retry {
		v.tried = now
		if keys, err := v.fetch(); err == nil {
			v.keys = keys
			v.fetched = now
		} else if v.keys == nil {
			return nil, fmt.Errorf("chaves do Access indisponíveis: %w", err)
		}
	}
	if v.keys == nil {
		return nil, errors.New("chaves do Access indisponíveis")
	}
	if now.Sub(v.fetched) > keysHardLimit {
		return nil, errors.New("chaves do Access vencidas e sem renovação")
	}
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	return nil, errors.New("kid desconhecido")
}

func (v *accessVerifier) fetch() (map[string]*rsa.PublicKey, error) {
	resp, err := v.client.Get(v.certsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("certs: %s", resp.Status)
	}
	var doc struct {
		Keys []struct {
			Kty string `json:"kty"`
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&doc); err != nil {
		return nil, err
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		if k.Kty != "RSA" || k.Kid == "" {
			continue
		}
		nb, err1 := b64(k.N)
		eb, err2 := b64(k.E)
		if err1 != nil || err2 != nil || len(eb) == 0 || len(eb) > 4 {
			continue
		}
		e := 0
		for _, c := range eb {
			e = e<<8 | int(c)
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}
	}
	if len(keys) == 0 {
		return nil, errors.New("nenhuma chave RSA")
	}
	return keys, nil
}
