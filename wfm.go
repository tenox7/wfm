// Web File Manager
//
// TODO:
// * file routines
// * checkboxes for files
// * authentication
// * setuid/setgid
// * https/certbot
// * git client
// * docker support (no chroot) - mount dir as / ?
// * drivers for different storage, like cloud/smb/ftp
// * html charset, currently US-ASCII ?!
// * generate icons on fly with encoding/gid
//   also for input type=image, or  least for favicon?
// * time/date format as flag?
// * webdav server
// * ftp server?
// * html as template

package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"syscall"
)

var (
	addr = flag.String("addr", ":8080", "Listen address and port")
	base = flag.String("base_dir", "", "Base directory path")
)

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

func wrp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	dir := filepath.Clean(r.FormValue("dir"))
	if r.FormValue("home") != "" {
		dir = "/"
	}
	if r.FormValue("up") != "" {
		dir = filepath.Dir(dir)
	}
	if dir == "" {
		dir = "/"
	}
	log.Printf("req from=%q uri=%q", r.RemoteAddr, r.RequestURI)

	switch r.FormValue("fn") {
	default:
		listFiles(w, dir, r.FormValue("sort"))
		return
	}
}

func main() {
	flag.Parse()
	var err error
	if *base != "" {
		err = syscall.Chroot(*base)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Starting WFM on %q for directory %q", *addr, *base)

	http.HandleFunc("/", wrp)
	http.HandleFunc("/favicon.ico", http.NotFound)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
