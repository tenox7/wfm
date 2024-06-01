package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

type wfmRequest struct {
	fs       afero.Fs
	w        http.ResponseWriter
	userName string
	remAddr  string
	rwAccess bool
	modern   bool
	eSort    string // escaped sort order
	uDir     string // unescaped directory name
	uFbn     string // unescaped file base name
}

func wfmMain(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(*formMaxMem)
	uName, uAccess := auth(w, r)
	if uName == "" {
		return
	}
	log.Printf("req from=%q user=%q uri=%q form=%v", r.RemoteAddr, uName, r.RequestURI, noText(r.Form))

	wfm := &wfmRequest{
		userName: uName,
		rwAccess: uAccess,
		remAddr:  r.RemoteAddr,
		w:        w,
		eSort:    r.FormValue("sort"),
		modern:   strings.HasPrefix(r.UserAgent(), "Mozilla/5"),
		fs:       wfmFs, // TODO(tenox): per user FS/homedir
		uFbn:     filepath.Base(r.FormValue("file")),
		uDir:     filepath.Clean(r.FormValue("dir")),
	}

	// directory can come either from form value or URI Path
	if wfm.uDir == "" || wfm.uDir == "." {
		// TODO(tenox): use url.Parse() instead
		u, err := url.PathUnescape(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wfm.uDir = filepath.Clean("/" + strings.TrimPrefix(u, wfmPfx))
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
		up, err := url.JoinPath(wfmPfx, filepath.Dir(wfm.uDir))
		if err != nil {
			htErr(w, "up path build", err)
			return
		}
		if wfm.eSort != "" {
			up += "?sort=" + wfm.eSort
		}
		redirect(w, up)
		return
	case r.FormValue("refresh") != "":
		re, err := url.JoinPath(wfmPfx, wfm.uDir)
		if err != nil {
			htErr(w, "up path build", err)
			return
		}
		if wfm.eSort != "" {
			re += "?sort=" + wfm.eSort
		}
		redirect(w, re)
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

func dispFavIcon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Write(favIcn)
}

func dispRobots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintln(w, "User-agent: *")
	if *robots {
		fmt.Fprintln(w, "Allow: /")
		return
	}
	fmt.Fprintln(w, "Disallow: /")
}
