package main

import (
	"fmt"
	"log"
	"net/http"
)

func htErr(w http.ResponseWriter, msg string, err error) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, msg, ":", err)
	log.Print(msg, err)
}

func header(w http.ResponseWriter, eDir string) {
	w.Header().Set("Content-Type", "text/html")
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
			html, body, table { width:100%; margin:0px; padding:0px; border:none; }
			td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; margin:0px; padding:2px; border:none; }
			input { border-color:#000000; border-style:none; font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; }
			.thov tr:hover { background-color: #FF8000; color: #FFFFFF; }
	--></STYLE>
	<META HTTP-EQUIV="Content-Type" CONTENT="text/html;charset=US-ASCII">
	<META HTTP-EQUIV="Content-Language" CONTENT="en-US">
	<META HTTP-EQUIV="google" CONTENT="notranslate">
	<META NAME="viewport" CONTENT="width=device-width">
	<!-- <LINK REL="icon" TYPE="image/gif" HREF="ICONGOESHERE"> -->
	</HEAD>
	<BODY BGCOLOR="#FFFFFF">
	<FORM ACTION="/" METHOD="POST" ENCTYPE="multipart/form-data">
	<INPUT TYPE="hidden" NAME="dir" VALUE="` + eDir + `">
	`))
}
