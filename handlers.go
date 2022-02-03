package main

import (
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

func wfm(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	user := auth(w, r)
	if user == "" {
		return
	}
	log.Printf("req from=%q user=%q uri=%q form=%#v", r.RemoteAddr, user, r.RequestURI, r.Form)
	modern := false
	if strings.HasPrefix(r.UserAgent(), "Mozilla/5") {
		modern = true
	}

	uDir := filepath.Clean(r.FormValue("dir"))
	if uDir == "" || uDir == "." {
		uDir = "/"
	}
	eSort := url.QueryEscape(r.FormValue("sort"))
	uFp := filepath.Clean(r.FormValue("fp"))
	uBn := filepath.Base(r.FormValue("file"))

	// button clicked
	switch {
	case r.FormValue("mkd") != "":
		prompt(w, uDir, "", eSort, "mkdir")
		return
	case r.FormValue("mkf") != "":
		prompt(w, uDir, "", eSort, "mkfile")
		return
	case r.FormValue("mkb") != "":
		prompt(w, uDir, "", eSort, "mkurl")
		return
	case r.FormValue("upload") != "":
		f, h, err := r.FormFile("filename")
		if err != nil {
			htErr(w, "upload", err)
			return
		}
		uploadFile(w, uDir, eSort, h, f)
		return
	case r.FormValue("save") != "":
		saveText(w, uDir, eSort, uFp, r.FormValue("text"))
		return
	case r.FormValue("home") != "":
		listFiles(w, "/", eSort, user, modern)
		return
	case r.FormValue("up") != "":
		listFiles(w, filepath.Dir(uDir), eSort, user, modern)
		return
	case r.FormValue("cancel") != "":
		listFiles(w, uDir, eSort, user, modern)
		return
	}

	// form action
	switch r.FormValue("fn") {
	case "disp":
		dispFile(w, uFp)
	case "down":
		downFile(w, uFp)
	case "edit":
		editText(w, uFp, eSort)
	case "mkdir":
		mkdir(w, uDir, uBn, eSort)
	case "mkfile":
		mkfile(w, uDir, uBn, eSort)
	case "mkurl":
		mkurl(w, uDir, uBn, r.FormValue("url"), eSort)
	case "rename":
		renFile(w, uDir, uBn, r.FormValue("dst"), eSort)
	case "renp":
		prompt(w, uDir, r.FormValue("oldf"), eSort, "rename")
	case "movp":
		prompt(w, uDir, uBn, eSort, "move")
	case "delp":
		prompt(w, uDir, uBn, eSort, "delete")
	case "move":
		log.Printf("move %v by %v @ %v", uFp, user, r.RemoteAddr)
		moveFile(w, uFp, r.FormValue("dst"), eSort)
	case "delete":
		log.Printf("delete %v by %v @ %v", uDir+"/"+uBn, user, r.RemoteAddr)
		delete(w, uDir, uDir+"/"+uBn, eSort)
	case "logout":
		logout(w)
	case "about":
		about(w, uDir, eSort, r.UserAgent())
	default:
		listFiles(w, uDir, eSort, user, modern)
	}
}

func favicon(w http.ResponseWriter, r *http.Request) {
	dispFavIcon(w)
}
