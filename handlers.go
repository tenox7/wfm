package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type wfmRequest struct {
	fs       afero.Fs
	pfx      string
	w        http.ResponseWriter
	req      *http.Request
	userName string
	remAddr  string
	rwAccess bool
	modern   bool
	eSort    string // escaped sort order
	uDir     string // unescaped directory name
	uFbn     string // unescaped file base name
	uFilter  string // file filter pattern for dir view
}

func wfmMain(w http.ResponseWriter, r *http.Request, p wfmPrefix) {
	r.ParseMultipartForm(*formMaxMem)
	uName, uAccess := auth(w, r)
	if uName == "" {
		return
	}
	if p.owner != "" && p.owner != uName {
		log.Printf("auth: user %q denied access to home prefix %q owner=%q", uName, p.uri, p.owner)
		htErrStatus(w, r, http.StatusForbidden, "Forbidden", "you are not allowed to access this area")
		return
	}
	log.Printf("req from=%q user=%q uri=%q form=%v agent=%v", r.RemoteAddr, uName, r.RequestURI, noText(r.Form), r.UserAgent())

	if *dumpHeader {
		dump, err := httputil.DumpRequest(r, false)
		if err == nil {
			log.Printf("debug: %v", string(dump))
		}
	}

	wfm := &wfmRequest{
		userName: uName,
		rwAccess: uAccess,
		remAddr:  r.RemoteAddr,
		w:        w,
		req:      r,
		eSort:    r.FormValue("sort"),
		uFilter:  strings.TrimSpace(r.FormValue("filter")),
		modern:   isModern(r),
		fs:       p.fs,
		pfx:      p.uri,
		uFbn:     filepath.Base(r.FormValue("file")),
		uDir:     filepath.Clean(r.FormValue("dir")),
	}

	// directory can come either from form value or URI Path
	if wfm.uDir == "" || wfm.uDir == "." {
		wfm.uDir = filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, wfm.pfx))
	}
	if wfm.uDir == "" || wfm.uDir == "." {
		wfm.uDir = "/"
	}

	// button clicked
	switch {
	case r.FormValue("mkd") != "":
		wfm.prompt("mkdir", nil)
		return
	case r.FormValue("mkf") != "":
		wfm.prompt("mkfile", nil)
		return
	case r.FormValue("mkb") != "":
		wfm.prompt("mkurl", nil)
		return
	case r.FormValue("mdelp") != "":
		wfm.prompt("multi_delete", r.Form["mulf"])
		return
	case r.FormValue("mmovp") != "":
		wfm.prompt("multi_move", r.Form["mulf"])
		return
	case r.FormValue("upload") != "":
		f, h, err := r.FormFile("filename")
		if err != nil {
			wfm.htErr("upload", err)
			return
		}
		wfm.uploadFile(h, f)
		return
	case r.FormValue("save") != "":
		wfm.saveText(r.FormValue("text"), r.FormValue("crlf"))
		return
	case r.FormValue("up") != "":
		up, err := url.JoinPath(wfm.pfx, filepath.Dir(wfm.uDir))
		if err != nil {
			wfm.htErr("up path build", err)
			return
		}
		q := url.Values{}
		if wfm.eSort != "" {
			q.Set("sort", wfm.eSort)
		}
		redirect(w, wfmURL(up, q))
		return
	case r.FormValue("refresh") != "":
		re, err := url.JoinPath(wfm.pfx, wfm.uDir)
		if err != nil {
			wfm.htErr("up path build", err)
			return
		}
		q := url.Values{}
		if wfm.eSort != "" {
			q.Set("sort", wfm.eSort)
		}
		if wfm.uFilter != "" {
			q.Set("filter", wfm.uFilter)
		}
		redirect(w, wfmURL(re, q))
		return
	case r.FormValue("home") != "":
		wfm.uDir = "/"
		wfm.listFiles(filepath.Base(r.FormValue("hi")))
		return
	case r.FormValue("cancel") != "":
		wfm.listFiles(filepath.Base(r.FormValue("hi")))
		return
	}

	// single-file operation on a full path: /pfx/dir/file?op=CODE. The path is
	// the subject; GET shows a prompt/opens, POST commits (cancel handled above).
	if op := r.FormValue("op"); op != "" {
		urlPath := filepath.Clean("/" + strings.TrimPrefix(r.URL.Path, wfm.pfx))
		wfm.uDir = filepath.Dir(urlPath)
		wfm.uFbn = filepath.Base(urlPath)
		if !validName(wfm.uFbn) {
			wfm.htErr("operation", fmt.Errorf("no file specified"))
			return
		}
		post := r.Method == http.MethodPost
		switch op {
		case "dn":
			wfm.downFile()
		case "ed":
			wfm.editText()
		case "re":
			if post {
				wfm.renFile(r.FormValue("dst"))
				return
			}
			wfm.prompt("rename", nil)
		case "mv":
			if post {
				wfm.moveFiles([]string{wfm.uFbn}, r.FormValue("dst"))
				return
			}
			wfm.prompt("move", nil)
		case "rm":
			if post {
				wfm.deleteFiles([]string{wfm.uFbn})
				return
			}
			wfm.prompt("delete", nil)
		default:
			wfm.htErr("operation", fmt.Errorf("unknown op %q", op))
		}
		return
	}

	// directory-context form action: create/multi ops on the current dir, and
	// misc (logout/about). The directory arrives via the hidden dir field.
	switch r.FormValue("fn") {
	case "mkdir":
		wfm.mkdir()
	case "mkfile":
		wfm.mkfile()
	case "mkurl":
		wfm.mkurl(r.FormValue("url"))
	case "multi_delete":
		wfm.deleteFiles(r.Form["mulf"])
	case "multi_move":
		wfm.moveFiles(r.Form["mulf"], r.FormValue("dst"))
	case "logout":
		wfm.logout()
	case "about":
		wfm.about(r.UserAgent())
	default:
		wfm.dispOrDir(filepath.Base(r.FormValue("hi")))
	}
}
