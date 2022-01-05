// Web File Manager
//
// TODO:
// * file routines
// * checkboxes, multi file routines
// * better symlink support
// * authentication
// * favicon
// * https/certbot
// * git client
// * file locking
// * docker support (no chroot) - mount dir as / ?
// * modern browser detection
// * fancy unicode icons
// * html charset, currently US-ASCII ?!
// * better unicode icons? test on old browsers
// * generate icons on fly with encoding/gid
//   also for input type=image, or least for favicon?
// * webdav server
// * ftp server?
// * html as template
// * archive files in main view / graphical/table form
// * support for different filesystems, S3, SMB, archive files

package main

import (
	"flag"
	"html"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

var (
	vers = "2.0.1"
	addr = flag.String("addr", "127.0.0.1:8080", "Listen address and port")
	chdr = flag.String("chroot", "", "Path to chroot to")
	susr = flag.String("setuid", "", "User to setuid to")
	sdot = flag.Bool("show_dot", false, "show dot files and folders")
	wpfx = flag.String("prefix", "/", "Default prefix for WFM access")
	dpfx = flag.String("http_pfx", "", "Serve regular http files at this prefix")
	ddir = flag.String("http_dir", "", "Serve regular http files from this directory")
	cctl = flag.String("cache_ctl", "no-cache", "HTTP Header Cache Control")
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
	switch {
	case r.FormValue("mkd") != "":
		prompt(w, html.EscapeString(dir), sort, "mkdir")
		return
	case r.FormValue("mkf") != "":
		prompt(w, html.EscapeString(dir), sort, "mkfile")
		return
	case r.FormValue("mkb") != "":
		prompt(w, html.EscapeString(dir), sort, "mkurl")
		return
	case r.FormValue("upload") != "":
		f, h, err := r.FormFile("filename")
		if err != nil {
			htErr(w, "upload", err)
			return
		}
		uploadFile(w, dir, sort, h, f)
		return
	case r.FormValue("save") != "":
		saveText(w, dir, sort, html.UnescapeString(r.FormValue("fp")), r.FormValue("text"))
		return
	}

	// these fall through to directory listing
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
	case "disp":
		dispFile(w, html.UnescapeString(r.FormValue("fp")))
	case "down":
		downFile(w, html.UnescapeString(r.FormValue("fp")))
	case "edit":
		editText(w, html.UnescapeString(r.FormValue("fp")), sort)
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

func userId(usr string) (int, int, error) {
	u, err := user.Lookup(usr)
	if err != nil {
		return 0, 0, err
	}
	ui, err := strconv.Atoi(u.Uid)
	if err != nil {
		return 0, 0, err
	}
	gi, err := strconv.Atoi(u.Gid)
	if err != nil {
		return 0, 0, err
	}
	return ui, gi, nil
}

func setUid(ui, gi int) error {
	if ui == 0 || gi == 0 {
		return nil
	}
	err := syscall.Setgid(gi)
	if err != nil {
		return err
	}
	err = syscall.Setuid(ui)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var err error
	flag.Parse()

	var suid, sgid int
	if *susr != "" {
		suid, sgid, err = userId(*susr)
		if err != nil {
			log.Fatal("unable to find setuid user", err)
		}
	}
	if *chdr != "" {
		err := syscall.Chroot(*chdr)
		if err != nil {
			log.Fatal("chroot", err)
		}
		log.Printf("Chroot to %q", *chdr)
	}

	http.HandleFunc(*wpfx, wrp)
	http.HandleFunc("/favicon.ico", http.NotFound)
	if *dpfx != "" && *ddir != "" {
		http.Handle(*dpfx, http.FileServer(http.Dir(*ddir)))
	}

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("unable to listen on %v: %v", *addr, err)
	}
	log.Printf("Listening on %q", *addr)

	err = setUid(suid, sgid)
	if err != nil {
		log.Fatalf("unable to suid for %v: %v", *susr, err)
	}
	if os.Getuid() == 0 {
		log.Fatal("you probably dont want to run wfm as root")
	}
	log.Printf("Setuid UID=%d GID=%d", os.Geteuid(), os.Getgid())

	err = http.Serve(l, nil)
	if err != nil {
		log.Fatal(err)
	}
}
