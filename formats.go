package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bodgit/sevenzip"
	"github.com/dustin/go-humanize"
	"github.com/kdomanski/iso9660"
	"github.com/mholt/archives"
	"github.com/spf13/afero"
	"gopkg.in/ini.v1"
	"howett.net/plist"
)

func gourl(w http.ResponseWriter, fp string, wfs afero.Fs) {
	var url string
	f, err := afero.ReadFile(wfs, fp)
	if err != nil {
		htErr(w, "go2url", fmt.Errorf("unable to read link file: %v", err))
		return
	}

	e := strings.ToLower(filepath.Ext(fp))
	switch e {
	case ".url":
		i, err := ini.Load(f)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		url = i.Section("InternetShortcut").Key("URL").String()
	case ".desktop":
		i, err := ini.Load(f)
		if err != nil {
			htErr(w, "go2url", err)
			return
		}
		url = i.Section("Desktop Entry").Key("URL").String()
	case ".webloc":
		var p struct {
			URL string
		}
		_, err = plist.Unmarshal(f, &p)
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

func listIso(w http.ResponseWriter, fp string, wfs afero.Fs) {
	// TODO: recursive file list
	f, err := wfs.Open(fp)
	if err != nil {
		htErr(w, "isoread", err)
		return
	}
	defer f.Close()
	// TODO(tenox): add UDF support https://github.com/mogaika/udf
	i, err := iso9660.OpenImage(f)
	if err != nil {
		htErr(w, "isoread", err)
		return
	}
	r, err := i.RootDir()
	if err != nil {
		htErr(w, "isoread", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cacheCtl)
	if r.IsDir() {
		cld, err := r.GetChildren()
		if err != nil {
			htErr(w, "isoread", err)
			return
		}

		for _, c := range cld {
			if c.IsDir() {
				fmt.Fprintf(w, "%v  [dir]\n", c.Name())
				continue
			}
			fmt.Fprintf(w, "%v  %v\n", c.Name(), humanize.Bytes(uint64(c.Size())))
		}
	} else {
		fmt.Fprintf(w, "%v  %v\n", r.Name(), r.Size())
	}
}

func list7z(w http.ResponseWriter, fp string, wfs afero.Fs) {
	f, err := wfs.Open(fp)
	if err != nil {
		htErr(w, "sevenzip: open: ", err)
		return
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		htErr(w, "sevenzip: stat: ", err)
		return
	}
	a, err := sevenzip.NewReader(f, s.Size())
	if err != nil {
		htErr(w, "sevenzip", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cacheCtl)
	for _, f := range a.File {
		fmt.Fprintln(w, f.Name)
	}
}

func listArchive(w http.ResponseWriter, fp string, wfs afero.Fs) {
	f, err := wfs.Open(fp)
	if err != nil {
		htErr(w, "archive: open: ", err)
		return
	}

	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		htErr(w, "archive: stat: ", err)
		return
	}

	af, _, err := archives.Identify(context.TODO(), f.Name(), f)
	if err != nil {
		htErr(w, "archive: identify: ", err)
		return
	}

	// TODO(tenox): https://github.com/mholt/archiver/issues/358
	aa := &archives.ArchiveFS{Stream: io.NewSectionReader(f, 0, s.Size()), Format: af.(archives.Archival)}
	a, err := aa.Sub(".")
	if err != nil {
		htErr(w, "archive: FS: ", err)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", *cacheCtl)

	err = fs.WalkDir(a, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintln(w, err)
			return err
		}
		fmt.Fprintln(w, path)
		return nil
	})
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

// TODO(tenox): finish implementing
// TODO(tenox): aferoize
func du() {
	rf, err := os.Stat("/")
	if err != nil {
		log.Fatal(err)
	}
	rdev := rf.Sys().(*syscall.Stat_t).Dev
	f := os.DirFS("/")
	err = fs.WalkDir(f, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		i, err := d.Info()
		if err != nil {
			return fs.SkipDir
		}
		sys := i.Sys()
		if sys == nil {
			return fs.SkipDir
		}
		dev := sys.(*syscall.Stat_t).Dev
		fmt.Printf("p=%v size=%v dev=%v\n", p, i.Size(), dev)
		if dev != rdev {
			return fs.SkipDir
		}
		return nil
	})
	if err != nil {
		log.Print(err)
	}
}
