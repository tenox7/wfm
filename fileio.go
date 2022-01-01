package main

import (
	"bufio"
	"fmt"
	"io"
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
