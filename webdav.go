package main

// WebDAV (RFC 4918) via golang.org/x/net/webdav, served on every WFM prefix
// alongside the web ui at the same url, unless -webdav=false. Only the methods
// the ui does not use are routed here, so a browser still gets the file
// manager. Auth is the ui's: password db users with their ro/rw flag,
// -anon_ro, f2b and owner checks. -webserver prefixes are not served.

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

// davMaxLock caps lock lifetime. x/net/webdav maps a missing Timeout header to
// an infinite duration and its memory lock system never expires those, so a
// single LOCK would block every other writer until the process restarts.
const davMaxLock = 30 * time.Minute

// net/http has no constants for the DAV methods
const (
	methodPropfind  = "PROPFIND"
	methodProppatch = "PROPPATCH"
	methodMkcol     = "MKCOL"
	methodLock      = "LOCK"
	methodUnlock    = "UNLOCK"
	methodCopy      = "COPY"
	methodMove      = "MOVE"
)

// davMethods are served by webdav instead of the ui. GET, HEAD and POST are
// missing on purpose: they are how browsers reach the file manager, and a dav
// client reading a file with GET gets its bytes from the ui anyway.
var davMethods = map[string]bool{
	http.MethodOptions: true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
	methodPropfind:     true,
	methodProppatch:    true,
	methodMkcol:        true,
	methodLock:         true,
	methodUnlock:       true,
	methodCopy:         true,
	methodMove:         true,
}

type davCtxKey int

const davMethod davCtxKey = 0

// davClean confines a name to the served tree. The url path is already cleaned
// by the mux, but the COPY/MOVE Destination header is not: x/net/webdav hands
// it to the FileSystem verbatim, and afero's BasePathFs only does a string
// prefix check, so "/../sibling" would escape into any directory whose name
// shares the base prefix. Cleaning also maps the empty name that stripPrefix
// returns for a request to the bare prefix onto the root.
func davClean(name string) string {
	return path.Clean("/" + name)
}

// davMutate cleans a name that is about to be removed or renamed, refusing the
// served root itself: a destination like "/dav/.." cleans to the mount point,
// and webdav removes an overwrite destination before writing it, which would
// recursively delete the whole share and replace it with a file.
func davMutate(name string) (string, error) {
	c := davClean(name)
	if c == "/" {
		return "", os.ErrPermission
	}
	return c, nil
}

// davStrip removes a mount prefix from an url path at a path boundary and
// cleans the remainder. webdav's own prefix strip is a plain TrimPrefix, so a
// sibling mount (/dav/alice2 against a /dav/alice prefix) would be trimmed to
// "2" and written inside this share instead of being rejected.
func davStrip(pfx, p string) (string, bool) {
	if pfx == "" {
		return davClean(p), true
	}
	if p == pfx {
		return "/", true
	}
	if !strings.HasPrefix(p, pfx+"/") {
		return "", false
	}
	return davClean(strings.TrimPrefix(p, pfx)), true
}

// davUnder reports whether child sits strictly below parent.
func davUnder(child, parent string) bool {
	if parent == "/" {
		return child != "/"
	}
	return strings.HasPrefix(child, parent+"/")
}

// davDest validates the COPY/MOVE Destination against the source and rewrites
// it in canonical form. webdav compares the two as the raw strings the client
// sent, so any alias of the source (a trailing slash, a dot segment) slips past
// its same-resource check and reaches the RemoveAll it performs before writing
// the destination: the source is recursively deleted and the client is told it
// succeeded. A destination below the source recurses until the filesystem
// refuses the path length.
func davDest(w http.ResponseWriter, r *http.Request, pfx string) bool {
	if r.Method != methodCopy && r.Method != methodMove {
		return true
	}
	u, err := url.Parse(r.Header.Get("Destination"))
	if err != nil || u.Path == "" {
		return true // malformed, let webdav report it
	}
	src, ok := davStrip(pfx, r.URL.Path)
	if !ok {
		return true // not under this mount, webdav will 404
	}
	dst, ok := davStrip(pfx, u.Path)
	if !ok {
		http.Error(w, "502 Bad Gateway", http.StatusBadGateway)
		return false
	}
	if dst == src || davUnder(src, dst) || davUnder(dst, src) {
		log.Printf("dav: refused %v of %q onto %q", r.Method, src, dst)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return false
	}
	u.Path, u.RawPath = pfx+dst, ""
	r.Header.Set("Destination", u.String())
	return true
}

// davFS adapts afero.Fs to webdav.FileSystem; afero.File already satisfies
// webdav.File.
type davFS struct{ fs afero.Fs }

func (d davFS) Mkdir(_ context.Context, name string, perm os.FileMode) error {
	return d.fs.Mkdir(davClean(name), perm)
}

func (d davFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := d.fs.OpenFile(davClean(name), flag, perm)
	if err != nil {
		return nil, err
	}
	// hide dot entries from listings only: webdav reads directories through
	// this same handle when recursively copying, where filtering would
	// silently drop them from the copy
	if *showDot || ctx.Value(davMethod) != methodPropfind {
		return f, nil
	}
	return noDotFile{f}, nil
}

func (d davFS) RemoveAll(_ context.Context, name string) error {
	n, err := davMutate(name)
	if err != nil {
		return err
	}
	return d.fs.RemoveAll(n)
}

func (d davFS) Rename(_ context.Context, oldName, newName string) error {
	o, err := davMutate(oldName)
	if err != nil {
		return err
	}
	n, err := davMutate(newName)
	if err != nil {
		return err
	}
	return d.fs.Rename(o, n)
}

func (d davFS) Stat(_ context.Context, name string) (os.FileInfo, error) {
	return d.fs.Stat(davClean(name))
}

type noDotFile struct{ afero.File }

func (f noDotFile) Readdir(count int) ([]os.FileInfo, error) {
	fis, err := f.File.Readdir(count)
	out := fis[:0]
	for _, fi := range fis {
		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}
		out = append(out, fi)
	}
	return out, err
}

// davLS enforces davMaxLock on stored locks. davTimeout already clamps the
// header webdav derives the duration from; this is the backstop that keeps a
// non-expiring lock impossible even if the two ever disagree.
type davLS struct{ webdav.LockSystem }

func (l davLS) Create(now time.Time, d webdav.LockDetails) (string, error) {
	d.Duration = davLockDur(d.Duration)
	return l.LockSystem.Create(now, d)
}

func (l davLS) Refresh(now time.Time, token string, dur time.Duration) (webdav.LockDetails, error) {
	return l.LockSystem.Refresh(now, token, davLockDur(dur))
}

func davLockDur(d time.Duration) time.Duration {
	if d < 0 || d > davMaxLock {
		return davMaxLock
	}
	return d
}

// davTimeout clamps the LOCK Timeout header to davMaxLock. webdav renders the
// granted lifetime in its response from this header rather than from what the
// lock system stored, so without this a client is told it holds a lock for
// longer (or forever, reported as Second-0) than the server will honor.
// A malformed value is left alone for webdav to reject.
func davTimeout(r *http.Request) {
	if r.Method != methodLock {
		return
	}
	max := fmt.Sprintf("Second-%d", davMaxLock/time.Second)
	s, _, _ := strings.Cut(r.Header.Get("Timeout"), ",")
	s = strings.TrimSpace(s)
	if s == "" || s == "Infinite" {
		r.Header.Set("Timeout", max)
		return
	}
	if !strings.HasPrefix(s, "Second-") {
		return
	}
	n, err := strconv.ParseInt(strings.TrimPrefix(s, "Second-"), 10, 64)
	if err == nil && time.Duration(n)*time.Second > davMaxLock {
		r.Header.Set("Timeout", max)
	}
}

// methods allowed for read-only access
var davRO = map[string]bool{
	http.MethodOptions: true,
	http.MethodGet:     true,
	http.MethodHead:    true,
	methodPropfind:     true,
}

// newDavHandler builds the webdav handler for a prefix, nil when -webdav=false.
func newDavHandler(uri string, fs afero.Fs) *webdav.Handler {
	if !*davEnable {
		return nil
	}
	stripPfx := uri
	if stripPfx == "/" {
		stripPfx = ""
	}
	return &webdav.Handler{
		Prefix:     stripPfx,
		FileSystem: davFS{fs},
		LockSystem: davLS{webdav.NewMemLS()},
		// webdav maps filesystem errors to bare statuses (a failed PUT becomes
		// a 404), so without this the cause is lost
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("dav error: uri=%q method=%v: %v", r.RequestURI, r.Method, err)
			}
		},
	}
}

// davDepth bounds PROPFIND recursion. An infinite walk is a cheap
// amplification for any reader, including anonymous ones under -anon_ro, so
// refuse it the way apache does by default. x/net/webdav also reads a missing
// header as infinite; downgrade that to one level instead of failing, since
// such clients want a listing rather than an error.
func davDepth(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != methodPropfind {
		return true
	}
	d := r.Header.Get("Depth")
	if d == "" {
		r.Header.Set("Depth", "1")
		return true
	}
	if !strings.EqualFold(d, "infinity") {
		return true
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, xml.Header+`<D:error xmlns:D="DAV:"><D:propfind-finite-depth/></D:error>`)
	return false
}

func davMain(w http.ResponseWriter, r *http.Request, p wfmPrefix) {
	uName, uAccess := auth(w, r)
	if uName == "" {
		return
	}
	// challenge anonymous users instead of 403 so DAV clients can prompt
	// for credentials and retry
	if p.owner != "" && p.owner != uName {
		if uName == anonUser {
			challenge(w, r)
			return
		}
		log.Printf("dav: user %q denied access to prefix %q owner=%q", uName, p.uri, p.owner)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	if !uAccess && !davRO[r.Method] {
		if uName == anonUser {
			challenge(w, r)
			return
		}
		log.Printf("dav: read-only user %q denied %v on %q", uName, r.Method, r.RequestURI)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}
	// anonymous writes are refused above, and property/lock bodies are always
	// small, so never stream either to no purpose
	switch {
	case uName == anonUser, r.Method == methodPropfind, r.Method == methodProppatch, r.Method == methodLock:
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	}
	log.Printf("dav from=%q user=%q method=%v uri=%q agent=%v", r.RemoteAddr, uName, r.Method, r.RequestURI, r.UserAgent())
	if !davDepth(w, r) || !davDest(w, r, p.dav.Prefix) {
		return
	}
	davTimeout(r)
	p.dav.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), davMethod, r.Method)))
}
