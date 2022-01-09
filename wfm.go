// Web File Manager
//
// TODO:
// * file routines: mv
// * checkboxes, multi file routines
// * resolve symlinks
// * two factor auth
// * better handle cert chdir issue
//   get and preload cert manually on start?
//   hide acm cache dir?
//   try different lib like lego?
// * rate limiter with bad auth punishment
// * favicon
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
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"golang.org/x/crypto/acme/autocert"
)

type multiString []string

var (
	vers = "2.0.1"
	addr = flag.String("addr", "127.0.0.1:8080", "Listen address, eg: 0.0.0.0:443")
	adde = flag.String("addr_extra", "", "Extra non-TLS listener address, eg: 0.0.0.0:8081")
	chdr = flag.String("chroot", "", "Path to chroot to")
	susr = flag.String("setuid", "", "User to setuid to")
	root = flag.Bool("allow_root", false, "allow to run as uid 0 / root user")
	logf = flag.String("logfile", "", "Log file name, default standard output")
	pwdf = flag.String("passwd", "", "wfm password file, eg: /usr/local/etc/wfmpw.json")
	sdot = flag.Bool("show_dot", false, "show dot files and folders")
	wpfx = flag.String("prefix", "/", "Default prefix for WFM access")
	dpfx = flag.String("http_pfx", "", "Serve regular http files at this prefix")
	ddir = flag.String("http_dir", "", "Serve regular http files from this directory")
	cctl = flag.String("cache_ctl", "no-cache", "HTTP Header Cache Control")
	adir = flag.String("acm_dir", "", "autocert cache, eg: /var/cache (affected by chroot)")
	ahwl multiString // flag set in main
	aadr = flag.String("acm_addr", "", "autocert manager listen address, eg: 0.0.0.0:80")

	users = []struct{ User, Salt, Hash string }{}
)

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

func (z *multiString) String() string {
	return "something"
}

func (z *multiString) Set(v string) error {
	*z = append(*z, v)
	return nil
}

func main() {
	var err error
	flag.Var(&ahwl, "acm_hosts", "autocert manager allowed hostnames")
	flag.Parse()

	// redirect log to file if needed
	if *logf != "" {
		lf, err := os.OpenFile(*logf, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		log.SetOutput(lf)
	}
	log.Print("WFM Starting up")

	// read password database before chroot
	if *pwdf != "" {
		pwd, err := ioutil.ReadFile(*pwdf)
		if err != nil {
			log.Fatal("unable to read password file: ", err)
		}
		err = json.Unmarshal(pwd, &users)
		if err != nil {
			log.Fatal("unable to parse password file: ", err)
		}
		log.Printf("Loaded %q (%d users)", *pwdf, len(users))
	}

	// find uid/gid for setuid before chroot
	var suid, sgid int
	if *susr != "" {
		suid, sgid, err = userId(*susr)
		if err != nil {
			log.Fatal("unable to find setuid user", err)
		}
	}

	// http handlers / mux
	mux := http.NewServeMux()
	mux.HandleFunc(*wpfx, wfm)
	mux.HandleFunc("/favicon.ico", http.NotFound)
	if *dpfx != "" && *ddir != "" {
		mux.Handle(*dpfx, http.FileServer(http.Dir(*ddir)))
	}

	// run autocert manager before chroot/setuid
	// however it doesn't matter for chroot as certs will land in chroot *adir anyway
	acm := autocert.Manager{}
	if *addr != "" && *adir != "" && len(ahwl) > 0 {
		acm.Prompt = autocert.AcceptTOS
		acm.Cache = autocert.DirCache(*adir)
		acm.HostPolicy = autocert.HostWhitelist(ahwl...)
		go http.ListenAndServe(*aadr, acm.HTTPHandler(nil))
		log.Printf("Autocert enabled")
	}

	// chroot
	if *chdr != "" {
		err := syscall.Chroot(*chdr)
		if err != nil {
			log.Fatal("chroot", err)
		}
		log.Printf("Chroot to %q", *chdr)
	}

	// listen/bind to port before setuid
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("unable to listen on %v: %v", *addr, err)
	}
	log.Printf("Listening on %q", *addr)

	// setuid now
	err = setUid(suid, sgid)
	if err != nil {
		log.Fatalf("unable to suid for %v: %v", *susr, err)
	}
	if !*root && os.Getuid() == 0 {
		log.Fatal("you probably dont want to run wfm as root, use --allow_root flag to force it")
	}
	log.Printf("Setuid UID=%d GID=%d", os.Geteuid(), os.Getgid())

	// serve http(s) as setuid user
	if *adde != "" {
		go http.ListenAndServe(*adde, mux)
	}
	if *addr != "" && *adir != "" && len(ahwl) > 0 {
		https := &http.Server{
			Addr:      *addr,
			Handler:   mux,
			TLSConfig: &tls.Config{GetCertificate: acm.GetCertificate},
		}
		log.Printf("Starting HTTPS TLS Server")
		err = https.ServeTLS(l, "", "")
	} else {
		log.Printf("Starting HTTP Server")
		err = http.Serve(l, mux)
	}
	if err != nil {
		log.Fatal(err)
	}
}
