package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/juju/ratelimit"
	"github.com/spf13/afero"
)

var (
	rlBu *ratelimit.Bucket
)

func (r *wfmRequest) dispFile() {
	fp := r.uDir + "/" + r.uFbn
	ext := strings.ToLower(filepath.Ext(fp))
	log.Printf("Dsiposition file=%v ext=%v", fp, ext)

	// inexpensive file handlers
	switch ext {
	case ".url", ".desktop", ".webloc":
		gourl(r.w, fp, r.fs)
		return
	}

	if !*listArc {
		dispInline(r.w, fp, r.fs)
		return
	}

	// expensive file handlers
	switch ext {
	case ".zip":
		listZip(r.w, fp, r.fs)
	case ".7z":
		list7z(r.w, fp, r.fs)
	// currently doesnt work with afero fs
	// case "tar", "rar", "gz", "bz2", "xz", "tgz", "tbz2", "txz":
	//listArchive(r.w, fp)
	case ".iso":
		listIso(r.w, fp, r.fs)

	default:
		dispInline(r.w, fp, r.fs)
	}
}

func (r *wfmRequest) downFile() {
	fp := r.uDir + "/" + r.uFbn
	f, err := r.fs.Stat(fp)
	if err != nil {
		htErr(r.w, "Unable to get file attributes", err)
		return
	}
	r.w.Header().Set("Content-Type", "application/octet-stream")
	r.w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(r.uFbn)+"\";")
	r.w.Header().Set("Content-Length", fmt.Sprint(f.Size()))
	r.w.Header().Set("Cache-Control", *cacheCtl)
	streamFile(r.w, fp, r.fs)
}

func dispInline(w http.ResponseWriter, uFilePath string, wfs afero.Fs) {
	f, err := wfs.Stat(uFilePath)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}

	fi, err := wfs.Open(uFilePath)
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
	streamFile(w, uFilePath, wfs)
}

func streamFile(w http.ResponseWriter, uFilePath string, wfs afero.Fs) {
	fi, err := wfs.Open(uFilePath)
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

	// TODO(tenox): this needs to be filepath.join
	fi, err := r.fs.OpenFile(r.uDir+"/"+filepath.Base(h.Filename), os.O_RDWR|os.O_CREATE, 0644)
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
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(h.Filename))
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
	err := afero.WriteFile(r.fs, tmpName, []byte(uData), 0644)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	f, err := r.fs.Stat(tmpName)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	if f.Size() != int64(len(uData)) {
		htErr(r.w, "text save", fmt.Errorf("temp file size != input size"))
		return
	}
	err = r.fs.Rename(tmpName, fp)
	if err != nil {
		htErr(r.w, "text save", err)
		return
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", r.uDir, fp, len(uData))
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
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
	err := r.fs.Mkdir(r.uDir+"/"+r.uFbn, 0755)
	if err != nil {
		htErr(r.w, "mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
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
	f, err := r.fs.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	f.Close()
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
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
	f, err := r.fs.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		htErr(r.w, "mkfile", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", eUrl)
	f.Close()
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(r.uFbn))
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
	err := r.fs.Rename(
		r.uDir+"/"+r.uFbn,
		r.uDir+"/"+newB,
	)
	if err != nil {
		htErr(r.w, "rename", err)
		return
	}
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort+"&hi="+url.PathEscape(newB))
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
		err := r.fs.Rename(
			r.uDir+"/"+fb,
			filepath.Clean(uDst+"/"+fb),
		)
		if err != nil {
			htErr(r.w, "move", err)
			return
		}
		lF = fb
	}
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(uDst)+"&sort="+r.eSort+"&hi="+url.PathEscape(lF))
}

func (r *wfmRequest) deleteFiles(uFilePaths []string) {
	if !r.rwAccess {
		htErr(r.w, "permission", fmt.Errorf("read only"))
		return
	}
	log.Printf("delete dir=%v files=%+v user=%v@%v", r.uDir, uFilePaths, r.userName, r.remAddr)

	for _, f := range uFilePaths {
		err := r.fs.RemoveAll(r.uDir + "/" + filepath.Base(f))
		if err != nil {
			htErr(r.w, "delete", err)
			return
		}
	}
	redirect(r.w, wfmPfx+"?dir="+url.PathEscape(r.uDir)+"&sort="+r.eSort)
}

func (r *wfmRequest) dispOrDir(hi string) {
	f, err := r.fs.Stat(r.uDir)
	if err != nil {
		htErr(r.w, "error checking file", err)
		return
	}
	if f.IsDir() {
		r.listFiles(hi)
		return
	}
	r.uFbn = filepath.Base(r.uDir)
	r.uDir = filepath.Dir(r.uDir)
	r.dispFile()
}
