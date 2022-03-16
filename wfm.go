// Web File Manager

package main

import (
	"crypto/tls"
	"flag"
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
	vers      = "2.0.2"
	bindAddr  = flag.String("addr", "127.0.0.1:8080", "Listen address, eg: 0.0.0.0:443")
	bindExtra = flag.String("addr_extra", "", "Extra non-TLS listener address, eg: 0.0.0.0:8081")
	chrootDir = flag.String("chroot", "", "Path to chroot to")
	suidUser  = flag.String("setuid", "", "User to setuid to")
	allowRoot = flag.Bool("allow_root", false, "allow to run as uid 0 / root user")
	logFile   = flag.String("logfile", "", "Log file name, default standard output")
	passwdDb  = flag.String("passwd", "", "wfm password file, eg: /usr/local/etc/wfmpw.json")
	aboutRnt  = flag.Bool("about_runtime", true, "Display runtime info in About Dialog")
	showDot   = flag.Bool("show_dot", false, "show dot files and folders")
	wfmPfx    = flag.String("prefix", "/", "Default prefix for WFM access")
	docPfx    = flag.String("doc_pfx", "", "Serve regular http files at this prefix")
	docDir    = flag.String("doc_dir", "", "Serve regular http files from this directory")
	cacheCtl  = flag.String("cache_ctl", "no-cache", "HTTP Header Cache Control")
	acmDir    = flag.String("acm_dir", "", "autocert cache, eg: /var/cache (affected by chroot)")
	acmBind   = flag.String("acm_addr", "", "autocert manager listen address, eg: 0.0.0.0:80")
	acmWhlist multiString // this flag set in main

	favIcn = genFavIcon()
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
	flag.Var(&acmWhlist, "acm_hosts", "autocert manager allowed hostnames")
	flag.Parse()

	// redirect log to file if needed
	if *logFile != "" {
		lf, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		log.SetOutput(lf)
	}
	log.Print("WFM Starting up")

	// read password database before chroot
	if *passwdDb != "" {
		loadPwdDb(*passwdDb)
	}

	// find uid/gid for setuid before chroot
	var suid, sgid int
	if *suidUser != "" {
		suid, sgid, err = userId(*suidUser)
		if err != nil {
			log.Fatal("unable to find setuid user", err)
		}
	}

	// run autocert manager before chroot/setuid
	// however it doesn't matter for chroot as certs will land in chroot *adir anyway
	acm := autocert.Manager{}
	if *bindAddr != "" && *acmDir != "" && len(acmWhlist) > 0 {
		acm.Prompt = autocert.AcceptTOS
		acm.Cache = autocert.DirCache(*acmDir)
		acm.HostPolicy = autocert.HostWhitelist(acmWhlist...)
		go http.ListenAndServe(*acmBind, acm.HTTPHandler(nil))
		log.Printf("Autocert enabled")
	}

	// chroot
	if *chrootDir != "" {
		err := syscall.Chroot(*chrootDir)
		if err != nil {
			log.Fatal("chroot", err)
		}
		log.Printf("Chroot to %q", *chrootDir)
	}

	// listen/bind to port before setuid
	l, err := net.Listen("tcp", *bindAddr)
	if err != nil {
		log.Fatalf("unable to listen on %v: %v", *bindAddr, err)
	}
	log.Printf("Listening on %q", *bindAddr)

	// setuid now
	err = setUid(suid, sgid)
	if err != nil {
		log.Fatalf("unable to suid for %v: %v", *suidUser, err)
	}
	if !*allowRoot && os.Getuid() == 0 {
		log.Fatal("you probably dont want to run wfm as root, use --allow_root flag to force it")
	}
	log.Printf("Setuid UID=%d GID=%d", os.Geteuid(), os.Getgid())

	// http handlers / mux
	mux := http.NewServeMux()
	mux.HandleFunc(*wfmPfx, wfm)
	mux.HandleFunc("/favicon.ico", favicon)
	mux.HandleFunc("/dumpf2b", dumpf2b)
	if *docPfx != "" && *docDir != "" {
		log.Printf("Starting doc handler for dir %v at %v", *docDir, *docPfx)
		mux.Handle(*docPfx, http.StripPrefix(*docPfx, http.FileServer(http.Dir(*docDir))))
	}

	// serve http(s)
	if *bindExtra != "" {
		go http.ListenAndServe(*bindExtra, mux)
	}
	if *bindAddr != "" && *acmDir != "" && len(acmWhlist) > 0 {
		https := &http.Server{
			Addr:      *bindAddr,
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
