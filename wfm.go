// Web File Manager
//
// TODO:
// * file routines
// * symlink support?
// * checkboxes for files
// * authentication
// * setuid/setgid
// * https/certbot
// * git client
// * docker support (no chroot) - mount dir as / ?
// * drivers for different storage, like cloud/smb/ftp
// * html charset, currently US-ASCII ?!
// * better unicode icons? test on old browsers
// * generate icons on fly with encoding/gid
//   also for input type=image, or  least for favicon?
// * time/date format as flag?
// * webdav server
// * ftp server?
// * html as template

package main

import (
	"flag"
	"html"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"syscall"
)

var (
	addr = flag.String("addr", ":8080", "Listen address and port")
	base = flag.String("base_dir", "", "Base directory path")
	disp = flag.String("disp", "open", "default disposition when you click on a file: open|save|edit")
	sdot = flag.Bool("show_dot", false, "show dot files and folders")
)

func wrp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	dir := filepath.Clean(html.UnescapeString(r.FormValue("dir")))
	if r.FormValue("home") != "" {
		dir = "/"
	}
	if r.FormValue("up") != "" {
		dir = filepath.Dir(dir)
	}
	if dir == "" || dir == "." {
		dir = "/"
	}
	log.Printf("req from=%q uri=%q", r.RemoteAddr, r.RequestURI)

	switch r.FormValue("fn") {
	case "di":
		fileDisp(w, html.UnescapeString(r.FormValue("fi")), "inline")
	case "dn":
		f := html.UnescapeString(r.FormValue("fi"))
		fileDisp(w, f, "attachment; filename=\""+path.Base(f)+"\"")
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
