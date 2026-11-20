package main

import (
	"bytes"
	"html"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	pySpace = `[\t\n\x0B\f\r ]`
	attrPat = `(?:"[^"]*"|'[^']*'|[^'">])*`
)

var (
	maskedRe   = regexp.MustCompile(`(?is)<!--.*?-->|<script\b.*?</script` + pySpace + `*>|<style\b.*?</style` + pySpace + `*>|<template\b.*?</template` + pySpace + `*>|<textarea\b.*?</textarea` + pySpace + `*>`)
	bodyRe     = regexp.MustCompile(`(?i)<body\b` + attrPat + `>`)
	headEndRe  = regexp.MustCompile(`(?i)</head` + pySpace + `*>`)
	htmlRe     = regexp.MustCompile(`(?i)<html\b` + attrPat + `>`)
	doctypeRe  = regexp.MustCompile(`(?i)<!doctype\b[^>]*>`)
	bodyEndRe  = regexp.MustCompile(`(?i)</body` + pySpace + `*>`)
	htmlEndRe  = regexp.MustCompile(`(?i)</html` + pySpace + `*>`)
	ogHeadRe   = regexp.MustCompile(`(?i)<head(?:>|[\t\n\x0B\f\r /]` + attrPat + `>)`)
	ogHTMLRe   = regexp.MustCompile(`(?i)<html(?:>|[\t\n\x0B\f\r /]` + attrPat + `>)`)
	ogTitleRe  = regexp.MustCompile(`(?is)<title(?:>|[\t\n\x0B\f\r /][^>]*>)(.*?)</title` + pySpace + `*>`)
	ogMetaRe   = regexp.MustCompile(`(?i)<meta([\t\n\x0B\f\r /]` + attrPat + `)>`)
	ogPairRe   = regexp.MustCompile(`([^\t\n\x0B\f\r "'=<>/]+)` + pySpace + `*=` + pySpace + "*(\"[^\"]*\"|'[^']*'|[^\\t\\n\\x0B\\f\\r \"'=<>`]+)")
	bomUTF8    = []byte("\xef\xbb\xbf")
	utf16Marks = [][]byte{[]byte("\xff\xfe"), []byte("\xfe\xff")}
)

type prepareSpec struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Version  string `json:"version"`
	Authors  string `json:"authors"`
	Bar      bool   `json:"bar"`
	Footer   bool   `json:"footer"`
	OGImage  string `json:"og_image"`
	OGAuthor string `json:"og_authors"`
}

func mask(data []byte) []byte {
	out := append([]byte(nil), data...)
	for _, m := range maskedRe.FindAllIndex(data, -1) {
		for i := m[0]; i < m[1]; i++ {
			out[i] = ' '
		}
	}
	return out
}

func htmlFiles(root string, fn func(path string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Type()&fs.ModeSymlink != 0 || !d.Type().IsRegular() {
			return nil
		}
		lower := strings.ToLower(d.Name())
		if !strings.HasSuffix(lower, ".html") && !strings.HasSuffix(lower, ".htm") {
			return nil
		}
		return fn(path)
	})
}

func readHTML(path string) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	for _, m := range utf16Marks {
		if bytes.HasPrefix(data, m) {
			return nil, false, nil
		}
	}
	return data, true, nil
}

func writeHTML(path string, data []byte) error {
	st, err := os.Stat(path)
	if err != nil {
		return err
	}
	if st.Mode().Perm()&0o200 == 0 {
		if err := os.Chmod(path, st.Mode().Perm()|0o200); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, st.Mode().Perm()|0o200)
}

func firstEnd(view []byte, res ...*regexp.Regexp) (int, bool) {
	for _, re := range res {
		if loc := re.FindIndex(view); loc != nil {
			return loc[1], true
		}
	}
	return 0, false
}

func injectChrome(root string, sp prepareSpec) error {
	bar := []byte(`<dumps-bar dump-title="` + pyEsc(sp.Title) + `" dump-url="` + pyEsc(sp.URL) + `"></dumps-bar>`)
	foot := []byte(`<dumps-footer dump-url="` + pyEsc(sp.URL) + `" authors="` + pyEsc(sp.Authors) + `"></dumps-footer>`)
	script := []byte(`<script src="/_/dumps-bar.js?v=` + sp.Version + `" defer></script>`)
	return htmlFiles(root, func(path string) error {
		data, ok, err := readHTML(path)
		if err != nil || !ok {
			return err
		}
		view := mask(data)
		addBar := sp.Bar && !bytes.Contains(view, []byte("<dumps-bar"))
		addFoot := sp.Footer && !bytes.Contains(view, []byte("<dumps-footer"))
		if !addBar && !addFoot {
			return nil
		}
		start, found := firstEnd(view, bodyRe, headEndRe, htmlRe, doctypeRe)
		if !found {
			start = 0
			if bytes.HasPrefix(data, bomUTF8) {
				start = 3
			}
		}
		end := len(data)
		for _, re := range []*regexp.Regexp{bodyEndRe, htmlEndRe} {
			all := re.FindAllIndex(view, -1)
			if len(all) > 0 {
				end = all[len(all)-1][0]
				break
			}
		}
		end = max(end, start)
		var head, tail []byte
		if addBar {
			head = append(append(head, bar...), script...)
		}
		if addFoot {
			tail = append(tail, foot...)
			if !addBar {
				tail = append(tail, script...)
			}
		}
		var out []byte
		out = append(out, data[:start]...)
		out = append(out, head...)
		out = append(out, data[start:end]...)
		out = append(out, tail...)
		out = append(out, data[end:]...)
		return writeHTML(path, out)
	})
}

func pyDecode(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	var sb strings.Builder
	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		if r == utf8.RuneError && size <= 1 {
			sb.WriteRune(utf8.RuneError)
			b = b[maximalSubpart(b):]
			continue
		}
		sb.WriteRune(r)
		b = b[size:]
	}
	return sb.String()
}

func maximalSubpart(b []byte) int {
	lo, hi := byte(0x80), byte(0xBF)
	need := 0
	switch c := b[0]; {
	case c >= 0xC2 && c <= 0xDF:
		need = 1
	case c == 0xE0:
		need, lo = 2, 0xA0
	case c >= 0xE1 && c <= 0xEC, c == 0xEE, c == 0xEF:
		need = 2
	case c == 0xED:
		need, hi = 2, 0x9F
	case c == 0xF0:
		need, lo = 3, 0x90
	case c >= 0xF1 && c <= 0xF3:
		need = 3
	case c == 0xF4:
		need, hi = 3, 0x8F
	default:
		return 1
	}
	i := 1
	for i <= need && i < len(b) {
		x := b[i]
		if i == 1 && (x < lo || x > hi) || i > 1 && (x < 0x80 || x > 0xBF) {
			break
		}
		i++
	}
	return i
}

var numericRefRe = regexp.MustCompile(`&#(?:[xX]([0-9a-fA-F]+)|([0-9]+));?`)

func isNoncharacter(n uint64) bool {
	return n >= 0xFDD0 && n <= 0xFDEF || n <= 0x10FFFF && n&0xFFFE == 0xFFFE
}

func dropNoncharacterRefs(v string) string {
	return numericRefRe.ReplaceAllStringFunc(v, func(ref string) string {
		m := numericRefRe.FindStringSubmatch(ref)
		var n uint64
		var err error
		if m[1] != "" {
			n, err = strconv.ParseUint(m[1], 16, 64)
		} else {
			n, err = strconv.ParseUint(m[2], 10, 64)
		}
		if err == nil && isNoncharacter(n) {
			return ""
		}
		return ref
	})
}

func cleanText(v string) string {
	v = html.UnescapeString(dropNoncharacterRefs(v))
	v = strings.Map(func(r rune) rune {
		if unicode.In(r, unicode.Cc, unicode.Cf, unicode.Zl, unicode.Zp) {
			return -1
		}
		return r
	}, v)
	return strings.Join(strings.Fields(v), " ")
}

func pyQuote(s string) string {
	var sb strings.Builder
	for _, c := range []byte(s) {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '_', c == '.', c == '-', c == '~', c == '/':
			sb.WriteByte(c)
		default:
			sb.WriteString("%")
			sb.WriteString(strings.ToUpper(string("0123456789abcdef"[c>>4]) + string("0123456789abcdef"[c&15])))
		}
	}
	return sb.String()
}

func pageURL(base, rel string) string {
	rel = filepath.ToSlash(rel)
	if rel == "index.html" || strings.HasSuffix(rel, "/index.html") {
		rel = rel[:len(rel)-len("index.html")]
	} else if strings.HasSuffix(rel, ".html") {
		rel = rel[:len(rel)-len(".html")]
	}
	return base + pyQuote(rel)
}

type metaTag map[string]string

func metas(data, view []byte, end int) []metaTag {
	var out []metaTag
	for _, m := range ogMetaRe.FindAllSubmatchIndex(view[:end], -1) {
		raw := data[m[2]:m[3]]
		d := metaTag{}
		for _, p := range ogPairRe.FindAllSubmatch(raw, -1) {
			k := strings.ToLower(pyDecode(p[1]))
			v := p[2]
			if len(v) >= 2 && (v[0] == '"' || v[0] == '\'') {
				v = v[1 : len(v)-1]
			}
			if _, ok := d[k]; !ok {
				d[k] = pyDecode(v)
			}
		}
		out = append(out, d)
	}
	return out
}

func (d metaTag) key() string {
	if v := d["property"]; v != "" {
		return strings.ToLower(v)
	}
	return strings.ToLower(d["name"])
}

func injectOG(root string, sp prepareSpec) error {
	return htmlFiles(root, func(path string) error {
		data, ok, err := readHTML(path)
		if err != nil || !ok {
			return err
		}
		view := mask(data)
		end := len(view)
		if loc := headEndRe.FindIndex(view); loc != nil {
			end = loc[0]
		}
		found := metas(data, view, end)
		keys := map[string]bool{}
		for _, d := range found {
			keys[d.key()] = true
		}
		value := func(k string) string {
			for _, d := range found {
				if d.key() == k {
					return d["content"]
				}
			}
			return ""
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		pageTitle := sp.Title
		if filepath.ToSlash(rel) != "index.html" {
			if m := ogTitleRe.FindSubmatchIndex(view[:end]); m != nil {
				if t := cleanText(pyDecode(data[m[2]:m[3]])); t != "" {
					pageTitle = t
				}
			}
		}
		if own := cleanText(value("og:title")); own != "" {
			pageTitle = own
		}
		desc := cleanText(value("description"))
		if desc == "" {
			if sp.OGAuthor != "" {
				desc = "Publicado por " + sp.OGAuthor + " em " + hostOf(sp.URL)
			} else {
				desc = "Publicado em " + hostOf(sp.URL)
			}
		}
		type tag struct{ key, kind, val string }
		tags := []tag{{"og:type", "property", "website"}, {"og:site_name", "property", hostOf(sp.URL)},
			{"og:title", "property", pageTitle}, {"og:description", "property", desc},
			{"og:url", "property", pageURL(sp.URL, rel)}, {"og:locale", "property", "pt_BR"}}
		switch {
		case keys["og:image"]:
			tags = append(tags, tag{"twitter:card", "name", "summary_large_image"})
		case sp.OGImage != "":
			tags = append(tags, tag{"og:image", "property", sp.OGImage}, tag{"og:image:width", "property", "640"},
				tag{"og:image:height", "property", "400"}, tag{"og:image:alt", "property", pageTitle},
				tag{"twitter:card", "name", "summary_large_image"})
		default:
			tags = append(tags, tag{"twitter:card", "name", "summary"})
		}
		tags = append(tags, tag{"twitter:title", "name", pageTitle}, tag{"twitter:description", "name", desc})
		var out []byte
		for _, t := range tags {
			if !keys[t.key] {
				out = append(out, `<meta `+t.kind+`="`+t.key+`" content="`+pyEsc(t.val)+`">`...)
			}
		}
		if len(out) == 0 {
			return nil
		}
		at, ok2 := firstEnd(view, ogHeadRe)
		if !ok2 {
			at, ok2 = firstEnd(view, ogHTMLRe, doctypeRe)
		}
		if !ok2 {
			at = 0
			if bytes.HasPrefix(data, bomUTF8) {
				at = 3
			}
		}
		res := append(append(append([]byte(nil), data[:at]...), out...), data[at:]...)
		return writeHTML(path, res)
	})
}

func totalBytes(root string) (int64, error) {
	var total int64
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			total += info.Size()
		}
		return nil
	})
	return total, err
}
