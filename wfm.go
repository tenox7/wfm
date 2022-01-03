// Web File Manager
//
// TODO:
// * file routines
// * multi file routines
// * symlink support?
// * checkboxes for files
// * favicon
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
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
)

var (
	addr = flag.String("addr", ":8080", "Listen address and port")
	chdr = flag.String("chroot", "", "Path to cheroot to")
	susr = flag.String("setuid", "", "User to setuid to")
	disp = flag.String("disp", "open", "default disposition when you click on a file: open|save|edit")
	sdot = flag.Bool("show_dot", false, "show dot files and folders")
)

func wrp(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	log.Printf("req from=%q uri=%q form=%v", r.RemoteAddr, r.RequestURI, r.Form)

	dir := filepath.Clean(html.UnescapeString(r.FormValue("dir")))
	if dir == "" || dir == "." {
		dir = "/"
	}
	sort := r.FormValue("sort")

	// toolbar buttons
	if r.FormValue("mkd") != "" {
		prompt(w, html.EscapeString(dir), sort, "mkdir")
		return
	}
	if r.FormValue("mkf") != "" {
		prompt(w, html.EscapeString(dir), sort, "mkfile")
		return
	}
	if r.FormValue("mkb") != "" {
		prompt(w, html.EscapeString(dir), sort, "mkurl")
		return
	}
	if r.FormValue("home") != "" {
		dir = "/"
	}
	if r.FormValue("up") != "" {
		dir = filepath.Dir(dir)
	}

	// cancel
	if r.FormValue("cancel") != "" {
		r.Form.Set("fn", "")
	}

	// form action
	switch r.FormValue("fn") {
	case "di":
		fileDisp(w, html.UnescapeString(r.FormValue("fi")), "inline")
	case "dn":
		f := html.UnescapeString(r.FormValue("fi"))
		fileDisp(w, f, "attachment; filename=\""+path.Base(f)+"\"")
	case "mkdir":
		mkdir(w, dir, html.UnescapeString(r.FormValue("newd")), sort)
	case "mkfile":
		mkfile(w, dir, html.UnescapeString(r.FormValue("newf")), sort)
	case "mkurl":
		mkurl(w, dir, html.UnescapeString(r.FormValue("newu")), r.FormValue("url"), sort)
	default:
		listFiles(w, dir, sort)
	}
}

func chroot(dir string) {
	err := syscall.Chroot(dir)
	if err != nil {
		log.Fatal("chroot", err)
	}
	log.Printf("Chroot to %q", dir)
}

func setuid(usr string) {
	u, err := user.Lookup(usr)
	if err != nil {
		log.Fatal("unable to find user", err)
	}
	gi, err := strconv.Atoi(u.Gid)
	if err != nil {
		log.Fatal("convert gid", err)

	}
	err = syscall.Setgid(gi)
	if err != nil {
		log.Fatal("setgid", err)
	}
	ui, err := strconv.Atoi(u.Uid)
	if err != nil {
		log.Fatal("convert uid", err)
	}
	err = syscall.Setuid(ui)
	if err != nil {
		log.Fatal("setuid", err)
	}
	log.Printf("Setuid as %q", usr)
}

func main() {
	flag.Parse()
	if *chdr != "" {
		chroot(*chdr)
	}
	if *susr != "" {
		setuid(*susr)
	}

	http.HandleFunc("/", wrp)
	http.HandleFunc("/favicon.ico", http.NotFound)
	log.Printf("Listening on %q", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
