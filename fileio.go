package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gabriel-vasile/mimetype"
)

func fileDisp(w http.ResponseWriter, fp, disp string) {
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
	defer fi.Close()

	mtype := "application/octet-stream"
	if disp == "inline" {
		mt, err := mimetype.DetectReader(fi)
		if err == nil {
			mtype = mt.String()
		}
	}
	fi.Seek(0, 0)

	w.Header().Set("Content-Type", mtype)
	w.Header().Set("Content-Disposition", disp)
	w.Header().Set("Content-Length", fmt.Sprint(f.Size()))

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
