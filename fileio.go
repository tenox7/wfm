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

func dispFile(w http.ResponseWriter, uFilePath string) {
	fp := filepath.Clean(uFilePath)
	s := strings.Split(strings.ToLower(fp), ".")
	log.Printf("Dsiposition file=%v ext=%v", fp, s[len(s)-1])
	switch s[len(s)-1] {
	case "url", "desktop", "webloc":
		gourl(w, fp)

	case "zip":
		listZip(w, fp)
	case "7z":
		list7z(w, fp)
	case "tar", "rar", "gz", "bz2", "xz", "tgz", "tbz2", "txz":
		listArchive(w, fp)
	case "iso":
		listIso(w, fp)

	default:
		dispInline(w, fp)
	}
}

func downFile(w http.ResponseWriter, uFilePath string) {
	f, err := os.Stat(uFilePath)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(uFilePath)+"\";")
	w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	w.Header().Set("Cache-Control", *cctl)
	streamFile(w, uFilePath)
}

func dispInline(w http.ResponseWriter, uFilePath string) {
	f, err := os.Stat(uFilePath)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}

	fi, err := os.Open(uFilePath)
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
	streamFile(w, uFilePath)
}

func streamFile(w http.ResponseWriter, uFilePath string) {
	fi, err := os.Open(uFilePath)
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

func uploadFile(w http.ResponseWriter, uDir, eSort string, h *multipart.FileHeader, f multipart.File) {
	defer f.Close()

	o, err := os.OpenFile(uDir+"/"+filepath.Base(h.Filename), os.O_RDWR|os.O_CREATE, 0644)
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
	log.Printf("Uploaded Dir=%v File=%v Size=%v", uDir, h.Filename, h.Size)
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}

func saveText(w http.ResponseWriter, uDir, eSort, uFilePath, eData string) {
	err := ioutil.WriteFile(uFilePath, []byte(html.UnescapeString(eData)), 0644)
	if err != nil {
		htErr(w, "unable to save text edit file: %v", err)
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", uDir, uFilePath, len(eData))
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}

func mkdir(w http.ResponseWriter, uDir, uNewd, eSort string) {
	if uNewd == "" {
		htErr(w, "mkdir", fmt.Errorf("directory name is empty"))
		return
	}
	err := os.Mkdir(uDir+"/"+filepath.Base(uNewd), 0755)
	if err != nil {
		htErr(w, "mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}

func mkfile(w http.ResponseWriter, uDir, uNewf, sort string) {
	if uNewf == "" {
		htErr(w, "mkfile", fmt.Errorf("file name is empty"))
		return
	}
	f, err := os.OpenFile(uDir+"/"+filepath.Base(uNewf), os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(w, "mkfile", err)
		return
	}
	f.Close()
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+sort)
}

func mkurl(w http.ResponseWriter, uDir, uNewu, eUrl, eSort string) {
	if uNewu == "" {
		htErr(w, "mkurl", fmt.Errorf("url file name is empty"))
		return
	}
	if !strings.HasSuffix(uNewu, ".url") {
		uNewu = uNewu + ".url"
	}
	f, err := os.OpenFile(uDir+"/"+uNewu, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(w, "mkfile", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", html.UnescapeString(eUrl))
	f.Close()
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}

func renFile(w http.ResponseWriter, uDir, eOldf, eNewf, eSort string) {
	if eOldf == "" || eNewf == "" {
		htErr(w, "rename", fmt.Errorf("filename is empty"))
		return
	}
	err := os.Rename(
		uDir+"/"+filepath.Base(html.UnescapeString(eOldf)),
		uDir+"/"+filepath.Base(html.UnescapeString(eNewf)),
	)
	if err != nil {
		htErr(w, "rename", err)
		return
	}
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}

func delete(w http.ResponseWriter, uDir, uFilePath, eSort string) {
	err := os.RemoveAll(uFilePath)
	if err != nil {
		htErr(w, "delete", err)
		return
	}
	redirect(w, *wpfx+"?dir="+html.EscapeString(uDir)+"&sort="+eSort)
}
