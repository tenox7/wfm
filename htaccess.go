package main

// Apache style .htaccess files: Options -Indexes and denying access to
// folders and files. Applied on -webserver prefixes (-htaccess_web, on by
// default) and on -prefix and home prefixes to users without rw access
// (-htaccess, off by default). rw users are not bound: they can rewrite any
// .htaccess file, and need to see everything to manage the tree.

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"unicode"

	"github.com/spf13/afero"
)

// htMax caps the size of a .htaccess file; a larger one denies access.
const htMax = 1 << 20

// htTri is a setting a .htaccess file may leave unset, so the parent's applies.
type htTri int8

const (
	htUnset htTri = iota
	htYes         // access granted, or +Indexes
	htNo          // access denied, or -Indexes
)

// set records v. Within one file a denying value wins over a granting one:
// rules inside containers WFM can not evaluate, such as <IfModule>, all
// count, so none of them may open up what another one closes.
func (t *htTri) set(v htTri) {
	if v == htNo || *t == htUnset {
		*t = v
	}
}

// htConf is one parsed .htaccess file.
type htConf struct {
	access  htTri // Require, Allow and Deny outside <Files>
	indexes htTri // Options
	files   []*htFiles
	broken  bool // unreadable or unparsable: deny everything below
}

// htFiles is a <Files> or <FilesMatch> section.
type htFiles struct {
	match  func(name string) bool
	access htTri
}

// htSection is a container open while parsing.
type htSection struct {
	name  string
	files *htFiles // <Files> and <FilesMatch>
	skip  bool     // <Limit> or <LimitExcept> that does not cover GET
	deny  bool     // <RequireNone>: what it holds can only deny
}

// parseHtaccess reads the supported subset of a .htaccess file: Options
// (Indexes only), Require all granted|denied, Allow from all, Deny from all,
// and <Files>, <Files ~> and <FilesMatch> sections. An access rule it can
// not evaluate, such as Require ip, denies access and is returned as a note.
// Other directives are ignored and other containers read through, except
// <Limit> and <LimitExcept>, which count only when they cover GET.
func parseHtaccess(src string) (*htConf, []string, error) {
	c := &htConf{}
	var notes []string
	var open []htSection
	lines := strings.Split(strings.TrimPrefix(src, "\uFEFF"), "\n")
	for i := 0; i < len(lines); i++ {
		n := i + 1
		l := strings.TrimSpace(lines[i])
		for strings.HasSuffix(l, `\`) && i+1 < len(lines) {
			i++
			l = strings.TrimSuffix(l, `\`) + " " + strings.TrimSpace(lines[i])
		}
		if l == "" || l[0] == '#' {
			continue
		}
		if name, ok := strings.CutPrefix(l, "</"); ok {
			name = strings.TrimSpace(strings.TrimSuffix(name, ">"))
			if len(open) == 0 || !strings.EqualFold(open[len(open)-1].name, name) {
				return nil, notes, fmt.Errorf("line %d: unexpected </%v>", n, name)
			}
			open = open[:len(open)-1]
			continue
		}
		var files *htFiles
		skip, deny := false, false
		for _, s := range open {
			skip, deny = skip || s.skip, deny || s.deny
			if s.files != nil {
				files = s.files
			}
		}
		if l[0] == '<' {
			s, err := htOpen(l, files != nil)
			if err != nil {
				return nil, notes, fmt.Errorf("line %d: %v", n, err)
			}
			if s.files != nil {
				c.files = append(c.files, s.files)
			}
			open = append(open, s)
			continue
		}
		if skip {
			continue
		}
		f := htFields(l)
		switch d := strings.ToLower(f[0]); d {
		case "options":
			if files == nil {
				c.indexes.set(htOptions(f[1:]))
			}
		case "require", "allow", "deny":
			v, ok := htAccess(d, f[1:])
			if !ok {
				notes = append(notes, fmt.Sprintf("line %d: %q not supported, denying access", n, l))
			}
			if deny {
				v = htNo
			}
			if files != nil {
				files.access.set(v)
				continue
			}
			c.access.set(v)
		}
	}
	if len(open) > 0 {
		return nil, notes, fmt.Errorf("<%v> not closed", open[len(open)-1].name)
	}
	c.files = slices.DeleteFunc(c.files, func(f *htFiles) bool { return f.access == htUnset })
	return c, notes, nil
}

// htOpen parses the opening line of a container.
func htOpen(l string, inFiles bool) (htSection, error) {
	if !strings.HasSuffix(l, ">") {
		return htSection{}, errors.New("missing >")
	}
	f := htFields(l[1 : len(l)-1])
	if len(f) == 0 {
		return htSection{}, errors.New("empty <>")
	}
	s := htSection{name: f[0]}
	get := slices.ContainsFunc(f[1:], func(m string) bool { return strings.EqualFold(m, "GET") })
	switch strings.ToLower(f[0]) {
	case "files", "filesmatch":
		if inFiles {
			return s, fmt.Errorf("<%v> inside <Files>", f[0])
		}
		m, err := htMatcher(strings.EqualFold(f[0], "filesmatch"), f[1:])
		if err != nil {
			return s, err
		}
		s.files = &htFiles{match: m}
	case "limit":
		s.skip = !get
	case "limitexcept":
		s.skip = get
	case "requirenone":
		s.deny = true
	}
	return s, nil
}

// htMatcher compiles a <Files> wildcard, or the regexp of <FilesMatch> and
// <Files ~>. Case is ignored: on a case-insensitive filesystem, the macOS
// default, /SECRET.TXT opens secret.txt.
func htMatcher(re bool, args []string) (func(string) bool, error) {
	if !re && len(args) == 2 && args[0] == "~" {
		re, args = true, args[1:]
	}
	if len(args) != 1 {
		return nil, errors.New("want one pattern")
	}
	if re {
		r, err := regexp.Compile("(?i)" + args[0])
		if err != nil {
			return nil, err
		}
		return r.MatchString, nil
	}
	p := strings.ToLower(args[0])
	if _, err := path.Match(p, ""); err != nil {
		return nil, err
	}
	return func(n string) bool {
		m, _ := path.Match(p, strings.ToLower(n))
		return m
	}, nil
}

// htAccess evaluates a Require, Allow or Deny directive. ok is false for a
// rule WFM can not evaluate, which denies access.
func htAccess(d string, args []string) (v htTri, ok bool) {
	is := func(a, b string) bool {
		return len(args) == 2 && strings.EqualFold(args[0], a) && strings.EqualFold(args[1], b)
	}
	switch {
	case d == "require" && is("all", "granted"), d == "allow" && is("from", "all"):
		return htYes, true
	case d == "require" && is("all", "denied"), d == "deny" && is("from", "all"):
		return htNo, true
	}
	return htNo, false
}

// htOptions returns the Indexes setting of an Options directive. A word
// without + or - starts a new set, holding Indexes if it is listed or All
// is; +Indexes and -Indexes change it; other options leave it alone.
func htOptions(args []string) htTri {
	v, reset := htUnset, false
	for _, a := range args {
		switch {
		case strings.EqualFold(a, "+Indexes"):
			v = htYes
		case strings.EqualFold(a, "-Indexes"):
			v = htNo
		case strings.HasPrefix(a, "+"), strings.HasPrefix(a, "-"):
		default:
			if !reset {
				v, reset = htNo, true
			}
			if strings.EqualFold(a, "Indexes") || strings.EqualFold(a, "All") {
				v = htYes
			}
		}
	}
	return v
}

// htFields splits a line into words as Apache does: on blanks, with words
// in "double" or 'single' quotes holding blanks and backslash-escaped quotes.
func htFields(s string) []string {
	var out []string
	for {
		s = strings.TrimLeftFunc(s, unicode.IsSpace)
		if s == "" {
			return out
		}
		q := s[0]
		if q != '"' && q != '\'' {
			i := strings.IndexFunc(s, unicode.IsSpace)
			if i < 0 {
				i = len(s)
			}
			out, s = append(out, s[:i]), s[i:]
			continue
		}
		var w strings.Builder
		i := 1
		for ; i < len(s) && s[i] != q; i++ {
			if s[i] == '\\' && i+1 < len(s) && s[i+1] == q {
				i++
			}
			w.WriteByte(s[i])
		}
		out, s = append(out, w.String()), s[min(i+1, len(s)):]
	}
}

// htRead loads the .htaccess file of dir, nil if there is none. A file that
// can not be read or parsed denies everything below it, as Apache's 500 does.
func htRead(wfs afero.Fs, dir string) *htConf {
	p := path.Join(dir, ".htaccess")
	f, err := wfs.Open(p)
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
		return nil
	}
	var c *htConf
	if err == nil {
		var b []byte
		b, err = io.ReadAll(io.LimitReader(f, htMax+1))
		f.Close()
		if err == nil && len(b) > htMax {
			err = errors.New("file too large")
		}
		if err == nil {
			var notes []string
			c, notes, err = parseHtaccess(string(b))
			for _, n := range notes {
				log.Printf("htaccess %v: %v", p, n)
			}
		}
	}
	if err != nil {
		log.Printf("htaccess %v: %v, denying access", p, err)
		return &htConf{broken: true}
	}
	return c
}

// htDir holds the rules in effect in a directory: its .htaccess merged over
// those of its parents, the nearest setting winning.
type htDir struct {
	deny, noIndex, broken bool
	files                 []*htFiles // from the root down
}

// enter returns the rules in effect in dir, a child directory of h.
func (h htDir) enter(wfs afero.Fs, dir string) htDir {
	c := htRead(wfs, dir)
	if c == nil {
		return h
	}
	h.broken = h.broken || c.broken
	if c.access != htUnset {
		h.deny = c.access == htNo
	}
	if c.indexes != htUnset {
		h.noIndex = c.indexes == htNo
	}
	if len(c.files) > 0 {
		h.files = slices.Concat(h.files, c.files)
	}
	return h
}

// htLoad returns the rules in effect in dir.
func htLoad(wfs afero.Fs, dir string) htDir {
	h := htDir{}.enter(wfs, "/")
	p := "/"
	for _, s := range strings.Split(path.Clean("/"+dir), "/") {
		if s != "" {
			p = path.Join(p, s)
			h = h.enter(wfs, p)
		}
	}
	return h
}

func (h htDir) denied() bool {
	return h.deny || h.broken
}

// fileDenied reports whether the file name in this directory is denied. The
// last <Files> section that matches decides, else the directory's access.
// Names starting with .ht are always denied, as in Apache's default config.
func (h htDir) fileDenied(name string) bool {
	if h.broken || strings.HasPrefix(strings.ToLower(name), ".ht") {
		return true
	}
	deny := h.deny
	for _, f := range h.files {
		if f.match(name) {
			deny = f.access == htNo
		}
	}
	return deny
}

// hides reports whether the entry name of directory dir is left out of
// listings: like Apache's autoindex, they show only what may be opened.
func (h htDir) hides(wfs afero.Fs, dir, name string, isDir bool) bool {
	if isDir {
		return h.enter(wfs, path.Join(dir, name)).denied()
	}
	return h.fileDenied(name)
}

// filter drops the listing entries of dir that h hides.
func (h htDir) filter(wfs afero.Fs, dir string, dirs, files []dirEntry) ([]dirEntry, []dirEntry) {
	keep := func(es []dirEntry, isDir bool) []dirEntry {
		return slices.DeleteFunc(es, func(e dirEntry) bool {
			return h.hides(wfs, dir, e.fi.Name(), isDir)
		})
	}
	return keep(dirs, true), keep(files, false)
}

// htDenied reports whether .htaccess rules deny access to p. A path that
// does not exist is judged as a file, so a 403 does not tell which names
// exist in a denied place.
func htDenied(wfs afero.Fs, p string) bool {
	p = path.Clean("/" + p)
	if fi, err := wfs.Stat(p); err == nil && fi.IsDir() {
		return htLoad(wfs, p).denied()
	}
	return htLoad(wfs, path.Dir(p)).fileDenied(path.Base(p))
}

// htForbid renders the 403 for a request .htaccess rules refuse.
func (r *wfmRequest) htForbid() {
	htErrStatus(r.w, r.req, http.StatusForbidden, "Forbidden", "you are not allowed to access this area")
}

// htBlock renders a 403 and reports true when .htaccess rules deny this
// request access to p.
func (r *wfmRequest) htBlock(p string) bool {
	if !r.ht || !htDenied(r.fs, p) {
		return false
	}
	r.htForbid()
	return true
}
