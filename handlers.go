package main

import (
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
	userName string
	remAddr  string
	rwAccess bool
	modern   bool
	eSort    string // escaped sort order
	uDir     string // unescaped directory name
	uFbn     string // unescaped file base name
}

func wfmMain(w http.ResponseWriter, r *http.Request, p wfmPrefix) {
	r.ParseMultipartForm(*formMaxMem)
	uName, uAccess := auth(w, r)
	if uName == "" {
		return
	}
	if p.owner != "" && p.owner != uName {
		log.Printf("auth: user %q denied access to home prefix %q owner=%q", uName, p.uri, p.owner)
		http.Error(w, "Forbidden", http.StatusForbidden)
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
		eSort:    r.FormValue("sort"),
		modern: func() bool {
			return strings.HasPrefix(r.UserAgent(), "Mozilla/5") && r.Header.Get("Accept-Charset") == ""
		}(),
		fs:   p.fs,
		pfx:  p.uri,
		uFbn: filepath.Base(r.FormValue("file")),
		uDir: filepath.Clean(r.FormValue("dir")),
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
			htErr(w, "upload", err)
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
			htErr(w, "up path build", err)
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
			htErr(w, "up path build", err)
			return
		}
		q := url.Values{}
		if wfm.eSort != "" {
			q.Set("sort", wfm.eSort)
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

	// form action submitted
	switch r.FormValue("fn") {
	case "disp":
		wfm.dispFile()
	case "down":
		wfm.downFile()
	case "edit":
		wfm.editText()
	case "mkdir":
		wfm.mkdir()
	case "mkfile":
		wfm.mkfile()
	case "mkurl":
		wfm.mkurl(r.FormValue("url"))
	case "rename":
		wfm.renFile(r.FormValue("dst"))
	case "renp":
		wfm.prompt("rename", nil)
	case "movp":
		wfm.prompt("move", nil)
	case "delp":
		wfm.prompt("delete", nil)
	case "move":
		wfm.moveFiles([]string{wfm.uFbn}, r.FormValue("dst"))
	case "delete":
		wfm.deleteFiles([]string{wfm.uFbn})
	case "multi_delete":
		wfm.deleteFiles(r.Form["mulf"])
	case "multi_move":
		wfm.moveFiles(r.Form["mulf"], r.FormValue("dst"))
	case "logout":
		logout(w)
	case "about":
		wfm.about(r.UserAgent())
	default:
		wfm.dispOrDir(filepath.Base(r.FormValue("hi")))
	}
}
