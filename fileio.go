package main

import (
	_ "embed"

	"bytes"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
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

//go:embed favicon.ico
var favIcn []byte

//go:embed robots.txt
var robotsTxt []byte

var (
	rlBu *ratelimit.Bucket
)

// validName reports whether n names a single real entry inside the current
// directory. It rejects components that filepath.Base collapses to the current
// or parent directory or the bare root separator ("", ".", "..", "/"). Without
// this a crafted op=rm/mv/re request (or a mulf value) of "/" resolves via
// filepath.Base to the mount root and would recursively delete or move the
// entire served tree.
func validName(n string) bool {
	switch filepath.Base(n) {
	case "", ".", "..", string(os.PathSeparator):
		return false
	}
	return true
}

func (r *wfmRequest) dispFile() {
	fp := r.uDir + "/" + r.uFbn
	ext := strings.ToLower(filepath.Ext(fp))
	log.Printf("Disposition file=%v ext=%v", fp, ext)

	if *convertPng != "" && !r.modern && ext == ".png" {
		r.convPng(fp, *convertPng)
		return
	}

	// inexpensive file handlers
	switch ext {
	case ".url", ".desktop", ".webloc":
		r.gourl(fp)
		return
	}

	if !*listArc {
		r.dispInline(fp)
		return
	}

	// expensive file handlers
	switch ext {
	case ".7z":
		r.list7z(fp)
	case ".zip", ".rar", ".tar", ".gz", ".bz2", ".xz", ".tgz", ".tbz2", ".txz", ".br", ".tbr":
		r.listArchive(fp)
	case ".iso":
		r.listIso(fp)

	default:
		r.dispInline(fp)
	}
}

func dispoHeader(disp, name string) string {
	var ascii strings.Builder
	utf8 := false
	for _, c := range name {
		switch {
		case c > 0x7e:
			utf8 = true
			ascii.WriteByte('_')
		case c < 0x20, c == '"', c == '\\':
			ascii.WriteByte('_')
		default:
			ascii.WriteRune(c)
		}
	}
	h := disp + `; filename="` + ascii.String() + `"`
	if utf8 {
		h += `; filename*=UTF-8''` + strings.ReplaceAll(url.PathEscape(name), "'", "%27")
	}
	return h
}

func (r *wfmRequest) downFile() {
	fp := r.uDir + "/" + r.uFbn
	fi, err := r.fs.Stat(fp)
	if err != nil {
		r.htErr("Unable to get file attributes", err)
		return
	}
	f, err := r.fs.Open(fp)
	if err != nil {
		r.htErr("Unable to open file", err)
		return
	}
	defer f.Close()
	r.w.Header().Set("Content-Type", "application/octet-stream")
	r.w.Header().Set("Content-Disposition", dispoHeader("attachment", r.uFbn))
	r.w.Header().Set("Cache-Control", *cacheCtl)
	serveContent(r.w, r.req, f, fi)
}

func (r *wfmRequest) dispInline(uFilePath string) {
	fi, err := r.fs.Stat(uFilePath)
	if err != nil {
		r.htErr("Unable to get file attributes", err)
		return
	}
	f, err := r.fs.Open(uFilePath)
	if err != nil {
		r.htErr("Unable to open file", err)
		return
	}
	defer f.Close()

	mt, err := mimetype.DetectReader(f)
	if err != nil {
		r.htErr("Unable to determine file type", err)
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		r.htErr("Unable to seek file", err)
		return
	}

	r.w.Header().Set("Content-Type", mt.String())
	r.w.Header().Set("Content-Disposition", dispoHeader("inline", filepath.Base(uFilePath)))
	r.w.Header().Set("Cache-Control", *cacheCtl)
	serveContent(r.w, r.req, f, fi)
}

// serveContent streams a file with RFC 7233 range support so media players and
// download managers can seek and resume. ServeContent adds Accept-Ranges,
// Content-Range, Content-Length and handles Range/HEAD/If-Modified-Since; the
// caller sets Content-Type/Disposition. When -rate_limit is set, output is
// throttled by wrapping writes (seeking stays on the file, so ranges still work).
func serveContent(w http.ResponseWriter, req *http.Request, f afero.File, fi os.FileInfo) {
	if *rateLim != 0 {
		w = &throttledWriter{ResponseWriter: w, w: ratelimit.Writer(w, rlBu)}
	}
	http.ServeContent(w, req, fi.Name(), fi.ModTime(), f)
}

type throttledWriter struct {
	http.ResponseWriter
	w io.Writer
}

func (t *throttledWriter) Write(p []byte) (int, error) { return t.w.Write(p) }

func (r *wfmRequest) convPng(uFilePath, format string) {
	fi, err := r.fs.Open(uFilePath)
	if err != nil {
		r.htErr("Unable to open file", err)
		return
	}
	defer fi.Close()

	img, err := png.Decode(fi)
	if err != nil {
		r.htErr("Unable to decode png", err)
		return
	}

	var buf bytes.Buffer
	var ctype string
	switch format {
	case "gif":
		ctype = "image/gif"
		err = gif.Encode(&buf, img, nil)
	case "jpg":
		ctype = "image/jpeg"
		err = jpeg.Encode(&buf, img, nil)
	}
	if err != nil {
		r.htErr("Unable to encode image", err)
		return
	}

	r.w.Header().Set("Content-Type", ctype)
	r.w.Header().Set("Content-Disposition", dispoHeader("inline", filepath.Base(uFilePath)))
	r.w.Header().Set("Content-Length", fmt.Sprint(buf.Len()))
	r.w.Header().Set("Cache-Control", *cacheCtl)

	var src io.Reader = &buf
	if *rateLim != 0 {
		src = ratelimit.Reader(&buf, rlBu)
	}
	io.Copy(r.w, src)
}

func (r *wfmRequest) uploadFile(h *multipart.FileHeader, f multipart.File) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}
	defer f.Close()

	h.Filename = strings.ReplaceAll(h.Filename, "\\", string(os.PathSeparator))
	fi, err := r.fs.OpenFile(filepath.Join(r.uDir, "/", filepath.Base(h.Filename)), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		r.htErr("unable to write file", err)
		return
	}
	defer fi.Close()

	var w io.Writer = fi
	if *rateLim != 0 {
		w = ratelimit.Writer(fi, rlBu)
	}

	oSize, err := io.Copy(w, f)
	if err != nil {
		r.htErr("uploading file", err)
		return
	}
	if oSize != h.Size {
		r.htErr("uploading file", fmt.Errorf("expected size=%v actual size=%v", h.Size, oSize))
	}
	log.Printf("Uploaded Dir=%v File=%v Size=%v", r.uDir, h.Filename, h.Size)
	r.redirectDir(r.uDir, filepath.Base(h.Filename))
}

func (r *wfmRequest) saveText(uData, crlf string) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}
	if uData == "" {
		r.htErr("text save", fmt.Errorf("zero lenght data"))
		return
	}
	// Normalize to the requested line endings authoritatively, independent of
	// what the client submitted: a textarea posts CRLF, CodeMirror posts LF.
	// Collapse everything to LF first (handles CRLF and bare CR), then expand
	// to CRLF if requested.
	uData = strings.ReplaceAll(uData, "\r\n", "\n")
	uData = strings.ReplaceAll(uData, "\r", "\n")
	if crlf == "CRLF" {
		uData = strings.ReplaceAll(uData, "\n", "\r\n")
	}
	fp := r.uDir + "/" + r.uFbn
	tmpName := fp + ".tmp"
	err := afero.WriteFile(r.fs, tmpName, []byte(uData), 0644)
	if err != nil {
		r.htErr("text save", err)
		return
	}
	f, err := r.fs.Stat(tmpName)
	if err != nil {
		r.htErr("text save", err)
		return
	}
	if f.Size() != int64(len(uData)) {
		r.htErr("text save", fmt.Errorf("temp file size != input size"))
		return
	}
	err = r.fs.Rename(tmpName, fp)
	if err != nil {
		r.htErr("text save", err)
		return
	}
	log.Printf("Saved Text Dir=%v File=%v Size=%v", r.uDir, fp, len(uData))
	r.redirectDir(r.uDir, r.uFbn)
}

func (r *wfmRequest) mkdir() {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}

	if !validName(r.uFbn) {
		r.htErr("mkdir", fmt.Errorf("invalid mkdir name"))
		return
	}
	err := r.fs.Mkdir(r.uDir+"/"+r.uFbn, 0755)
	if err != nil {
		r.htErr("mkdir", err)
		log.Printf("mkdir error: %v", err)
		return
	}
	r.redirectDir(r.uDir, r.uFbn)
}

func (r *wfmRequest) mkfile() {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}

	if !validName(r.uFbn) {
		r.htErr("mkfile", fmt.Errorf("invalid mkfile name"))
		return
	}
	f, err := r.fs.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		r.htErr("mkfile", err)
		return
	}
	f.Close()
	r.redirectDir(r.uDir, r.uFbn)
}

func (r *wfmRequest) mkurl(eUrl string) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}
	if !validName(r.uFbn) {
		r.htErr("mkurl", fmt.Errorf("invalid mkurl name"))
		return
	}
	if !strings.HasSuffix(r.uFbn, ".url") {
		r.uFbn = r.uFbn + ".url"
	}
	f, err := r.fs.OpenFile(r.uDir+"/"+r.uFbn, os.O_RDWR|os.O_EXCL|os.O_CREATE, 0644)
	if err != nil {
		r.htErr("mkfile", err)
		return
	}
	// TODO(tenox): add upport for creating webloc, desktop and other formats
	fmt.Fprintf(f, "[InternetShortcut]\r\nURL=%s\r\n", eUrl)
	f.Close()
	r.redirectDir(r.uDir, r.uFbn)
}

func (r *wfmRequest) renFile(uNewf string) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}

	newB := filepath.Base(uNewf)
	if !validName(r.uFbn) || !validName(newB) {
		r.htErr("rename", fmt.Errorf("invalid file name"))
		return
	}
	err := r.fs.Rename(
		r.uDir+"/"+r.uFbn,
		r.uDir+"/"+newB,
	)
	if err != nil {
		r.htErr("rename", err)
		return
	}
	r.redirectDir(r.uDir, newB)
}

func (r *wfmRequest) moveFiles(uFilePaths []string, uDst string) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}
	uDst = filepath.Clean(uDst)
	log.Printf("move dir=%v files=%+v dst=%v user=%v@%v", r.uDir, uFilePaths, uDst, r.userName, r.remAddr)

	lF := ""
	for _, f := range uFilePaths {
		if !validName(f) {
			r.htErr("move", fmt.Errorf("invalid file name %q", f))
			return
		}
		fb := filepath.Base(f)
		err := r.fs.Rename(
			r.uDir+"/"+fb,
			filepath.Clean(uDst+"/"+fb),
		)
		if err != nil {
			r.htErr("move", err)
			return
		}
		lF = fb
	}
	r.redirectDir(uDst, lF)
}

func (r *wfmRequest) deleteFiles(uFilePaths []string) {
	if !r.rwAccess {
		r.htErr("permission", fmt.Errorf("read only"))
		return
	}
	log.Printf("delete dir=%v files=%+v user=%v@%v", r.uDir, uFilePaths, r.userName, r.remAddr)

	for _, f := range uFilePaths {
		if !validName(f) {
			r.htErr("delete", fmt.Errorf("invalid file name %q", f))
			return
		}
		err := r.fs.RemoveAll(r.uDir + "/" + filepath.Base(f))
		if err != nil {
			r.htErr("delete", err)
			return
		}
	}
	r.redirectDir(r.uDir, "")
}

func (r *wfmRequest) dispOrDir(hi string) {
	f, err := r.fs.Stat(r.uDir)
	if err != nil {
		switch r.uDir {
		case "/favicon.ico":
			if len(favIcn) == 0 {
				break
			}
			r.w.Header().Set("Content-Type", "image/x-icon")
			r.w.Write(favIcn)
			return
		case "/robots.txt":
			if len(robotsTxt) == 0 {
				break
			}
			r.w.Header().Set("Content-Type", "text/plain")
			r.w.Write(robotsTxt)
			return
		}
		r.htErr("error stat() file", err)
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
