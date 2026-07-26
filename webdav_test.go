package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/net/webdav"
)

// davClean confines COPY/MOVE Destination headers, which x/net/webdav passes
// through uncleaned; afero's BasePathFs only string-prefixes, so "/../sibling"
// would otherwise reach any directory sharing the base name prefix.
func TestDavClean(t *testing.T) {
	cases := map[string]string{
		"":                     "/",
		"/":                    "/",
		"/f.txt":               "/f.txt",
		"/../alice2/f.txt":     "/alice2/f.txt",
		"/../../../etc/passwd": "/etc/passwd",
		"/a/../../b":           "/b",
		"/a/./b//c":            "/a/b/c",
		"..":                   "/",
		"../../x":              "/x",
	}
	for in, want := range cases {
		if got := davClean(in); got != want {
			t.Errorf("davClean(%q) = %q, want %q", in, got, want)
		}
	}
}

// A COPY/MOVE destination that aliases the source, or sits above or below it,
// must be refused: webdav removes the destination before writing it, so these
// recursively delete the source (reporting success) or recurse until the
// filesystem refuses the path.
func TestDavDest(t *testing.T) {
	cases := []struct {
		pfx, src, dst string
		want          int // 0 = allowed
	}{
		{"/dav", "/dav/d", "/dav/d/", http.StatusForbidden},         // trailing slash alias
		{"/dav", "/dav/d", "/dav/./d", http.StatusForbidden},        // dot alias
		{"/dav", "/dav/d", "/dav//d", http.StatusForbidden},         // empty segment alias
		{"/dav", "/dav/d", "/dav/x/../d", http.StatusForbidden},     // traversal alias
		{"/dav", "/dav/d", "/dav/d", http.StatusForbidden},          // identical
		{"/dav", "/dav/d/f.txt", "/dav/d/..", http.StatusForbidden}, // ancestor: wipes source
		{"/dav", "/dav/d/f.txt", "/dav/d", http.StatusForbidden},    // ancestor
		{"/dav", "/dav/d/f.txt", "/dav/", http.StatusForbidden},     // share root
		{"/dav", "/dav/d", "/dav/d/sub", http.StatusForbidden},      // descendant: recurses
		{"/dav", "/dav/d", "/dav/d2", 0},                            // sibling name prefix
		{"/dav", "/dav/d/f.txt", "/dav/d/g.txt", 0},
		{"/dav", "/dav/d", "/dav/e/f", 0},
		{"/dav/alice", "/dav/alice/f.txt", "/dav/alice2/f.txt", http.StatusBadGateway}, // other mount
		{"", "/d", "/d/", http.StatusForbidden},                                        // root mount
		{"", "/d/f.txt", "/e/f.txt", 0},
	}
	for _, c := range cases {
		for _, m := range []string{methodCopy, methodMove} {
			r := httptest.NewRequest(m, c.src, nil)
			r.Header.Set("Destination", "http://example.org"+c.dst)
			w := httptest.NewRecorder()
			if got := davDest(w, r, c.pfx); got != (c.want == 0) {
				t.Errorf("davDest(%v %q -> %q, pfx %q) = %v, want %v", m, c.src, c.dst, c.pfx, got, c.want == 0)
			}
			if c.want != 0 && w.Code != c.want {
				t.Errorf("davDest(%v %q -> %q) status = %v, want %v", m, c.src, c.dst, w.Code, c.want)
			}
		}
	}
	// an allowed destination is handed on in canonical form
	r := httptest.NewRequest(methodCopy, "/dav/d/f.txt", nil)
	r.Header.Set("Destination", "http://example.org/dav/./e//g.txt")
	if !davDest(httptest.NewRecorder(), r, "/dav") {
		t.Fatal("davDest refused a legitimate destination")
	}
	if got, want := r.Header.Get("Destination"), "http://example.org/dav/e/g.txt"; got != want {
		t.Errorf("canonical Destination = %q, want %q", got, want)
	}
	// a method without a Destination is untouched
	if !davDest(httptest.NewRecorder(), httptest.NewRequest(http.MethodPut, "/dav/f.txt", nil), "/dav") {
		t.Error("davDest refused a PUT")
	}
}

// davFS must confine names and refuse to remove or rename the served root,
// whatever webdav passes down.
func TestDavFS(t *testing.T) {
	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, "/a/f.txt", []byte("x"), 0644)
	d, ctx := davFS{fs}, context.Background()

	if _, err := d.Stat(ctx, "/../a/f.txt"); err != nil {
		t.Errorf("traversal not confined into the share: %v", err)
	}
	for _, root := range []string{"/", "", "..", "/..", "/a/.."} {
		if err := d.RemoveAll(ctx, root); err == nil {
			t.Errorf("RemoveAll(%q) removed the served root", root)
		}
		if err := d.Rename(ctx, "/a/f.txt", root); err == nil {
			t.Errorf("Rename onto %q replaced the served root", root)
		}
	}
	if err := d.Rename(ctx, "/a/f.txt", "/../g.txt"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if ok, _ := afero.Exists(fs, "/g.txt"); !ok {
		t.Error("traversal rename did not land inside the share")
	}
	if err := d.RemoveAll(ctx, "/g.txt"); err != nil {
		t.Errorf("RemoveAll of a regular file: %v", err)
	}
}

// the served root must survive a destination that cleans onto it: webdav
// removes an overwrite destination before writing, so allowing it would delete
// the whole share.
func TestDavMutateRoot(t *testing.T) {
	for _, n := range []string{"/", "", "..", "/..", "/../", "/sub/..", "/a/b/../.."} {
		if _, err := davMutate(n); err == nil {
			t.Errorf("davMutate(%q) allowed removing the served root", n)
		}
	}
	for _, n := range []string{"/f.txt", "/sub/f.txt", "/../f.txt", "/sub/../f.txt"} {
		if _, err := davMutate(n); err != nil {
			t.Errorf("davMutate(%q) = %v, want allowed", n, err)
		}
	}
}

func TestDavTimeout(t *testing.T) {
	max := fmt.Sprintf("Second-%d", davMaxLock/time.Second)
	cases := []struct{ method, in, want string }{
		{methodLock, "", max},
		{methodLock, "Infinite", max},
		{methodLock, "Second-99999", max},
		{methodLock, "Infinite, Second-60", max},
		{methodLock, "Second-60", "Second-60"},
		{methodLock, "Second-60, Infinite", "Second-60, Infinite"}, // webdav honors the first value only
		{methodLock, "bogus", "bogus"},                             // left for webdav to reject
		{http.MethodPut, "", ""},                                   // only LOCK carries a Timeout
	}
	for _, c := range cases {
		r := httptest.NewRequest(c.method, "/f.txt", nil)
		if c.in != "" {
			r.Header.Set("Timeout", c.in)
		}
		davTimeout(r)
		if got := r.Header.Get("Timeout"); got != c.want {
			t.Errorf("davTimeout(%v %q) = %q, want %q", c.method, c.in, got, c.want)
		}
	}
}

// an infinite lock (no Timeout header) must still expire: x/net's memLS never
// collects negative durations, so one LOCK would block writers until restart.
func TestDavLockExpires(t *testing.T) {
	ls := davLS{webdav.NewMemLS()}
	now := time.Unix(1e9, 0)
	ld := webdav.LockDetails{Root: "/f.txt", Duration: -1, ZeroDepth: true}
	if _, err := ls.Create(now, ld); err != nil {
		t.Fatalf("Create: %v", err)
	}
	// a conflicting locker is how the handler detects a live lock (423)
	if _, err := ls.Create(now.Add(davMaxLock-time.Second), ld); err != webdav.ErrLocked {
		t.Errorf("Create before expiry = %v, want ErrLocked", err)
	}
	if _, err := ls.Create(now.Add(davMaxLock+time.Second), ld); err != nil {
		t.Errorf("Create after expiry = %v, want the lock to have expired", err)
	}
}

func TestDavLockDur(t *testing.T) {
	cases := map[time.Duration]time.Duration{
		-1:             davMaxLock,
		davMaxLock * 2: davMaxLock,
		time.Minute:    time.Minute,
		0:              0,
		davMaxLock:     davMaxLock,
	}
	for in, want := range cases {
		if got := davLockDur(in); got != want {
			t.Errorf("davLockDur(%v) = %v, want %v", in, got, want)
		}
	}
}
