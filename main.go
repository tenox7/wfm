package main

import (
	"html"
	"log"
	"net/http"
	"path/filepath"
)

func wfm(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	user := auth(w, r)
	if user == "" {
		return
	}
	log.Printf("req from=%q user=%q uri=%q form=%#v", r.RemoteAddr, user, r.RequestURI, r.Form)

	dir := filepath.Clean(html.UnescapeString(r.FormValue("dir")))
	if dir == "" || dir == "." {
		dir = "/"
	}
	sort := html.EscapeString(r.FormValue("sort"))

	// toolbar buttons
	switch {
	case r.FormValue("mkd") != "":
		prompt(w, dir, "", sort, "mkdir")
		return
	case r.FormValue("mkf") != "":
		prompt(w, dir, "", sort, "mkfile")
		return
	case r.FormValue("mkb") != "":
		prompt(w, dir, "", sort, "mkurl")
		return
	case r.FormValue("upload") != "":
		f, h, err := r.FormFile("filename")
		if err != nil {
			htErr(w, "upload", err)
			return
		}
		uploadFile(w, dir, sort, h, f)
		return
	case r.FormValue("save") != "":
		saveText(w, dir, sort, filepath.Clean(html.UnescapeString(r.FormValue("fp"))), html.UnescapeString(r.FormValue("text")))
		return
	}

	// these fall through to directory listing
	if r.FormValue("home") != "" {
		dir = "/"
	}
	if r.FormValue("up") != "" {
		dir = filepath.Dir(dir)
	}

	// cancel
	if r.FormValue("cancel") != "" {
		r.Form.Set("fn", "")
	}

	// form action
	switch r.FormValue("fn") {
	case "disp":
		dispFile(w, html.UnescapeString(r.FormValue("fp")))
	case "down":
		downFile(w, html.UnescapeString(r.FormValue("fp")))
	case "edit":
		editText(w, html.UnescapeString(r.FormValue("fp")), sort)
	case "mkdir":
		mkdir(w, dir, html.UnescapeString(r.FormValue("newd")), sort)
	case "mkfile":
		mkfile(w, dir, html.UnescapeString(r.FormValue("newf")), sort)
	case "mkurl":
		mkurl(w, dir, html.UnescapeString(r.FormValue("newu")), r.FormValue("url"), sort)
	case "rename":
		renFile(w, dir, r.FormValue("oldf"), r.FormValue("newf"), sort)
	case "renp":
		prompt(w, dir, r.FormValue("oldf"), sort, "rename")
	case "logout":
		logout(w)
	default:
		listFiles(w, dir, sort, user)
	}
}
