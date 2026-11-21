package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

func TestParseHtaccess(t *testing.T) {
	cases := []struct {
		src             string
		access, indexes htTri
		err             bool
	}{
		{src: "Options -Indexes", indexes: htNo},
		{src: "options -indexes", indexes: htNo},
		{src: "Options +FollowSymLinks"},
		{src: "Options FollowSymLinks", indexes: htNo},
		{src: "Options None", indexes: htNo},
		{src: "Options Indexes FollowSymLinks", indexes: htYes},
		{src: "Options All -Indexes", indexes: htNo},
		{src: "Options +Indexes", indexes: htYes},
		{src: "Options +Indexes\nOptions -Indexes", indexes: htNo},
		{src: "Options -Indexes\nOptions +Indexes", indexes: htNo}, // deny wins within a file
		{src: "\uFEFFOptions -Indexes", indexes: htNo},
		{src: "Options \\\n  -Indexes", indexes: htNo},
		{src: "# Options -Indexes\n  # Require all denied"},
		{src: "Require all denied", access: htNo},
		{src: "Require all granted", access: htYes},
		{src: "Require ALL Denied", access: htNo},
		{src: "Order allow,deny\nAllow from all", access: htYes},
		{src: "Order deny,allow\nDeny from all", access: htNo},
		{src: "Require ip 10.0.0.0/8", access: htNo},
		{src: "AuthType Basic\nRequire valid-user", access: htNo},
		{src: "Deny from 192.0.2.1", access: htNo},
		{src: "Allow from 192.0.2.1", access: htNo},
		{src: "Require all granted\n<IfModule mod_x.c>\nRequire all denied\n</IfModule>", access: htNo},
		{src: "<IfModule mod_authz_core.c>\n  Require all denied\n</IfModule>\n<IfModule !mod_authz_core.c>\n  Order allow,deny\n  Deny from all\n</IfModule>", access: htNo},
		{src: "<Limit GET POST>\nRequire all denied\n</Limit>", access: htNo},
		{src: "<Limit PUT DELETE>\nRequire all denied\n</Limit>"},
		{src: "<LimitExcept GET HEAD>\nRequire all denied\n</LimitExcept>"},
		{src: "<LimitExcept POST>\nRequire all denied\n</LimitExcept>", access: htNo},
		{src: "<RequireNone>\nRequire all granted\n</RequireNone>", access: htNo},
		{src: "<Files a.txt>\nOptions -Indexes\nRequire all denied\n</Files>"},
		{src: "RewriteEngine On\nRewriteRule ^ - [F]\nErrorDocument 404 /404.html"},
		{src: "<Files a.txt>\nRequire all denied", err: true},
		{src: "</Files>", err: true},
		{src: "<IfModule x>\n</Files>", err: true},
		{src: "<Files>\n</Files>", err: true},
		{src: "<Files a b>\n</Files>", err: true},
		{src: "<Files a.txt\n</Files>", err: true},
		{src: "<Files [>\n</Files>", err: true},
		{src: "<FilesMatch (?<=x)>\n</FilesMatch>", err: true},
		{src: "<Files a>\n<Files b>\n</Files>\n</Files>", err: true},
	}
	for _, c := range cases {
		h, _, err := parseHtaccess(c.src)
		if (err != nil) != c.err {
			t.Errorf("parseHtaccess(%q) error = %v, want error %v", c.src, err, c.err)
			continue
		}
		if err == nil && (h.access != c.access || h.indexes != c.indexes) {
			t.Errorf("parseHtaccess(%q) = access %v indexes %v, want %v %v", c.src, h.access, h.indexes, c.access, c.indexes)
		}
	}
}

func TestHtFileDenied(t *testing.T) {
	c, _, err := parseHtaccess(`
<Files "*.cfg">
	Require all denied
</Files>
<Files "public.cfg">
	Require all granted
</Files>
<FilesMatch "\.(bak|log)$">
	Order allow,deny
	Deny from all
</FilesMatch>
<Files ~ "^draft-">
	Require all denied
</Files>
<Files 'my file.txt'>
	<IfModule mod_authz_core.c>
		Require all denied
	</IfModule>
</Files>
<Files "notes.txt">
	Options -Indexes
</Files>
`)
	if err != nil {
		t.Fatal(err)
	}
	h := htDir{files: c.files}
	for name, want := range map[string]bool{
		"app.cfg":        true,
		"public.cfg":     false, // the later section wins
		"PUBLIC.CFG":     false,
		"a.bak":          true,
		"X.LOG":          true, // case-insensitive filesystems open x.log
		"a.bak.txt":      false,
		"draft-1.md":     true,
		"my file.txt":    true,
		"MY FILE.TXT":    true,
		"my file.txt.gz": false,
		"notes.txt":      false,
		".htaccess":      true,
		".HTpasswd":      true,
		"readme.txt":     false,
	} {
		if got := h.fileDenied(name); got != want {
			t.Errorf("fileDenied(%q) = %v, want %v", name, got, want)
		}
	}
	if !(htDir{deny: true}).fileDenied("readme.txt") {
		t.Error("a file in a denied directory is not denied")
	}
	if (htDir{deny: true, files: c.files}).fileDenied("public.cfg") {
		t.Error("a <Files> grant does not open a file in a denied directory")
	}
}

// htTree is a served tree with every kind of rule in it.
func htTree(t *testing.T) afero.Fs {
	t.Helper()
	base := t.TempDir()
	for n, data := range map[string]string{
		".htaccess":            "<Files \"*.bak\">\nRequire all denied\n</Files>\n",
		"a.txt":                "a",
		"a.bak":                "SECRET",
		"noidx/.htaccess":      "Options -Indexes",
		"noidx/f.txt":          "f",
		"noidx/sub/g.txt":      "g",
		"noidx/open/.htaccess": "Options +Indexes",
		"noidx/open/h.txt":     "h",
		"priv/.htaccess":       "Require all denied",
		"priv/s.txt":           "SECRET",
		"priv/pub/.htaccess":   "Require all granted",
		"priv/pub/p.txt":       "p",
		"priv/pub/p.bak":       "SECRET",
		"broken/.htaccess":     "<Files x>",
		"broken/ok/.htaccess":  "Require all granted",
		"broken/ok/c.txt":      "SECRET",
		"ipdeny/.htaccess":     "Require ip 10.0.0.0/8",
		"ipdeny/d.txt":         "SECRET",
		"hidx/.htaccess":       "<Files index.html>\nRequire all denied\n</Files>",
		"hidx/index.html":      "SECRET",
		"hidx/index.htm":       "htm",
		"loop/l.txt":           "SECRET",
		"isdir/i.txt":          "SECRET",
		"big/.htaccess":        strings.Repeat("# comment\n", htMax/10+1),
		"big/b.txt":            "SECRET",
	} {
		p := filepath.Join(base, n)
		os.MkdirAll(filepath.Dir(p), 0755)
		if err := os.WriteFile(p, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// a .htaccess that can not be read denies, like a broken one
	os.Symlink(".htaccess", filepath.Join(base, "loop/.htaccess"))
	os.Mkdir(filepath.Join(base, "isdir/.htaccess"), 0755)
	return afero.NewBasePathFs(afero.NewOsFs(), base)
}

func TestHtaccessWebserver(t *testing.T) {
	loadTemplates()
	p := wfmPrefix{uri: "/", fs: htTree(t), web: true, autoIndex: true, htaccess: true}
	get := func(u string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		webMain(w, httptest.NewRequest(http.MethodGet, u, nil), p)
		return w
	}
	for u, want := range map[string]int{
		"/a.txt":           200,
		"/a.bak":           403,
		"/A.BAK":           403,
		"/.htaccess":       403,
		"/.HTACCESS":       403,
		"/nope.txt":        404,
		"/noidx/":          403,
		"/noidx/f.txt":     200,
		"/noidx/sub/":      403, // inherited
		"/noidx/sub/g.txt": 200,
		"/noidx/open/":     200,
		"/priv":            403, // not redirected to priv/ first
		"/priv/":           403,
		"/priv/s.txt":      403,
		"/priv/nope.txt":   403, // does not tell what exists
		"/priv/pub/":       200, // a nearer .htaccess wins
		"/priv/pub/p.txt":  200,
		"/priv/pub/p.bak":  403, // <Files> from the root still applies
		"/broken/":         403,
		"/broken/ok/c.txt": 403, // a broken file can not be overridden
		"/ipdeny/d.txt":    403,
		"/loop/l.txt":      403,
		"/isdir/i.txt":     403,
		"/big/b.txt":       403, // over htMax
	} {
		if got := get(u).Code; got != want {
			t.Errorf("GET %v = %v, want %v", u, got, want)
		}
	}

	body := get("/").Body.String()
	for _, n := range []string{"a.txt", "noidx/"} {
		if !strings.Contains(body, ">"+n+"<") {
			t.Errorf("listing of / misses %v", n)
		}
	}
	for _, n := range []string{"a.bak", "priv", "broken", "ipdeny", "loop", "isdir", "big", ".htaccess"} {
		if strings.Contains(body, n) {
			t.Errorf("listing of / shows %v", n)
		}
	}
	if b := get("/priv/pub/").Body.String(); !strings.Contains(b, "p.txt") || strings.Contains(b, "p.bak") {
		t.Errorf("listing of /priv/pub/ is wrong:\n%v", b)
	}

	p.index = true
	if b := get("/hidx/").Body.String(); b != "htm" {
		t.Errorf("/hidx/ served %q, want the index.htm after the denied index.html", b)
	}

	p.htaccess = false
	for _, u := range []string{"/a.bak", "/.htaccess", "/priv/s.txt", "/noidx/f.txt", "/broken/ok/c.txt"} {
		if got := get(u).Code; got != 200 {
			t.Errorf("-htaccess_web=false: GET %v = %v, want 200", u, got)
		}
	}
	p.index = false
	if got := get("/noidx/").Code; got != 200 {
		t.Errorf("-htaccess_web=false: GET /noidx/ = %v, want 200", got)
	}
}

// .htaccess rules bind users without rw access on file manager prefixes; rw
// users manage the tree, .htaccess files included.
func TestHtaccessUI(t *testing.T) {
	loadTemplates()
	p := wfmPrefix{uri: "/files", fs: htTree(t), htaccess: true}
	t.Cleanup(func() { *noPwdDbRW = false })
	get := func(u string, rw bool, ua string) *httptest.ResponseRecorder {
		*noPwdDbRW = rw
		r := httptest.NewRequest(http.MethodGet, u, nil)
		r.Header.Set("User-Agent", ua)
		w := httptest.NewRecorder()
		wfmMain(w, r, p)
		return w
	}
	for _, ua := range []string{"Mozilla/5.0", "Mozilla/4.7", "curl/8.0"} {
		for u, want := range map[string]int{
			"/files/a.txt":               200,
			"/files/a.bak":               403,
			"/files/a.bak?op=dn":         403,
			"/files/a.bak?op=ed":         403,
			"/files/.htaccess?op=ed":     403,
			"/files/priv/":               403,
			"/files/priv/s.txt?op=dn":    403,
			"/files/priv/nope.txt":       403,
			"/files/noidx/":              403,
			"/files/noidx/f.txt":         200,
			"/files/noidx/open/":         200,
			"/files/?home=1":             200,
			"/files/noidx/?cancel=1":     403,
			"/files/?dir=/priv&cancel=1": 403,
		} {
			w := get(u, false, ua)
			if w.Code != want {
				t.Errorf("ro %v GET %v = %v, want %v", ua, u, w.Code, want)
			}
			if strings.Contains(w.Body.String(), "SECRET") {
				t.Errorf("ro %v GET %v leaked a denied file", ua, u)
			}
		}
		body := get("/files/", false, ua).Body.String()
		if !strings.Contains(body, "a.txt") || strings.Contains(body, "a.bak") || strings.Contains(body, "priv") {
			t.Errorf("ro %v listing is wrong:\n%v", ua, body)
		}
	}

	for _, u := range []string{"/files/a.bak", "/files/priv/s.txt", "/files/priv/s.txt?op=dn", "/files/.htaccess?op=ed"} {
		if w := get(u, true, "Mozilla/5.0"); w.Code != 200 {
			t.Errorf("rw GET %v = %v, want 200", u, w.Code)
		}
	}
	body := get("/files/", true, "Mozilla/5.0").Body.String()
	if !strings.Contains(body, "a.bak") || !strings.Contains(body, "priv") {
		t.Errorf("rw listing hides files:\n%v", body)
	}

	p.htaccess = false
	if w := get("/files/priv/s.txt", false, "Mozilla/5.0"); w.Code != 200 {
		t.Errorf("-htaccess=false: ro GET /files/priv/s.txt = %v, want 200", w.Code)
	}
}

func TestHtaccessDav(t *testing.T) {
	fs := htTree(t)
	p := wfmPrefix{uri: "/files", fs: fs, dav: newDavHandler("/files", fs), htaccess: true}
	t.Cleanup(func() { *noPwdDbRW = false })
	propfind := func(u, depth string, rw bool) *httptest.ResponseRecorder {
		*noPwdDbRW = rw
		r := httptest.NewRequest(methodPropfind, u, nil)
		r.Header.Set("Depth", depth)
		w := httptest.NewRecorder()
		davMain(w, r, p)
		return w
	}
	for _, c := range []struct {
		u, depth string
		want     int
	}{
		{"/files/", "1", http.StatusMultiStatus},
		{"/files/priv/", "0", http.StatusForbidden},
		{"/files/priv/s.txt", "0", http.StatusForbidden},
		{"/files/a.bak", "0", http.StatusForbidden},
		{"/files/noidx/", "1", http.StatusForbidden},
		{"/files/noidx/", "0", http.StatusMultiStatus},
		{"/files/priv/pub/", "1", http.StatusMultiStatus},
	} {
		if got := propfind(c.u, c.depth, false).Code; got != c.want {
			t.Errorf("ro PROPFIND %v depth %v = %v, want %v", c.u, c.depth, got, c.want)
		}
	}
	body := propfind("/files/", "1", false).Body.String()
	if !strings.Contains(body, "/files/a.txt") || strings.Contains(body, "a.bak") || strings.Contains(body, "/files/priv/") {
		t.Errorf("ro PROPFIND listing is wrong:\n%v", body)
	}
	body = propfind("/files/", "1", true).Body.String()
	if !strings.Contains(body, "/files/a.bak") || !strings.Contains(body, "/files/priv/") {
		t.Errorf("rw PROPFIND listing hides files:\n%v", body)
	}
}
