package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func deniedPfx(pfx string) bool {
	cPfx := filepath.Clean(pfx)
	for _, p := range denyPfxs {
		if strings.HasPrefix(cPfx, p) {
			return true
		}
	}
	return false
}

func (r wfmRequest) dispFile() {
	fp := r.uFp // TODO(tenox): uDir + uBn
	// TODO(tenox): deniedpfx should be in handlers???
	if deniedPfx(fp) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	s := strings.Split(fp, ".")
	log.Printf("Dsiposition file=%v ext=%v", fp, s[len(s)-1])
	switch strings.ToLower(s[len(s)-1]) {
	case "url", "desktop", "webloc":
		gourl(r.w, fp)

	case "zip":
		listZip(r.w, fp)
	case "7z":
		list7z(r.w, fp)
	case "tar", "rar", "gz", "bz2", "xz", "tgz", "tbz2", "txz":
		listArchive(r.w, fp)
	case "iso":
		listIso(r.w, fp)

	default:
		dispInline(r.w, fp)
	}
}

func (r wfmRequest) downFile() {
	fp := r.uFp // TODO(tenox): uDir + uBn
	if deniedPfx(fp) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	f, err := os.Stat(fp)
	if err != nil {
		htErr(r.w, "Unable to get file attributes", err)
		return
	}
	r.w.Header().Set("Content-Type", "application/octet-stream")
	r.w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(fp)+"\";")
	r.w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	r.w.Header().Set("Cache-Control", *cacheCtl)
	streamFile(r.w, fp)
}

func dispInline(w http.ResponseWriter, uFilePath string) {
	if deniedPfx(uFilePath) {
		htErr(w, "access", fmt.Errorf("forbidden"))
		return
	}
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
	w.Header().Set("Cache-Control", *cacheCtl)
	streamFile(w, uFilePath)
}

func streamFile(w http.ResponseWriter, uFilePath string) {
	if deniedPfx(uFilePath) {
		htErr(w, "access", fmt.Errorf("forbidden"))
		return
	}
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

func (r wfmRequest) uploadFile(h *multipart.FileHeader, f multipart.File) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	defer f.Close()

	o, err := os.OpenFile(r.uDir+"/"+filepath.Base(h.Filename), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "unable to write file", err)
		return
	}
	defer o.Close()
	rb := bufio.NewReader(f)
	wb := bufio.NewWriter(o)
	bu := make([]byte, 1<<20)

	for {
		n, err := rb.Read(bu)
		if err != nil && err != io.EOF {
			htErr(r.w, "Unable to write file", err)
			return
		}
		if n == 0 {
			break
		}
		wb.Write(bu[:n])
	}
	wb.Flush()
	log.Printf("Uploaded Dir=%v File=%v Size=%v", r.uDir, h.Filename, h.Size)
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(h.Filename))
}

func (r wfmRequest) saveText(uData string) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	if uData == "" {
		htErr(r.w, "text save", fmt.Errorf("zero lenght data"))
		return
	}
	fp := r.uFp // TODO(tenox): uDir + uBn
	tmpName := fp + ".tmp"
	err := ioutil.WriteFile(tmpName, []byte(uData), 0644)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	f, err := os.Stat(tmpName)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	if f.Size() != int64(len(uData)) {
		htErr(r.w, "text save", fmt.Errorf("temp file size != input size"))
		return
	}
	err = os.Rename(tmpName, r.uFp)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", r.uDir, r.uFp, len(uData))
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(filepath.Base(r.uFp)))
}

func (r wfmRequest) mkdir() {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}

	if r.uBn == "" {
		htErr(r.w, "mkdir", fmt.Errorf("directory name is empty"))
		return
	}
	uB := filepath.Base(r.uBn)
	err := os.Mkdir(r.uDir+"/"+uB, 0755)
	if err != nil {
		htErr(r.w, "mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(uB))
}

func (r wfmRequest) mkfile() {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}

	if r.uBn == "" {
		htErr(r.w, "mkfile", fmt.Errorf("file name is empty"))
		return
	}
	fB := filepath.Base(r.uBn)
	f, err := os.OpenFile(r.uDir+"/"+fB, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	f.Close()
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(fB))
}

func (r wfmRequest) mkurl(eUrl string) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	if r.uBn == "" {
		htErr(r.w, "mkurl", fmt.Errorf("url file name is empty"))
		return
	}
	fB := filepath.Base(r.uBn)
	if !strings.HasSuffix(fB, ".url") {
		fB = fB + ".url"
	}
	f, err := os.OpenFile(r.uDir+"/"+fB, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", eUrl)
	f.Close()
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(fB))
}

func (r wfmRequest) renFile(uNewf string) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}

	if r.uBn == "" || uNewf == "" {
		htErr(r.w, "rename", fmt.Errorf("filename is empty"))
		return
	}
	fB := filepath.Base(uNewf)
	err := os.Rename(
		r.uDir+"/"+r.uBn,
		r.uDir+"/"+fB,
	)
	if err != nil {
		htErr(r.w, "rename", err)
		return
	}
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.QueryEscape(fB))
}

func (r wfmRequest) moveFiles(uFilePaths []string, uDst string) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	uDst = filepath.Clean(uDst)
	if deniedPfx(r.uDir) || deniedPfx(uDst) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	log.Printf("move dir=%v files=%+v dst=%v user=%v@%v", r.uDir, uFilePaths, uDst, r.user, r.remAddr)

	lF := ""
	for _, f := range uFilePaths {
		fb := filepath.Base(f)
		err := os.Rename(
			r.uDir+"/"+fb,
			filepath.Clean(uDst+"/"+fb),
		)
		if err != nil {
			htErr(r.w, "move", err)
			return
		}
		lF = fb
	}
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(uDst)+"&sort="+r.eSort+"&hi="+url.QueryEscape(lF))
}

func (r wfmRequest) deleteFiles(uFilePaths []string) {
	if !r.rw {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if deniedPfx(r.uDir) {
		htErr(r.w, "access", fmt.Errorf("forbidden"))
		return
	}
	log.Printf("delete dir=%v files=%+v user=%v@%v", r.uDir, uFilePaths, r.user, r.remAddr)

	for _, f := range uFilePaths {
		err := os.RemoveAll(r.uDir + "/" + filepath.Base(f))
		if err != nil {
			htErr(r.w, "delete", err)
			return
		}
	}
	redirect(r.w, *wfmPfx+"?dir="+url.QueryEscape(r.uDir)+"&sort="+r.eSort)
}
