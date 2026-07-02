// Web File Manager

package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"

	_ "github.com/breml/rootcerts"
	"github.com/gorilla/mux"
	"github.com/juju/ratelimit"
	"github.com/spf13/afero"
	"github.com/tenox7/tkvs"
	"golang.org/x/crypto/acme/autocert"
)

type multiString []string

type wfmPrefix struct {
	uri       string
	fs        afero.Fs
	owner     string // if set, only this user may access the prefix
	web       bool   // plain static web server mode (no WFM UI)
	index     bool   // web: serve index.html/index.htm if present
	autoIndex bool   // web: generate a directory listing
}

var (
	vers        = "2.4.0"
	bindProto   = flag.String("proto", "tcp", "tcp, tcp4, tcp6, etc")
	bindAddr    = flag.String("addr", ":8080", "Listen address, eg: :443")
	bindExtra   = flag.String("addr_extra", "", "Extra non-TLS listener address, eg: :8081")
	chrootDir   = flag.String("chroot", "", "Directory to chroot to")
	suidUser    = flag.String("setuid", "", "Username or uid:gid pair to setuid to")
	allowRoot   = flag.Bool("allow_root", false, "allow to run as uid=0/root without setuid")
	siteName    = flag.String("site_name", "WFM", "local site name to display")
	siteDesc    = flag.String("site_desc", "Web File Manager", "site description")
	logFile     = flag.String("logfile", "", "Log file name (default stdout)")
	passwdDb    = flag.String("passwd", "", "wfm password file, eg: /usr/local/etc/wfmpw.json")
	noPwdDbRW   = flag.Bool("nopass_rw", false, "allow read-write access if there is no password file")
	showDot     = flag.Bool("show_dot", false, "show dot files and folders")
	listArc     = flag.Bool("list_archive_contents", false, "list contents of archives (expensive!)")
	rateLim     = flag.Int("rate_limit", 0, "rate limit for upload/download in MB/s, 0 no limit")
	formMaxMem  = flag.Int64("form_maxmem", 10<<20, "maximum memory used for form parsing, increase for large uploads")
	defLe       = flag.String("txt_le", "LF", "default line endings when editing text files")
	textEdit    = flag.String("textedit", "textarea", "text editor in modern browsers: textarea or codemirror")
	cmCDN       = flag.String("codemirror_url", "https://cdn.jsdelivr.net/npm/codemirror@5", "CodeMirror CDN base url, npm layout, tracks latest 5.x (-textedit=codemirror)")
	convertPng  = flag.String("convertpng", "", "convert .png to gif|jpg on the fly for legacy browsers (default off)")
	dumpHeader  = flag.Bool("dump_headers", false, "dump headers sent by client")
	pfxList     multiString // this flag set in main
	webList     multiString // this flag set in main
	cacheCtl    = flag.String("cache_ctl", "no-cache", "HTTP Header Cache Control")
	acmFile     = flag.String("acm_file", "", "autocert cache, eg: /var/cache/wfm-acme.json")
	acmBind     = flag.String("acm_addr", "", "autocert manager listen address, eg: :80")
	acmWhlist   multiString // this flag set in main
	tlsCert     = flag.String("tls_cert", "", "TLS certificate file (PEM), eg: /etc/ssl/wfm.crt")
	tlsKey      = flag.String("tls_key", "", "TLS private key file (PEM), eg: /etc/ssl/wfm.key")
	fastCgi     = flag.Bool("fastcgi", false, "enable FastCGI mode")
	f2bEnabled  = flag.Bool("f2b", true, "ban ip addresses on user/pass failures")
	f2bDump     = flag.String("f2b_dump", "", "enable f2b dump at this prefix, eg. /f2bdump (default no)")
	htmlTmplDir = flag.String("html-templates", "", "directory of html templates overriding the built-in ones")
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

func emit(s string, c int) string {
	o := strings.Builder{}
	for c > 0 {
		o.WriteString(s)
		c--
	}
	return o.String()
}

func noText(m map[string][]string) map[string][]string {
	o := make(map[string][]string)
	for k, v := range m {
		if k == "text" {
			continue
		}
		o[k] = v
	}
	return o
}

func atoiOrFatal(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("WFM %v Starting up", vers)

	flag.Var(&acmWhlist, "acm_host", "autocert manager allowed hostname (multi)")
	flag.Var(&pfxList, "prefix", "Prefix for WFM access /fsdir:/httppath eg.: /var/files:/myfiles (multi, default /:/)")
	flag.Var(&webList, "webserver", "Plain static web server /fsdir:/httppath[:flags], flags i,ai eg.: /srv/www:/:i,ai (multi)")
	flag.Parse()
	var err error

	if flag.Arg(0) == "user" {
		manageUsers()
		return
	}

	switch *textEdit {
	case "textarea", "codemirror":
	default:
		log.Fatalf("--textedit %q must be 'textarea' or 'codemirror'", *textEdit)
	}

	switch *convertPng {
	case "", "gif", "jpg":
	default:
		log.Fatalf("--convertpng %q must be 'gif' or 'jpg'", *convertPng)
	}

	loadTemplates()

	if *passwdDb != "" {
		loadUsers()
	}

	if *logFile != "" {
		lf, err := os.OpenFile(*logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer lf.Close()
		log.SetOutput(lf)
	}

	// find uid/gid for setuid before chroot
	var suid, sgid int
	if *suidUser != "" {
		uidSm := regexp.MustCompile(`^(\d+):(\d+)$`).FindStringSubmatch(*suidUser)
		switch len(uidSm) {
		case 3:
			suid = atoiOrFatal(uidSm[1])
			sgid = atoiOrFatal(uidSm[2])
		default:
			suid, sgid, err = userId(*suidUser)
			if err != nil {
				log.Fatal("unable to find setuid user", err)
			}
		}
		log.Printf("Requested setuid for %q suid=%v sgid=%v", *suidUser, suid, sgid)
	}

	// run autocert manager before chroot/setuid
	acm := autocert.Manager{}
	if *bindAddr != "" && *acmFile != "" && len(acmWhlist) > 0 {
		acm.Prompt = autocert.AcceptTOS
		acm.Cache = tkvs.New(*acmFile, autocert.ErrCacheMiss)
		acm.HostPolicy = autocert.HostWhitelist(acmWhlist...)
		go http.ListenAndServe(*acmBind, acm.HTTPHandler(nil))
		log.Printf("Autocert enabled for %v", acmWhlist)
	}

	// load TLS keypair before chroot/setuid, files may live outside chroot
	var keyPair *tls.Certificate
	switch {
	case *tlsCert != "" && *tlsKey != "":
		c, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("unable to load TLS cert/key: %v", err)
		}
		keyPair = &c
		log.Printf("Loaded TLS certificate %q key %q", *tlsCert, *tlsKey)
	case *tlsCert != "" || *tlsKey != "":
		log.Fatal("both -tls_cert and -tls_key must be specified together")
	}

	// chroot now
	if *chrootDir != "" {
		err := syscall.Chroot(*chrootDir)
		if err != nil {
			log.Fatal("chroot", err)
		}
		log.Printf("Chroot to %q", *chrootDir)
	}

	// listen/bind to port before setuid
	l, err := net.Listen(*bindProto, *bindAddr)
	if err != nil {
		log.Fatalf("unable to listen on %v: %v", *bindAddr, err)
	}
	log.Printf("Listening on %q", *bindAddr)

	// setuid now
	if *suidUser != "" {
		err = setUid(suid, sgid)
		if err != nil {
			log.Fatalf("unable to suid for %v: %v", *suidUser, err)
		}
		if !*allowRoot && os.Getuid() == 0 {
			log.Fatal("you probably dont want to run wfm as root, use --allow_root flag to force it")
		}
		log.Printf("Setuid UID=%d GID=%d", os.Geteuid(), os.Getgid())
	}

	// rate limit setup
	if *rateLim != 0 {
		rlBu = ratelimit.NewBucketWithRate(float64(*rateLim<<20), 1<<10)
	}

	// http routing
	mux := mux.NewRouter()
	if len(pfxList) == 0 && len(webList) == 0 {
		pfxList = multiString{"/:/"}
	}
	var prefixes []wfmPrefix
	for _, p := range pfxList {
		s := strings.Split(p, ":")
		if len(s) != 2 || s[0][0] != '/' || s[1][0] != '/' {
			log.Fatalf("--prefix %q must be in format '/dir:/path'", p)
		}
		fs := afero.NewOsFs()
		if s[0] != "/" {
			fs = afero.NewBasePathFs(fs, s[0])
		}
		uri := strings.TrimRight(s[1], "/")
		if uri == "" {
			uri = "/"
		}
		prefixes = append(prefixes, wfmPrefix{uri: uri, fs: fs})
		log.Printf("Prefix fs=%v uri=%v", s[0], uri)
	}
	// plain static web server prefixes
	for _, p := range webList {
		wp, err := parseWebPrefix(p)
		if err != nil {
			log.Fatalf("--webserver %q %v", p, err)
		}
		prefixes = append(prefixes, wp)
		log.Printf("Webserver uri=%v index=%v autoindex=%v", wp.uri, wp.index, wp.autoIndex)
	}
	// per-user home directories become owner-restricted /username prefixes
	for _, u := range users {
		if u.Home == "" {
			continue
		}
		uri := "/" + u.User
		fs := afero.NewBasePathFs(afero.NewOsFs(), u.Home)
		prefixes = append(prefixes, wfmPrefix{uri: uri, fs: fs, owner: u.User})
		log.Printf("Prefix (home) fs=%v uri=%v owner=%v", u.Home, uri, u.User)
	}
	// reject overlapping prefixes: same uri would register twice on the mux
	// and the length-desc sort below leaves the winner nondeterministic
	seenUri := map[string]bool{}
	for _, p := range prefixes {
		if seenUri[p.uri] {
			log.Fatalf("duplicate prefix uri %q (check --prefix/--webserver/home dirs)", p.uri)
		}
		seenUri[p.uri] = true
	}
	// longest uri first so specific prefixes match before catch-all
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i].uri) > len(prefixes[j].uri)
	})
	for _, p := range prefixes {
		h := func(w http.ResponseWriter, r *http.Request) {
			wfmMain(w, r, p)
		}
		if p.web {
			h = func(w http.ResponseWriter, r *http.Request) {
				webMain(w, r, p)
			}
		}
		if p.uri == "/" {
			mux.PathPrefix("/").HandlerFunc(h)
			continue
		}
		// match only at path boundaries: exact /pfx or /pfx/... not /pfxOther
		mux.Path(p.uri).HandlerFunc(h)
		mux.PathPrefix(p.uri + "/").HandlerFunc(h)
	}
	if *f2bDump != "" {
		mux.HandleFunc(*f2bDump, dumpf2b)
	}

	if *bindExtra != "" {
		log.Printf("Listening (extra) on %q", *bindAddr)
		go http.ListenAndServe(*bindExtra, mux)
	}
	switch {
	case *acmBind != "" && *acmFile != "" && len(acmWhlist) > 0:
		https := &http.Server{
			Addr:      *bindAddr,
			Handler:   mux,
			TLSConfig: &tls.Config{GetCertificate: acm.GetCertificate},
		}
		log.Printf("Starting HTTPS TLS Server (autocert)")
		err = https.ServeTLS(l, "", "")
	case keyPair != nil:
		https := &http.Server{
			Addr:      *bindAddr,
			Handler:   mux,
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{*keyPair}},
		}
		log.Printf("Starting HTTPS TLS Server (cert/key)")
		err = https.ServeTLS(l, "", "")
	case *fastCgi:
		log.Print("Starting FastCGI Server")
		err = fcgi.Serve(l, mux)
	default:
		log.Printf("Starting HTTP Server")
		err = http.Serve(l, mux)
	}
	if err != nil {
		log.Fatal(err)
	}
}
