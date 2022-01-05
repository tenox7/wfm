package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func dispFile(w http.ResponseWriter, fp string) {
	s := strings.Split(strings.ToLower(fp), ".")
	log.Printf("Dsiposition file=%v ext=%v", fp, s[len(s)-1])
	switch s[len(s)-1] {
	case "url", "desktop", "webloc":
		gourl(w, fp)

	case "zip":
		readZip(w, fp)
	case "iso":
		readIso(w, fp)

	default:
		dispInline(w, fp)
	}
}

func downFile(w http.ResponseWriter, fp string) {
	f, err := os.Stat(fp)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(fp)+"\";")
	w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	w.Header().Set("Cache-Control", *cctl)
	streamFile(w, fp)
}

func dispInline(w http.ResponseWriter, fp string) {
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
	w.Header().Set("Cache-Control", *cctl)
	streamFile(w, fp)
}

func streamFile(w http.ResponseWriter, fp string) {
	fi, err := os.Open(fp)
	if err != nil {
		htErr(w, "Unable top open file", err)
		log.Printf("unable to read file: %v", err)
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
			log.Printf("unable to read file: %v", err)
			return
		}
		if n == 0 {
			break
		}
		wb.Write(bu[:n])
	}
	wb.Flush()
}

func uploadFile(w http.ResponseWriter, dir, sort string, h *multipart.FileHeader, f multipart.File) {
	defer f.Close()

	o, err := os.OpenFile(dir+"/"+filepath.Base(h.Filename), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		htErr(w, "unable to write file", err)
		return
	}
	defer o.Close()
	rb := bufio.NewReader(f)
	wb := bufio.NewWriter(o)
	bu := make([]byte, 1<<20)

	for {
		n, err := rb.Read(bu)
		if err != nil && err != io.EOF {
			htErr(w, "Unable to write file", err)
			return
		}
		if n == 0 {
			break
		}
		wb.Write(bu[:n])
	}
	wb.Flush()
	log.Printf("Uploaded Dir=%v File=%v Size=%v", dir, h.Filename, h.Size)
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
}

func saveText(w http.ResponseWriter, dir, sort, fp, data string) {
	err := ioutil.WriteFile(fp, []byte(data), 0644)
	if err != nil {
		htErr(w, "unable to save text edit file: %v", err)
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", dir, fp, len(data))
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
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
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", url)
	f.Close()
	redirect(w, "/?dir="+html.EscapeString(dir)+"&sort="+sort)
}
