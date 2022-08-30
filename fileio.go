package main

import (
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
	"github.com/juju/ratelimit"
)

var (
	rlBu *ratelimit.Bucket
)

func (r *wfmRequest) dispFile() {
	fp := r.uDir + "/" + r.uFbn
	s := strings.Split(fp, ".")
	log.Printf("Dsiposition file=%v ext=%v", fp, s[len(s)-1])
	ext := strings.ToLower(s[len(s)-1])
	// inexpensive file handlers
	switch ext {
	case "url", "desktop", "webloc":
		gourl(r.w, fp)
		return
	}

	if !*listArc {
		dispInline(r.w, fp)
		return
	}

	// expensive file handlers
	switch ext {
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

func (r *wfmRequest) downFile() {
	fp := r.uDir + "/" + r.uFbn
	f, err := os.Stat(fp)
	if err != nil {
		htErr(r.w, "Unable to get file attributes", err)
		return
	}
	r.w.Header().Set("Content-Type", "application/octet-stream")
	r.w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(r.uFbn)+"\";")
	r.w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	r.w.Header().Set("Cache-Control", *cacheCtl)
	streamFile(r.w, fp)
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
	w.Header().Set("Content-Disposition", "inline; filename=\""+url.PathEscape(filepath.Base(uFilePath))+"\";")
	w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	w.Header().Set("Cache-Control", *cacheCtl)
	streamFile(w, uFilePath)
}

func streamFile(w http.ResponseWriter, uFilePath string) {
	fi, err := os.Open(uFilePath)
	if err != nil {
		htErr(w, "Unable top open file", err)
		return
	}
	defer fi.Close()

	var r io.Reader = fi
	if *rateLim != 0 {
		r = ratelimit.Reader(fi, rlBu)
	}

	_, err = io.Copy(w, r)
	if err != nil {
		htErr(w, "streaming file", err)
	}
}

func (r *wfmRequest) uploadFile(h *multipart.FileHeader, f multipart.File) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	defer f.Close()

	fi, err := os.OpenFile(r.uDir+"/"+filepath.Base(h.Filename), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "unable to write file", err)
		return
	}
	defer fi.Close()

	var w io.Writer = fi
	if *rateLim != 0 {
		w = ratelimit.Writer(fi, rlBu)
	}

	oSize, err := io.Copy(w, f)
	if err != nil {
		htErr(r.w, "uploading file", err)
		return
	}
	if oSize != h.Size {
		htErr(r.w, "uploading file", fmt.Errorf("expected size=%v actual size=%v", h.Size, oSize))
	}
	log.Printf("Uploaded Dir=%v File=%v Size=%v", r.uDir, h.Filename, h.Size)
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(h.Filename))
}

func (r *wfmRequest) saveText(uData string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if uData == "" {
		htErr(r.w, "text save", fmt.Errorf("zero lenght data"))
		return
	}
	fp := r.uDir + "/" + r.uFbn
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
	err = os.Rename(tmpName, fp)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", r.uDir, fp, len(uData))
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
}

func (r *wfmRequest) mkdir() {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}

	if r.uFbn == "" {
		htErr(r.w, "mkdir", fmt.Errorf("directory name is empty"))
		return
	}
	err := os.Mkdir(r.uDir+"/"+r.uFbn, 0755)
	if err != nil {
		htErr(r.w, "mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
}

func (r *wfmRequest) mkfile() {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}

	if r.uFbn == "" {
		htErr(r.w, "mkfile", fmt.Errorf("file name is empty"))
		return
	}
	f, err := os.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	f.Close()
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
}

func (r *wfmRequest) mkurl(eUrl string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	if r.uFbn == "" {
		htErr(r.w, "mkurl", fmt.Errorf("url file name is empty"))
		return
	}
	if !strings.HasSuffix(r.uFbn, ".url") {
		r.uFbn = r.uFbn + ".url"
	}
	f, err := os.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", eUrl)
	f.Close()
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
}

func (r *wfmRequest) renFile(uNewf string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}

	if r.uFbn == "" || uNewf == "" {
		htErr(r.w, "rename", fmt.Errorf("filename is empty"))
		return
	}
	newB := filepath.Base(uNewf)
	err := os.Rename(
		r.uDir+"/"+r.uFbn,
		r.uDir+"/"+newB,
	)
	if err != nil {
		htErr(r.w, "rename", err)
		return
	}
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(newB))
}

func (r *wfmRequest) moveFiles(uFilePaths []string, uDst string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	uDst = filepath.Clean(uDst)
	log.Printf("move dir=%v files=%+v dst=%v user=%v@%v", r.uDir, uFilePaths, uDst, r.userName, r.remAddr)

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
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(uDst)+"&sort="+r.eSort+"&hi="+url.PathEscape(lF))
}

func (r *wfmRequest) deleteFiles(uFilePaths []string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	log.Printf("delete dir=%v files=%+v user=%v@%v", r.uDir, uFilePaths, r.userName, r.remAddr)

	for _, f := range uFilePaths {
		err := os.RemoveAll(r.uDir + "/" + filepath.Base(f))
		if err != nil {
			htErr(r.w, "delete", err)
			return
		}
	}
	redirect(r.w, *wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort)
}
