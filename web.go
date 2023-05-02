package main

import (
	_ "embed"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

var (
	//go:embed favicon.ico
	favIcn []byte
	disTag = map[bool]string{
		true:  "",
		false: "DISABLED",
	}
	charset = map[bool]string{
		true:  "UTF-8",
		false: "ISO-8859-1",
	}
	padding = map[bool]string{
		true:  "2px",
		false: "0px",
	}
)

func htErr(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cacheCtl)
	fmt.Fprintln(w, msg, ":", err)
	log.Printf("error: %v : %v", msg, err)
}

func header(w http.ResponseWriter, uDir, sort, extraCSS string, modern bool) {
	eDir := html.EscapeString(uDir)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", *cacheCtl)
	w.Write([]byte(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<HTML LANG="en">
<HEAD>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html;charset=` + charset[modern] + `">
<META HTTP-EQUIV="Content-Language" CONTENT="en-US">
<META HTTP-EQUIV="google" CONTENT="notranslate">
<META HTTP-EQUIV="charset" CONTENT="` + charset[modern] + `">
<META HTTP-EQUIV="encoding" CONTENT="` + charset[modern] + `">
<META NAME="viewport" CONTENT="width=device-width">
<LINK REL="icon" TYPE="image/x-icon" HREF="/favicon.ico">
<LINK REL="shortcut icon" HREF="/favicon.ico?">
<TITLE>` + *siteName + ` : ` + eDir + `</TITLE>
<STYLE TYPE="text/css"><!--
	A:link {text-decoration: none; color:#0000CE; }
	A:visited {text-decoration: none; color:#0000CE; }
	A:active {text-decoration: none; color:#FF0000; }
	A:hover {text-decoration: none; background-color: #FF8000; color: #FFFFFF; }
	html, body, table { margin:0px; padding:0px; border:none;  }
	td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; margin:0px; padding:` + padding[modern] + `; border:none; }
	input { border-color:#000000; border-style:solid; font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; }
	.thov tr:hover { background-color: #FF8000; color: #FFFFFF; }
	.tbr { border-width: 1px; border-style: solid solid solid solid; border-color: #AAAAAA #555555 #555555 #AAAAAA; }
	.nb { border-style:none; background-color: #EEEEEE; }
	` + extraCSS + `
--></STYLE>
</HEAD>
<BODY BGCOLOR="#FFFFFF">
<FORM ACTION="` + wfmPfx + `" METHOD="POST" ENCTYPE="multipart/form-data">
<INPUT TYPE="hidden" NAME="dir" VALUE="` + eDir + `">
<INPUT TYPE="hidden" NAME="sort" VALUE="` + sort + `">
`))
}

func footer(w http.ResponseWriter) {
	w.Write([]byte("\n</FORM></BODY></HTML>\n"))
}

func redirect(w http.ResponseWriter, uUrl string) {
	w.Header().Set("Location", uUrl)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", *cacheCtl)
	w.WriteHeader(302)

	w.Write([]byte(`<HTML><BODY>
	<A HREF="` + html.EscapeString(uUrl) + `">Go here...</A>
    </BODY></HTML>
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
