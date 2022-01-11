package main

import (
	"html"
	"log"
	"net/http"
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
	log.Printf("ua=%v", r.UserAgent())
	modern := false
	if strings.HasPrefix(r.UserAgent(), "Mozilla/5") {
		modern = true
	}

	uDir := filepath.Clean(html.UnescapeString(r.FormValue("dir")))
	if uDir == "" || uDir == "." {
		uDir = "/"
	}
	eSort := html.EscapeString(r.FormValue("sort"))
	uFp := filepath.Clean(html.UnescapeString(r.FormValue("fp")))  // full file path
	uBn := filepath.Base(html.UnescapeString(r.FormValue("file"))) // base file name

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
		saveText(w, uDir, eSort, uFp, html.UnescapeString(r.FormValue("text")))
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
		renFile(w, uDir, r.FormValue("oldf"), r.FormValue("newf"), eSort)
	case "renp":
		prompt(w, uDir, r.FormValue("oldf"), eSort, "rename")
	case "delp":
		prompt(w, uDir, uBn, eSort, "delete")
	case "delete":
		log.Printf("delete %v by %v @ %v", uDir+"/"+uBn, user, r.RemoteAddr)
		delete(w, uDir, uDir+"/"+uBn, eSort)
	case "logout":
		logout(w)
	default:
		listFiles(w, uDir, eSort, user, modern)
	}
}

func favicon(w http.ResponseWriter, r *http.Request) {
	dispFavIcon(w)
}
