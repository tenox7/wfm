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

var (
	vers       = "2.2.4"
	bindProto  = flag.String("proto", "tcp", "tcp, tcp4, tcp6, etc")
	bindAddr   = flag.String("addr", ":8080", "Listen address, eg: :443")
	bindExtra  = flag.String("addr_extra", "", "Extra non-TLS listener address, eg: :8081")
	chrootDir  = flag.String("chroot", "", "Directory to chroot to")
	suidUser   = flag.String("setuid", "", "Username or uid:gid pair to setuid to")
	allowRoot  = flag.Bool("allow_root", false, "allow to run as uid=0/root without setuid")
	siteName   = flag.String("site_name", "WFM", "local site name to display")
	siteDesc   = flag.String("site_desc", "Web File Manager", "site description")
	logFile    = flag.String("logfile", "", "Log file name (default stdout)")
	passwdDb   = flag.String("passwd", "", "wfm password file, eg: /usr/local/etc/wfmpw.json")
	noPwdDbRW  = flag.Bool("nopass_rw", false, "allow read-write access if there is no password file")
	aboutRnt   = flag.Bool("about_runtime", true, "Display runtime info in About Dialog")
	showDot    = flag.Bool("show_dot", false, "show dot files and folders")
	listArc    = flag.Bool("list_archive_contents", false, "list contents of archives (expensive!)")
	rateLim    = flag.Int("rate_limit", 0, "rate limit for upload/download in MB/s, 0 no limit")
	formMaxMem = flag.Int64("form_maxmem", 10<<20, "maximum memory used for form parsing, increase for large uploads")
	prefix     = flag.String("prefix", "/:/", "Prefix for WFM access, /fsdir:/httppath eg.: /var/files:/myfiles")
	defLe      = flag.String("txt_le", "LF", "default line endings when editing text files")
	dumpHeader = flag.Bool("dump_headers", false, "dump headers sent by client")
	wfmFs      afero.Fs
	wfmPfx     string
	cacheCtl   = flag.String("cache_ctl", "no-cache", "HTTP Header Cache Control")
	acmFile    = flag.String("acm_file", "", "autocert cache, eg: /var/cache/wfm-acme.json")
	acmBind    = flag.String("acm_addr", "", "autocert manager listen address, eg: :80")
	acmWhlist  multiString // this flag set in main
	fastCgi    = flag.Bool("fastcgi", false, "enable FastCGI mode")
	f2bEnabled = flag.Bool("f2b", true, "ban ip addresses on user/pass failures")
	f2bDump    = flag.String("f2b_dump", "", "enable f2b dump at this prefix, eg. /f2bdump (default no)")
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
	flag.Parse()
	var err error

	if flag.Arg(0) == "user" {
		manageUsers()
		return
	}

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
	pfx := strings.Split(*prefix, ":")
	if len(pfx) != 2 || pfx[0][0] != '/' || pfx[1][0] != '/' {
		log.Fatal("--prefix must be in format '/dir:/path'")
	}
	log.Printf("Prefix fs=%v uri=%v", pfx[0], pfx[1])
	wfmFs = afero.NewOsFs()
	if pfx[0] != "/" {
		wfmFs = afero.NewBasePathFs(wfmFs, pfx[0])
	}
	wfmPfx = pfx[1]
	mux.PathPrefix(wfmPfx).HandlerFunc(wfmMain)
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
		log.Printf("Starting HTTPS TLS Server")
		err = https.ServeTLS(l, "", "")
	case *fastCgi:
		log.Print("Starting FastCGI Server")
		fcgi.Serve(l, http.DefaultServeMux)
	default:
		log.Printf("Starting HTTP Server")
		err = http.Serve(l, mux)
	}
	if err != nil {
		log.Fatal(err)
	}
}
