package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type wfmRequest struct {
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
	wfm := new(wfmRequest)
	r.ParseMultipartForm(10 << 20)
	wfm.userName, wfm.rwAccess = auth(w, r)
	if wfm.userName == "" {
		return
	}
	go log.Printf("req from=%q user=%q uri=%q form=%v", r.RemoteAddr, wfm.userName, r.RequestURI, noText(r.Form))

	wfm.w = w
	wfm.remAddr = r.RemoteAddr
	wfm.eSort = url.QueryEscape(r.FormValue("sort"))
	if strings.HasPrefix(r.UserAgent(), "Mozilla/5") {
		wfm.modern = true
	}
	wfm.uFbn = filepath.Base(r.FormValue("file"))
	wfm.uDir = filepath.Clean(r.FormValue("dir"))
	// directory can come from form value or URI
	if wfm.uDir == "" || wfm.uDir == "." {
		u, _ := url.QueryUnescape(r.RequestURI)
		wfm.uDir = filepath.Clean("/" + strings.TrimPrefix(u, *wfmPfx))
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
		wfm.saveText(r.FormValue("text"))
		return
	case r.FormValue("home") != "":
		wfm.uDir = "/"
		wfm.listFiles(filepath.Base(r.FormValue("hi")))
		return
	case r.FormValue("up") != "":
		wfm.uDir = filepath.Dir(wfm.uDir)
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
		wfm.listFiles(filepath.Base(r.FormValue("hi")))
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
