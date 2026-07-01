package main

// Plain static web server mode: serve files directly, optional index.html
// lookup and directory autoindex, no WFM management UI. Enabled per prefix
// via -webserver=/fsdir:/httppath[:flags]. Read-only, public by default.

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/spf13/afero"
)

// name column width in the Apache-style <pre> autoindex listing
const webNameCol = 44

type webIndexRow struct {
	Href, Name, Pad, Meta string
}

type webIndexPage struct {
	Dir, Header, UpTail, Vers string
	Up                        bool
	Rows                      []webIndexRow
}

// webCol returns the html-escaped (possibly truncated) display name and the
// space padding that right-fills the name column. Padding is computed on the
// rune length of the visible name, not the escaped bytes, so <pre> aligns.
func webCol(disp string) (string, string) {
	r := []rune(disp)
	if len(r) > webNameCol-1 {
		disp = string(r[:webNameCol-3]) + ".."
		r = []rune(disp)
	}
	return html.EscapeString(disp), strings.Repeat(" ", webNameCol-len(r))
}

func webMeta(modTime, size string) string {
	return fmt.Sprintf("%-16s  %8s", modTime, size)
}

func webMain(w http.ResponseWriter, r *http.Request, p wfmPrefix) {
	upath := path.Clean("/" + strings.TrimPrefix(r.URL.Path, p.uri))
	fi, err := p.fs.Stat(upath)
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	if !fi.IsDir() {
		serveWebFile(w, r, p.fs, upath, fi)
		return
	}

	// directories need a trailing slash so relative links resolve
	if !strings.HasSuffix(r.URL.Path, "/") {
		target := r.URL.Path + "/"
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}

	if p.index {
		for _, idx := range []string{"index.html", "index.htm"} {
			ip := path.Join(upath, idx)
			ifi, err := p.fs.Stat(ip)
			if err == nil && !ifi.IsDir() {
				serveWebFile(w, r, p.fs, ip, ifi)
				return
			}
		}
	}

	if p.autoIndex {
		webAutoIndex(w, r, p, upath)
		return
	}

	http.Error(w, "403 Forbidden", http.StatusForbidden)
}

func serveWebFile(w http.ResponseWriter, r *http.Request, wfs afero.Fs, upath string, fi os.FileInfo) {
	f, err := wfs.Open(upath)
	if err != nil {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	defer f.Close()
	w.Header().Set("Cache-Control", *cacheCtl)
	// ServeContent handles Content-Type, Range, HEAD and If-Modified-Since.
	// serveContent applies -rate_limit on the write side; seeking stays on the
	// file so ranges still work.
	serveContent(w, r, f, fi)
}

func webAutoIndex(w http.ResponseWriter, r *http.Request, p wfmPrefix, upath string) {
	d, err := afero.ReadDir(p.fs, upath)
	if err != nil {
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	var dirs, files []webIndexRow
	for _, fi := range d {
		if !*showDot && strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		isDir := fi.IsDir()
		if fi.Mode()&os.ModeSymlink != 0 {
			ls, err := p.fs.Stat(path.Join(upath, fi.Name()))
			if err != nil {
				continue
			}
			isDir = ls.IsDir()
		}
		name := fi.Name()
		href := url.PathEscape(fi.Name())
		size := humanize.Bytes(uint64(fi.Size()))
		if isDir {
			name += "/"
			href += "/"
			size = "-"
		}
		escName, pad := webCol(name)
		row := webIndexRow{
			Href: html.EscapeString(href),
			Name: escName,
			Pad:  pad,
			Meta: webMeta(fi.ModTime().Format("2006-01-02 15:04"), size),
		}
		if isDir {
			dirs = append(dirs, row)
			continue
		}
		files = append(files, row)
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name < dirs[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	_, hdrPad := webCol("Name")
	page := webIndexPage{
		Dir:    html.EscapeString(r.URL.Path),
		Header: "Name" + hdrPad + webMeta("Last modified", "Size"),
		Vers:   vers,
		Rows:   append(dirs, files...),
	}
	if upath != "/" {
		_, upPad := webCol("Parent Directory")
		page.Up = true
		page.UpTail = upPad + webMeta("", "-")
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Cache-Control", *cacheCtl)
	if err := htmlTmpl.ExecuteTemplate(w, "autoindex.html", page); err != nil {
		log.Printf("autoindex template: %v", err)
	}
}

// parseWebPrefix parses a -webserver spec /fsdir:/httppath[:flags].
// flags are comma separated and opt-in: i (index.html/.htm), ai (autoindex).
// No flags (absent or empty field) serves only explicit file URLs, dirs 403.
func parseWebPrefix(spec string) (wfmPrefix, error) {
	s := strings.SplitN(spec, ":", 3)
	if len(s) < 2 || s[0] == "" || s[1] == "" || s[0][0] != '/' || s[1][0] != '/' {
		return wfmPrefix{}, errBadWebSpec
	}
	fs := afero.NewOsFs()
	if s[0] != "/" {
		fs = afero.NewBasePathFs(fs, s[0])
	}
	uri := strings.TrimRight(s[1], "/")
	if uri == "" {
		uri = "/"
	}
	wp := wfmPrefix{uri: uri, fs: fs, web: true}
	if len(s) == 3 {
		for _, fl := range strings.Split(s[2], ",") {
			switch strings.TrimSpace(fl) {
			case "i":
				wp.index = true
			case "ai":
				wp.autoIndex = true
			case "":
			default:
				return wfmPrefix{}, fmt.Errorf("unknown flag %q (want i, ai)", fl)
			}
		}
	}
	return wp, nil
}

var errBadWebSpec = fmt.Errorf("must be in format '/dir:/path[:flags]'")
