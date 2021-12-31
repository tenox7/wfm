// Web File Manager
//
// TODO:
// * dilist sorting
// * dirlist alternate line colors
// * file routines
// * authentication
// * setuid/setgid
// * https/certbot
// * git client
// * docker support (no chroot) - mount dir as / ?
// * drivers for different storage, like cloud/smb/ftp
// * html charset, currently US-ASCII ?!
// * generate icons on fly with encoding/gid
//   also for input type=image, or  least for favicon?
// time/date format as flag?

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

func header(w http.ResponseWriter, dir string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(
		"<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n\"http://www.w3.org/TR/html4/loose.dtd\">\n" +
			"<HTML LANG=\"en\">\n" +
			"<HEAD>\n" +
			"<TITLE>WFM " + dir + "</TITLE>\n" +
			"<STYLE TYPE=\"text/css\">\n<!--\n" +
			"A:link {text-decoration: none; color:#0000CE; } \n" +
			"A:visited {text-decoration: none; color:#0000CE; } \n" +
			"A:active {text-decoration: none; color:#FF0000; } \n" +
			"A:hover {text-decoration: none; color:#FF0000; } \n" +
			"html, body, table { width:100%%; margin:0px; padding:0px; border:none; } \n" +
			"td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; margin:0px; padding:2px; border:none; } \n" +
			"input { border-color:#000000; border-style:none; font-family: Tahoma, Arial, Geneva, sans-serif; font-size:13px; }\n" +
			".hovout { border: none; padding: 0px; background-color: transparent; color: #0000CE; }\n" +
			".hovin  { border: none; padding: 0px; background-color: transparent; color: #FF0000; }\n" +
			"-->\n</STYLE>\n" +
			"<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html;charset=US-ASCII\">\n" +
			"<META HTTP-EQUIV=\"Content-Language\" CONTENT=\"en-US\">\n" +
			"<META HTTP-EQUIV=\"google\" CONTENT=\"notranslate\">\n" +
			"<META NAME=\"viewport\" CONTENT=\"width=device-width\">\n" +
			/*"<LINK REL=\"icon\" TYPE=\"image/gif\" HREF=\"ICONGOESHERE\">\n" +*/
			"</HEAD>\n" +
			"<BODY BGCOLOR=\"#FFFFFF\">\n" +
			"<FORM ACTION=\"/\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n" +
			"<INPUT TYPE=\"hidden\" NAME=\"dir\" VALUE=\"" + dir + "\">\n",
	))
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
