package main

import (
	"fmt"
	"log"
	"net/http"
)

func htErr(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cctl)
	fmt.Fprintln(w, msg, ":", err)
	log.Printf("error: %v : %v", msg, err)
}

func header(w http.ResponseWriter, eDir, sort string) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", *cctl)
	w.Write([]byte(`
    <!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
    <HTML LANG="en">
    <HEAD>
    <TITLE>WFM ` + eDir + `</TITLE>
    <STYLE TYPE="text/css"><!--
            A:link {text-decoration: none; color:#0000CE; }
            A:visited {text-decoration: none; color:#0000CE; }
            A:active {text-decoration: none; color:#FF0000; }
            A:hover {text-decoration: none; background-color: #FF8000; color: #FFFFFF; }
            html, body, table { margin:0px; padding:0px; border:none;  }
            td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; margin:0px; padding:2px; border:none; }
            input { border-color:#000000; border-style:solid; font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; }
            .thov tr:hover { background-color: #FF8000; color: #FFFFFF; }
            .tbr { border-width: 1px; border-style: solid solid solid solid; border-color: #AAAAAA #555555 #555555 #AAAAAA; }
            .nb { border-style:none; }
    --></STYLE>
    <META HTTP-EQUIV="Content-Type" CONTENT="text/html;charset=US-ASCII">
    <META HTTP-EQUIV="Content-Language" CONTENT="en-US">
    <META HTTP-EQUIV="google" CONTENT="notranslate">
    <META NAME="viewport" CONTENT="width=device-width">
    <!-- <LINK REL="icon" TYPE="image/gif" HREF="ICONGOESHERE"> -->
    </HEAD>
    <BODY BGCOLOR="#FFFFFF">
    <FORM ACTION="` + *wpfx + `" METHOD="POST" ENCTYPE="multipart/form-data">
    <INPUT TYPE="hidden" NAME="dir" VALUE="` + eDir + `">
    <INPUT TYPE="hidden" NAME="sort" VALUE="` + sort + `">
    `))
}

func footer(w http.ResponseWriter) {
	w.Write([]byte(`
    </FORM></BODY></HTML>
    `))
}

func redirect(w http.ResponseWriter, url string) {
	w.Header().Set("Location", url)
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Cache-Control", *cctl)
	w.WriteHeader(302)

	w.Write([]byte(`
    <HTML><BODY>
    <A HREF="` + url + `">Go here...</A>
    </BODY></HTML>
    `))
}
