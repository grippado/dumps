package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const defaultDescription = "Artifacts published with dumps."

type featuredLink struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Note  string `json:"note"`
}

type siteConfig struct {
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	TaglineHTML  string         `json:"tagline_html"`
	HostCardHTML string         `json:"host_card_html"`
	FooterHTML   string         `json:"footer_html"`
	Featured     []featuredLink `json:"featured"`
	ListingDir   string         `json:"listing_dir"`
}

func sitePath(root string) string {
	return filepath.Join(root, "dumps-meta", ".site.json")
}

func hostOf(baseURL string) string {
	if u, err := url.Parse(baseURL); err == nil && u.Host != "" {
		return u.Host
	}
	return strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
}

func loadSite(root, baseURL string) (siteConfig, error) {
	var sc siteConfig
	data, err := os.ReadFile(sitePath(root))
	switch {
	case errors.Is(err, os.ErrNotExist):
	case err != nil:
		return sc, err
	default:
		dec := json.NewDecoder(strings.NewReader(string(data)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&sc); err != nil {
			return sc, fmt.Errorf("site.json inválido: %w", err)
		}
	}
	if sc.Title == "" {
		sc.Title = hostOf(baseURL)
	}
	if sc.Description == "" {
		sc.Description = defaultDescription
	}
	if sc.TaglineHTML == "" {
		sc.TaglineHTML = pyEsc(sc.Description)
	}
	if sc.ListingDir == "" {
		sc.ListingDir = "admin"
	}
	clean := filepath.Clean(sc.ListingDir)
	first := strings.ToLower(strings.SplitN(filepath.ToSlash(clean), "/", 2)[0])
	if filepath.IsAbs(clean) || clean == "." || strings.HasPrefix(clean, "..") || first == "dumps" || first == "dumps-meta" || first == "dumps-staging" || first == "bin" || strings.HasPrefix(first, ".") {
		return sc, fmt.Errorf("listing_dir inválido: %q", sc.ListingDir)
	}
	sc.ListingDir = clean
	for _, f := range sc.Featured {
		if !idRe.MatchString(f.ID) {
			return sc, fmt.Errorf("featured com id inválido: %q", f.ID)
		}
	}
	return sc, nil
}

func brandParts(host string) (string, string) {
	if i := strings.Index(host, "."); i > 0 {
		return host[:i], host[i:]
	}
	return host, ""
}
