package main

import (
	"archive/zip"
	"bufio"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"gopkg.in/ini.v1"
	"howett.net/plist"
)

func dispFile(w http.ResponseWriter, fp string) {
	// first see if there is an internal handler for a file type
	if strings.HasSuffix(strings.ToLower(fp), ".url") ||
		strings.HasSuffix(strings.ToLower(fp), ".desktop") ||
		strings.HasSuffix(strings.ToLower(fp), ".webloc") {
		gourl(w, fp)
		return
	}

	// for everything else disposition inline
	f, err := os.Stat(fp)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}

	fi, err := os.Open(fp)
	if err != nil {
		htErr(w, "Unable top open file", err)
		return
	}
	mt, err := mimetype.DetectReader(fi)
	if err != nil {
		htErr(w, "Unable to determine file type", err)
		return
	}

	fi.Close()

	w.Header().Set("Content-Type", mt.String())
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	streamFile(w, fp)
}

func streamFile(w http.ResponseWriter, fp string) {
	fi, err := os.Open(fp)
	if err != nil {
		htErr(w, "Unable top open file", err)
		return
	}
	defer fi.Close()

	rb := bufio.NewReader(fi)
	wb := bufio.NewWriter(w)
	bu := make([]byte, 1<<20)

	for {
		n, err := rb.Read(bu)
		if err != nil && err != io.EOF {
			htErr(w, "Unable to read file", err)
			return
		}
		if n == 0 {
			break
		}
		wb.Write(bu[:n])
	}
	wb.Flush()
}

func mkdir(w http.ResponseWriter, dir, newd, sort string) {
	if newd == "" {
		htErr(w, "mkdir", fmt.Errorf("directory name is empty"))
		return
	}
	err := os.Mkdir(dir+"/"+newd, 0755)
	if err != nil {
		htErr(w, "mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
}

func mkfile(w http.ResponseWriter, dir, newf, sort string) {
	if newf == "" {
		htErr(w, "mkfile", fmt.Errorf("file name is empty"))
		return
	}
	f, err := os.OpenFile(dir+"/"+newf, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(w, "mkfile", err)
		log.Printf("mkfile error: %v", err)
		return
	}
	f.Close()
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
}

func mkurl(w http.ResponseWriter, dir, newu, url, sort string) {
	if newu == "" {
		htErr(w, "mkurl", fmt.Errorf("url file name is empty"))
		return
	}
	if !strings.HasSuffix(newu, ".url") {
		newu = newu + ".url"
	}
	f, err := os.OpenFile(dir+"/"+newu, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(w, "mkfile", err)
		log.Printf("mkfile error: %v", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", url)
	f.Close()
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
}

func gourl(w http.ResponseWriter, fp string) {
	var url string
	if strings.HasSuffix(strings.ToLower(fp), ".url") {
		i, err := ini.Load(fp)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		url = i.Section("InternetShortcut").Key("URL").String()
	}

	if strings.HasSuffix(strings.ToLower(fp), ".desktop") {
		i, err := ini.Load(fp)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		url = i.Section("Desktop Entry").Key("URL").String()
	}

	if strings.HasSuffix(strings.ToLower(fp), ".webloc") {
		x, err := ioutil.ReadFile(fp)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		var p struct {
			URL string
		}
		_, err = plist.Unmarshal(x, &p)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		url = p.URL
	}

	if url == "" {
		htErr(w, "go2url", fmt.Errorf("url not found in link file"))
		return
	}
	log.Print("Redirecting to: ", url)
	redirect(w, url)
}

func archList(w http.ResponseWriter, fp string) {
	// TODO: add graphical/table view reader instead of text dump
	if strings.HasSuffix(strings.ToLower(fp), ".zip") {
		z, err := zip.OpenReader(fp)
		if err != nil {
			htErr(w, "unzip", err)
			return
		}
		defer z.Close()
		w.Header().Set("Content-Type", "text/plain")
		for _, f := range z.File {
			fmt.Fprintf(w, "%v  %v\n", f.Name, humanize.Bytes(f.UncompressedSize64))
		}
	}
}
