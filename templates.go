package main

import (
	"bufio"
	"embed"
	"html"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

//go:embed html/*.html
var tmplFS embed.FS

var htmlTmpl *template.Template

func loadTemplates() {
	htmlTmpl = template.Must(template.New("wfm").Option("missingkey=zero").ParseFS(tmplFS, "html/*.html"))
	if *htmlTmplDir == "" {
		return
	}
	m, err := filepath.Glob(filepath.Join(*htmlTmplDir, "*.html"))
	if err != nil {
		log.Fatalf("html-templates %q: %v", *htmlTmplDir, err)
	}
	if len(m) == 0 {
		return
	}
	htmlTmpl = template.Must(htmlTmpl.ParseFiles(m...))
	log.Printf("Loaded %d html template override(s) from %q", len(m), *htmlTmplDir)
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
