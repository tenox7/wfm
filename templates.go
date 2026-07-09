package main

import (
	"bufio"
	"embed"
	"html"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
)

//go:embed html/*.html
var tmplFS embed.FS

var htmlTmpl *template.Template

// stripIndent removes per-line indentation, trailing whitespace and blank
// lines from template source so they are not resent to the client with every
// page (in a big dir listing the per-row indentation adds up fast). Newlines
// are kept: between inline elements they render as the same single collapsed
// space the indentation would. Literal template text must therefore not rely
// on leading whitespace; values expanded inside <PRE>/<TEXTAREA> are not
// affected.
func stripIndent(src []byte) string {
	var b strings.Builder
	b.Grow(len(src))
	for _, l := range strings.Split(string(src), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			b.WriteString(l)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func parseTemplates(fsys fs.FS, glob string) int {
	m, err := fs.Glob(fsys, glob)
	if err != nil {
		log.Fatalf("templates %q: %v", glob, err)
	}
	for _, n := range m {
		src, err := fs.ReadFile(fsys, n)
		if err != nil {
			log.Fatalf("template %q: %v", n, err)
		}
		template.Must(htmlTmpl.New(path.Base(n)).Parse(stripIndent(src)))
	}
	return len(m)
}

func loadTemplates() {
	htmlTmpl = template.New("wfm").Option("missingkey=zero")
	parseTemplates(tmplFS, "html/*.html")
	if *htmlTmplDir == "" {
		return
	}
	n := parseTemplates(os.DirFS(*htmlTmplDir), "*.html")
	if n > 0 {
		log.Printf("Loaded %d html template override(s) from %q", n, *htmlTmplDir)
	}
}

func mode(m bool) string {
	if m {
		return "modern"
	}
	return "legacy"
}

type chrome struct {
	Dir        string
	Sort       string
	ExtraCSS   string
	FormAction string // where the page's <FORM> posts; defaults to the prefix
	SiteName   string
	SiteDesc   string
}

func (r *wfmRequest) chrome(extraCSS string) chrome {
	return chrome{
		Dir:        html.EscapeString(r.uDir),
		Sort:       html.EscapeString(r.eSort),
		ExtraCSS:   extraCSS,
		FormAction: r.pfx,
		SiteName:   *siteName,
		SiteDesc:   *siteDesc,
	}
}

func (r *wfmRequest) render(name string, data any) {
	renderStatus(r.w, r.modern, 0, name, data)
}

// renderStatus renders template name for the given browser mode. A non-zero code
// sets the HTTP status before the body (used by auth errors that must keep 401/403/429).
func renderStatus(w http.ResponseWriter, modern bool, code int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset="+charset[modern])
	w.Header().Set("Cache-Control", *cacheCtl)
	if code != 0 {
		w.WriteHeader(code)
	}
	bw := bufio.NewWriterSize(w, 1<<15)
	defer bw.Flush()
	if err := htmlTmpl.ExecuteTemplate(bw, name+"_"+mode(modern)+".html", data); err != nil {
		log.Printf("template %q: %v", name, err)
	}
}
