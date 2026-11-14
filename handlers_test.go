package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// uiDo sends one request to the web ui of a read-write /files prefix rooted at
// base. A non-nil form is posted as multipart like the ui's own forms; its
// "filename" field becomes an uploaded file of that name. hdr is a list of
// extra request header names and values.
func uiDo(t *testing.T, base, target string, form map[string]string, hdr ...string) *httptest.ResponseRecorder {
	t.Helper()
	loadTemplates()
	*noPwdDbRW = true
	t.Cleanup(func() { *noPwdDbRW = false })

	r := httptest.NewRequest(http.MethodGet, target, nil)
	if form != nil {
		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		for k, v := range form {
			if k != "filename" {
				mw.WriteField(k, v)
				continue
			}
			fw, err := mw.CreateFormFile(k, v)
			if err != nil {
				t.Fatal(err)
			}
			fw.Write([]byte("uploaded"))
		}
		mw.Close()
		r = httptest.NewRequest(http.MethodPost, target, body)
		r.Header.Set("Content-Type", mw.FormDataContentType())
	}
	r.Header.Set("User-Agent", "Mozilla/5.0")
	for i := 0; i+1 < len(hdr); i += 2 {
		r.Header.Set(hdr[i], hdr[i+1])
	}
	w := httptest.NewRecorder()
	wfmMain(w, r, wfmPrefix{uri: "/files", fs: afero.NewBasePathFs(afero.NewOsFs(), base)})
	return w
}

// The dir and dst form fields must stay inside the served tree. afero's
// BasePathFs only string-prefixes, so "../data-sibling" against a base of
// /x/data used to resolve to /x/data-sibling, for reads as well as writes.
func TestUIContainment(t *testing.T) {
	top := t.TempDir()
	base, sib := filepath.Join(top, "data"), filepath.Join(top, "data-sibling")
	for _, d := range []string{base, sib} {
		if err := os.Mkdir(d, 0755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(base, "in.txt"), []byte("in"), 0644)
	os.WriteFile(filepath.Join(sib, "secret.txt"), []byte("TOP-SECRET"), 0644)

	for _, d := range []string{"../data-sibling", "/../data-sibling", "sub/../../data-sibling", "..//data-sibling/."} {
		q := "/files?dir=" + url.QueryEscape(d)
		if w := uiDo(t, base, q, nil); strings.Contains(w.Body.String(), "secret.txt") {
			t.Errorf("dir=%q listed the sibling directory", d)
		}
		if w := uiDo(t, base, q+url.QueryEscape("/secret.txt"), nil); strings.Contains(w.Body.String(), "TOP-SECRET") {
			t.Errorf("dir=%q read a file in the sibling directory", d)
		}
		for _, form := range []map[string]string{
			{"upload": "1", "dir": d, "filename": "up.txt"},
			{"fn": "mkdir", "dir": d, "file": "newdir"},
			{"fn": "mkfile", "dir": d, "file": "new.txt"},
			{"fn": "mkurl", "dir": d, "file": "new", "url": "http://example.org"},
			{"save": "1", "dir": d, "file": "saved.txt", "text": "x"},
			{"fn": "multi_delete", "dir": d, "mulf": "secret.txt"},
			{"fn": "multi_move", "dir": d, "mulf": "secret.txt", "dst": "/"},
			{"fn": "multi_move", "dir": "/", "mulf": "in.txt", "dst": d},
		} {
			uiDo(t, base, "/files", form)
		}
		uiDo(t, base, "/files/in.txt?op=mv", map[string]string{"dst": d})

		var got []string
		des, _ := os.ReadDir(sib)
		for _, de := range des {
			got = append(got, de.Name())
		}
		if !slices.Equal(got, []string{"secret.txt"}) {
			t.Fatalf("dir/dst=%q changed the sibling directory: %v", d, got)
		}
	}
}

// Login is Basic auth, which a browser attaches to any request to this host.
// A change must therefore be a POST from this origin: a link or an image on
// any page can fire a GET, a form on any page can fire a POST.
func TestWriteNeedsSameOriginPost(t *testing.T) {
	base := t.TempDir()
	keep := filepath.Join(base, "keep.txt")
	os.WriteFile(keep, []byte("x"), 0644)
	os.Mkdir(filepath.Join(base, "sub"), 0755)

	for _, q := range []string{
		"fn=multi_delete&dir=/&mulf=keep.txt",
		"fn=multi_move&dir=/&mulf=keep.txt&dst=/sub",
		"fn=mkdir&dir=/&file=newdir",
		"fn=mkfile&dir=/&file=new.txt",
		"fn=mkurl&dir=/&file=new&url=http://example.org",
		"save=1&dir=/&file=saved.txt&text=x",
	} {
		uiDo(t, base, "/files?"+q, nil)
	}
	var got []string
	des, _ := os.ReadDir(base)
	for _, de := range des {
		got = append(got, de.Name())
	}
	if !slices.Equal(got, []string{"keep.txt", "sub"}) {
		t.Errorf("GET requests changed the directory: %v", got)
	}

	// httptest requests are for host example.com
	cases := []struct {
		hdr  []string
		want bool // the change goes through
	}{
		{[]string{"Sec-Fetch-Site", "cross-site"}, false},
		{[]string{"Sec-Fetch-Site", "same-site"}, false},
		{[]string{"Origin", "https://evil.example"}, false},
		{[]string{"Origin", "null"}, false}, // a sandboxed page
		{[]string{"Sec-Fetch-Site", "same-origin"}, true},
		{[]string{"Origin", "http://example.com"}, true}, // browsers without Sec-Fetch-Site
		{nil, true}, // scripts and legacy browsers send neither header
	}
	for _, c := range cases {
		os.WriteFile(keep, []byte("x"), 0644)
		uiDo(t, base, "/files", map[string]string{"fn": "multi_delete", "dir": "/", "mulf": "keep.txt"}, c.hdr...)
		if _, err := os.Stat(keep); (err != nil) != c.want {
			t.Errorf("POST with %v: change went through = %v, want %v", c.hdr, err != nil, c.want)
		}
	}
}

// Files that can carry script must be displayed sandboxed, or a file one user
// uploads runs script as whoever opens it. Inert types are left alone so they
// keep rendering, and nothing may be content-sniffed into something else.
func TestInlineSandbox(t *testing.T) {
	base := t.TempDir()
	cases := []struct {
		name, data string
		sandbox    bool
	}{
		{"a.html", "<!DOCTYPE html><script>alert(1)</script>", true},
		{"a.svg", `<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`, true},
		{"a.xml", `<?xml version="1.0"?><html xmlns="http://www.w3.org/1999/xhtml"><script>alert(1)</script></html>`, true},
		{"a.png", "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR", false},
		{"a.pdf", "%PDF-1.4\n", false},
	}
	for _, c := range cases {
		os.WriteFile(filepath.Join(base, c.name), []byte(c.data), 0644)
		h := uiDo(t, base, "/files/"+c.name, nil).Header()
		if got := h.Get("Content-Security-Policy") == "sandbox"; got != c.sandbox {
			t.Errorf("%v (%v): sandboxed = %v, want %v", c.name, h.Get("Content-Type"), got, c.sandbox)
		}
		if h.Get("X-Content-Type-Options") != "nosniff" {
			t.Errorf("%v: no X-Content-Type-Options: nosniff", c.name)
		}
	}
}
