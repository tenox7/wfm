package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var (
	charset = map[bool]string{
		true:  "UTF-8",
		false: "ISO-8859-1",
	}
	panelGrey = map[bool]string{
		true:  "#EEEEEE",
		false: "#CCCCCC",
	}
)

func wfmURL(base string, q url.Values) string {
	e := q.Encode()
	if e == "" {
		return base
	}
	return base + "?" + e
}

func wfmHref(base string, q url.Values) string {
	return html.EscapeString(wfmURL(base, q))
}

func htErr(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cacheCtl)
	fmt.Fprintln(w, msg, ":", err)
	log.Printf("error: %v : %v", msg, err)
}

func redirect(w http.ResponseWriter, uUrl string) {
	w.Header().Set("Location", uUrl)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", *cacheCtl)
	w.WriteHeader(302)

	u := html.EscapeString(uUrl)
	w.Write([]byte(`<HTML>
	<HEAD>
	<META HTTP-EQUIV="refresh" CONTENT="0; URL=` + u + `">
	</HEAD>
	<BODY>
	If you see this, your browser did not redirect. <A HREF="` + u + `">Click here...</A>
    </BODY>
	</HTML>
    `))
}

func upDnDir(uDir, uBn string, wfs afero.Fs) string {
	o := strings.Builder{}
	o.WriteString("<OPTION VALUE=\"/\">/ - Root</OPTION>\n")
	p := "/"
	i := 0
	for _, n := range strings.Split(uDir, string(os.PathSeparator))[1:] {
		p = p + n + "/"
		opt := ""
		if p == uDir+"/" {
			opt = "DISABLED"
		}
		i++
		o.WriteString("<OPTION " + opt + " VALUE=\"" +
			html.EscapeString(filepath.Clean(p+"/"+uBn)) + "\">" +
			emit("&nbsp;&nbsp;", i) + " L " +
			html.EscapeString(n) + "</OPTION>\n")
	}
	d, err := afero.ReadDir(wfs, uDir)
	if err != nil {
		return o.String()
	}
	for _, n := range d {
		if !n.IsDir() || strings.HasPrefix(n.Name(), ".") {
			continue
		}
		o.WriteString("<OPTION VALUE=\"" +
			html.EscapeString(uDir+"/"+n.Name()+"/"+uBn) + "\">" +
			emit("&nbsp;&nbsp;&nbsp;", i) + " L " +
			html.EscapeString(n.Name()) + "</OPTION>\n")
	}
	return o.String()
}
