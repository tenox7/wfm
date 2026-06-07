package main

import (
	"bufio"
	"embed"
	"html"
	"log"
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
	Dir      string
	Sort     string
	ExtraCSS string
	Pfx      string
	SiteName string
	SiteDesc string
}

func (r *wfmRequest) chrome(extraCSS string) chrome {
	return chrome{
		Dir:      html.EscapeString(r.uDir),
		Sort:     html.EscapeString(r.eSort),
		ExtraCSS: extraCSS,
		Pfx:      r.pfx,
		SiteName: *siteName,
		SiteDesc: *siteDesc,
	}
}

func (r *wfmRequest) render(name string, data any) {
	r.w.Header().Set("Content-Type", "text/html; charset="+charset[r.modern])
	r.w.Header().Set("Cache-Control", *cacheCtl)
	bw := bufio.NewWriterSize(r.w, 1<<15)
	defer bw.Flush()
	if err := htmlTmpl.ExecuteTemplate(bw, name+"_"+mode(r.modern)+".html", data); err != nil {
		log.Printf("template %q: %v", name, err)
	}
}
