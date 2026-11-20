package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

//go:embed home.tmpl
var homeTmpl string

//go:embed icon.svg
var homeIcon string

//go:embed listing-head.html
var listingHead string

//go:embed listing-thead.html
var listingThead string

//go:embed listing-foot.html
var listingFoot string

type manifest map[string]any

var months = []string{"jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"}

func loadManifests(dir string) ([]manifest, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, ".") || !strings.HasSuffix(n, ".json") {
			continue
		}
		if st, err := os.Stat(filepath.Join(dir, n)); err != nil || !st.Mode().IsRegular() {
			continue
		}
		names = append(names, n)
	}
	sort.Strings(names)
	var out []manifest
	for _, n := range names {
		data, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			return nil, err
		}
		dec := json.NewDecoder(bytes.NewReader(data))
		dec.UseNumber()
		for {
			var m manifest
			err := dec.Decode(&m)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("manifest inválido %s: %w", n, err)
			}
			if m == nil {
				return nil, fmt.Errorf("manifest inválido %s: não é objeto", n)
			}
			out = append(out, m)
		}
	}
	return out, nil
}

func barVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])[:10]
}

func barHasGA(path string) bool {
	data, err := os.ReadFile(path)
	return err == nil && bytes.Contains(data, []byte("__dumpsGa"))
}

func str(m manifest, k string) (string, bool) {
	v, ok := m[k]
	if !ok || v == nil {
		return "", false
	}
	if s, ok := v.(string); ok {
		return s, true
	}
	b, _ := json.Marshal(v)
	return string(b), true
}

func truthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case json.Number:
		f, _ := x.Float64()
		return f != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	}
	return true
}

func jqTruthy(v any) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return true
}

var pyEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#x27;")
var jqEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&apos;")

func pyEsc(s string) string { return pyEscaper.Replace(s) }

func jqHTML(m manifest, k string) string {
	v, ok := m[k]
	if !ok || v == nil {
		return "null"
	}
	if s, ok := v.(string); ok {
		return jqEscaper.Replace(s)
	}
	b, _ := json.Marshal(v)
	return jqEscaper.Replace(string(b))
}

func prefixRunes(s string, n int) string {
	i := 0
	for pos := range s {
		if i == n {
			return s[:pos]
		}
		i++
	}
	return s
}

func pyInt(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	neg := false
	if s[0] == '+' || s[0] == '-' {
		neg = s[0] == '-'
		s = s[1:]
	}
	if s == "" || strings.HasPrefix(s, "_") || strings.HasSuffix(s, "_") || strings.Contains(s, "__") {
		return 0, false
	}
	s = strings.ReplaceAll(s, "_", "")
	n := 0
	for _, r := range s {
		if !unicode.IsDigit(r) || r > unicode.MaxASCII {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	if neg {
		n = -n
	}
	return n, true
}

func day(iso string) string {
	parts := strings.Split(prefixRunes(iso, 10), "-")
	if len(parts) != 3 {
		return ""
	}
	y := parts[0]
	mo, ok1 := pyInt(parts[1])
	d, ok2 := pyInt(parts[2])
	if !ok1 || !ok2 {
		return ""
	}
	idx := mo - 1
	if idx < 0 {
		idx += len(months)
	}
	if idx < 0 || idx >= len(months) {
		return ""
	}
	return strconv.Itoa(d) + " " + months[idx] + " " + y
}

func firstUpper(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if size == 0 {
		return ""
	}
	return strings.ToUpper(string(r))
}

func renderHome(all []manifest, baseURL, ver string, site siteConfig) []byte {
	var items []manifest
	for _, m := range all {
		if !truthy(m["archived"]) {
			items = append(items, m)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, _ := str(items[i], "updated")
		b, _ := str(items[j], "updated")
		return a > b
	})
	var cards strings.Builder
	for _, m := range items {
		id, _ := str(m, "id")
		title, _ := str(m, "title")
		if title == "" {
			title = id
		}
		updated, _ := str(m, "updated")
		thumb := ""
		if truthy(m["thumb"]) {
			thumb = fmt.Sprintf(`<img src="/_/thumbs/%s.jpg?v=%s" alt="" loading="lazy" decoding="async" onerror="this.remove()">`, pyEsc(id), pyEsc(prefixRunes(updated, 19)))
		}
		fmt.Fprintf(&cards, `<li class="card"><a href="%s"><div class="thumb"><span class="ph">%s</span>%s</div><div class="meta"><h2>%s</h2><time datetime="%s">%s</time></div></a></li>`,
			pyEsc("/"+id+"/"), pyEsc(firstUpper(title)), thumb, pyEsc(title), pyEsc(updated), pyEsc(day(updated)))
	}
	label := fmt.Sprintf("%d dumps", len(items))
	if len(items) == 1 {
		label = "1 dump"
	}
	featured := renderFeatured(all, site.Featured)
	empty := ""
	if len(items) == 0 {
		empty = `<p class="empty">Nada publicado por aqui ainda.</p>`
	}
	ga := ""
	if ver != "" {
		ga = fmt.Sprintf(`<script src="/_/dumps-bar.js?v=%s" defer></script>`, pyEsc(ver))
	}
	host := hostOf(baseURL)
	brand, suffix := brandParts(host)
	r := strings.NewReplacer(
		"%(ga)s", ga, "%(icon)s", homeIcon, "%(label)s", pyEsc(label), "%(empty)s", empty,
		"%(cards)s", cards.String(), "%(base)s", pyEsc(strings.Replace(baseURL, "https://", "", -1)), "%(baseurl)s", pyEsc(baseURL),
		"%(title)s", pyEsc(site.Title), "%(description)s", pyEsc(site.Description), "%(host)s", pyEsc(host),
		"%(brand)s", pyEsc(brand), "%(brand_suffix)s", pyEsc(suffix), "%(tagline)s", site.TaglineHTML,
		"%(host_card)s", site.HostCardHTML, "%(footer)s", site.FooterHTML, "%(featured)s", featured,
	)
	return []byte(r.Replace(homeTmpl) + "\n")
}

func renderFeatured(all []manifest, links []featuredLink) string {
	if len(links) == 0 {
		return ""
	}
	byID := map[string]manifest{}
	for _, m := range all {
		if id, ok := str(m, "id"); ok {
			byID[id] = m
		}
	}
	var b strings.Builder
	for _, f := range links {
		m, ok := byID[f.ID]
		if !ok {
			continue
		}
		title, _ := str(m, "title")
		if title == "" {
			title = f.ID
		}
		label := f.Label
		if label == "" {
			label = "Em destaque"
		}
		note := ""
		if f.Note != "" {
			note = "<span>" + pyEsc(f.Note) + "</span>"
		}
		fmt.Fprintf(&b, `<a href="%s"><span class="fk">%s</span><span class="ft"><b>%s</b>%s</span><span class="fa" aria-hidden="true">→</span></a>`,
			pyEsc("/"+f.ID+"/"), pyEsc(label), pyEsc(title), note)
	}
	if b.Len() == 0 {
		return ""
	}
	return `<nav class="featured" aria-label="Em destaque">` + b.String() + "</nav>\n"
}

func kb(v any) string {
	f := 0.0
	switch x := v.(type) {
	case json.Number:
		f, _ = x.Float64()
	}
	return strconv.FormatFloat(math.Floor(f/1024), 'f', -1, 64) + " KB"
}

func renderListing(all []manifest, baseURL string) []byte {
	items := append([]manifest(nil), all...)
	sort.SliceStable(items, func(i, j int) bool {
		a, _ := str(items[i], "updated")
		b, _ := str(items[j], "updated")
		return a < b
	})
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	arq := 0
	for _, m := range items {
		if jqTruthy(m["archived"]) {
			arq++
		}
	}
	var b strings.Builder
	b.WriteString(listingHead)
	b.WriteString("\n")
	fmt.Fprintf(&b, "<h1>dumps</h1><p>%d publicados em %s", len(items), jqEscaper.Replace(baseURL))
	if arq > 0 {
		fmt.Fprintf(&b, ", %d arquivados (fora da home)", arq)
	}
	b.WriteString("</p>\n")
	b.WriteString(listingThead)
	b.WriteString("\n")
	for _, m := range items {
		arch := jqTruthy(m["archived"])
		cls, badge := "", ""
		fromAgent := false
		if src, _ := str(m, "source"); src == "agent" {
			fromAgent = true
		}
		if arch {
			cls = ` class="arquivado"`
			badge = ` <span class="arq">arquivado</span>`
		}
		if fromAgent {
			badge += ` <span class="arq ag">do agente</span>`
		}
		updated, _ := str(m, "updated")
		upd := strings.Replace(prefixRunes(updated, 16), "T", " ", 1)
		machine := ""
		if jqTruthy(m["machine"]) {
			machine = jqHTML(m, "machine")
		}
		bytesV := m["bytes"]
		if !jqTruthy(bytesV) {
			bytesV = json.Number("0")
		}
		toggle, toggleLabel, toggleCls := "archive", "Arquivar", ""
		if arch {
			toggle, toggleLabel = "unarchive", "Desarquivar"
			if fromAgent {
				toggleLabel, toggleCls = "Publicar", ` class="pub"`
			}
		}
		title, _ := str(m, "title")
		fmt.Fprintf(&b, "<tr%s data-id=\"%s\" data-title=\"%s\"><td><a href=\"%s\">%s</a>%s</td><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td><td class=\"act\"><button type=\"button\"%s data-act=\"%s\">%s</button><button type=\"button\" data-act=\"title\">Renomear</button><button type=\"button\" class=\"del\" data-act=\"delete\">Apagar</button></td></tr>\n",
			cls, jqHTML(m, "id"), jqEscaper.Replace(title), jqHTML(m, "url"), jqHTML(m, "title"), badge, jqHTML(m, "id"), jqEscaper.Replace(upd), kb(bytesV), machine, toggleCls, toggle, toggleLabel)
	}
	b.WriteString(listingFoot)
	b.WriteString("\n")
	return []byte(b.String())
}
