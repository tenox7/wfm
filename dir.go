package main

import (
	"html"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/spf13/afero"
)

var (
	rorw = map[bool]string{
		true:  "rw",
		false: "ro",
	}
)

type dirEntry struct {
	fi   os.FileInfo
	link bool
}

type sortLink struct{ Href, Label string }

type dirRow struct {
	BgColor, Name, Href string
	IsLink              bool
	Time                string
	Ren, Mov, Del       string
}

type fileRow struct {
	BgColor, Name, Href, Icon string
	IsLink                    bool
	Size, Time, Down          string
	Edit, Ren, Mov, Del       string
}

type dirPage struct {
	chrome
	RW                           bool
	I                            map[string]string
	RwIcon, User, Vers           string
	LogoutHref, AboutHref        string
	SortName, SortSize, SortTime sortLink
	Dirs                         []dirRow
	Files                        []fileRow
	Count                        int
	Total                        string
}

func rowColor(name, hi string, z int, modern bool) string {
	switch {
	case name == hi:
		return "#33CC33"
	case z%2 == 0:
		return "#FFFFFF"
	default:
		return panelGrey[modern]
	}
}

func (r *wfmRequest) listFiles(hi string) {
	i := icons(r.modern)
	d, err := afero.ReadDir(r.fs, r.uDir)
	if err != nil {
		htErr(r.w, "Unable to read directory", err)
		return
	}
	sl := []string{}
	sortFiles(d, &sl, r.eSort)

	var dirs, files []dirEntry
	for _, f := range d {
		if !*showDot && strings.HasPrefix(f.Name(), ".") {
			continue
		}
		isDir, link := f.IsDir(), false
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ls, err := r.fs.Stat(r.uDir + "/" + f.Name())
			if err != nil {
				continue
			}
			isDir, link = ls.IsDir(), true
		}
		if isDir {
			dirs = append(dirs, dirEntry{f, link})
			continue
		}
		files = append(files, dirEntry{f, link})
	}

	qeDir := strings.ReplaceAll(url.PathEscape(r.uDir), `%2F`, `/`)
	sortBase, err := url.JoinPath(r.pfx, qeDir)
	if err != nil {
		log.Printf("Unable to build sort url: %v", err)
	}

	page := dirPage{
		chrome:     r.chrome(""),
		RW:         r.rwAccess,
		I:          i,
		RwIcon:     i[rorw[r.rwAccess]],
		User:       r.userName,
		Vers:       vers,
		LogoutHref: wfmHref(r.pfx, url.Values{"fn": {"logout"}}),
		AboutHref:  wfmHref(r.pfx, url.Values{"fn": {"about"}, "dir": {r.uDir}}),
		SortName:   sortLink{wfmHref(sortBase, url.Values{"sort": {sl[0]}}), sl[1]},
		SortSize:   sortLink{wfmHref(sortBase, url.Values{"sort": {sl[2]}}), sl[3]},
		SortTime:   sortLink{wfmHref(sortBase, url.Values{"sort": {sl[4]}}), sl[5]},
	}

	z := 0
	for _, e := range dirs {
		f := e.fi
		nUrl, err := url.JoinPath(r.pfx, qeDir, url.PathEscape(f.Name()))
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		if r.eSort != "" {
			nUrl = wfmURL(nUrl, url.Values{"sort": {r.eSort}})
		}
		q := url.Values{"dir": {r.uDir}, "file": {f.Name()}, "sort": {r.eSort}}
		q.Set("fn", "renp")
		ren := wfmHref(r.pfx, q)
		q.Set("fn", "movp")
		mov := wfmHref(r.pfx, q)
		q.Set("fn", "delp")
		del := wfmHref(r.pfx, q)
		page.Dirs = append(page.Dirs, dirRow{
			BgColor: rowColor(f.Name(), hi, z, r.modern),
			Name:    html.EscapeString(f.Name()),
			Href:    html.EscapeString(nUrl),
			IsLink:  e.link,
			Time:    "(" + humanize.Time(f.ModTime()) + ") " + f.ModTime().Format(time.Stamp),
			Ren:     ren,
			Mov:     mov,
			Del:     del,
		})
		z++
	}

	var total uint64
	for _, e := range files {
		f := e.fi
		qeFile := url.PathEscape(f.Name())
		nUrl, err := url.JoinPath(r.pfx, qeDir, qeFile)
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		q := url.Values{"dir": {r.uDir}, "file": {f.Name()}}
		q.Set("fn", "down")
		down := wfmHref(r.pfx, q)
		q.Set("sort", r.eSort)
		q.Set("fn", "edit")
		edit := wfmHref(r.pfx, q)
		q.Set("fn", "renp")
		ren := wfmHref(r.pfx, q)
		q.Set("fn", "movp")
		mov := wfmHref(r.pfx, q)
		q.Set("fn", "delp")
		del := wfmHref(r.pfx, q)
		page.Files = append(page.Files, fileRow{
			BgColor: rowColor(f.Name(), hi, z, r.modern),
			Name:    html.EscapeString(f.Name()),
			Href:    html.EscapeString(nUrl),
			Icon:    fileIcon(qeFile, r.modern),
			IsLink:  e.link,
			Size:    humanize.Bytes(uint64(f.Size())),
			Time:    "(" + humanize.Time(f.ModTime()) + ") " + f.ModTime().Format(time.Stamp),
			Down:    down,
			Edit:    edit,
			Ren:     ren,
			Mov:     mov,
			Del:     del,
		})
		total += uint64(f.Size())
		z++
	}

	page.Count = len(dirs) + len(files)
	page.Total = humanize.Bytes(total)
	r.render("dir", page)
}

func sortFiles(f []os.FileInfo, l *[]string, by string) {
	switch by {
	// size
	case "sa":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Size() < f[j].Size()
		})
		*l = []string{"na", "Name", "sd", "v Size", "ta", "Time Modified"}
		return
	case "sd":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Size() > f[j].Size()
		})
		*l = []string{"na", "Name", "sa", "^ Size", "ta", "Time Modified"}
		return

	// time
	case "ta":
		sort.Slice(f, func(i, j int) bool {
			return f[i].ModTime().Before(f[j].ModTime())
		})
		*l = []string{"na", "Name", "sa", "Size", "td", "v Time Modified"}
		return
	case "td":
		sort.Slice(f, func(i, j int) bool {
			return f[i].ModTime().After(f[j].ModTime())
		})
		*l = []string{"na", "Name", "sa", "Size", "ta", "^ Time Modified"}
		return

	// name
	case "nd":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Name() > f[j].Name()
		})
		*l = []string{"na", "^ Name", "sa", "Size", "ta", "Time Modified"}
		return
	default:
		*l = []string{"nd", "v Name", "sa", "Size", "ta", "Time Modified"}
		return
	}
}

func icons(m bool) map[string]string {
	if m {
		return map[string]string{
			"fi": "&#x1F5D2; ",
			"di": "&#x1F4C2; ",
			"li": " &#x1F517; ",

			"rm": "&#x274C;",
			"mv": "&#x1F69A;",
			"re": "&#x1F4AC;",
			"ed": "&#x1F4DD;",
			"dn": "&#x1F4BE;",

			"tcd": "&#x1F371; ",
			"tup": "&#x1F53A; ",
			"tho": "&#x1F3E0; ",
			"tre": "&#x1F300; ",
			"trm": "&#x274C; ",
			"tmv": "&#x1F69A; ",
			"tln": "&#x1F310; ",
			"tfi": "&#x1F4D2; ",
			"tdi": "&#x1F4C2; ",
			"tul": "&#x1F680; ",

			"tid": "&#x1F3AB; ",
			"tve": "&#x1F9F0; ",

			"rw": "&#x1F511; rw",
			"ro": "&#x1F512; ro",
		}
	}

	return map[string]string{
		"fi": " ",
		"di": " ",
		"li": " (link)",

		"rm": "[rm]",
		"mv": "[mv]",
		"re": "[re]",
		"ed": "[ed]",
		"dn": "[dn]",

		"tup": "^ ",
		"tho": "~ ",
		"tre": "&reg; ",
		"tid": "User: ",
		"tve": "WFM ",

		"rw": "[rw]",
		"ro": "[ro]",
	}
}

func fileIcon(f string, m bool) string {
	if !m {
		return ""
	}
	s := strings.Split(f, ".")
	switch strings.ToLower(s[len(s)-1]) {
	case "iso", "udf":
		return "&#x1F4BF;"
	case "mp4", "mov", "qt", "avi", "mpg", "mpeg", "mkv":
		return "&#x1F3AC;"
	case "gif", "png", "jpg", "jpeg", "ico", "webp", "bmp", "tif", "tiff", "heif", "heic":
		return "&#x1F5BC;"
	case "deb", "rpm", "dpkg", "apk", "msi", "pkg":
		return "&#x1F381;"
	case "zip", "rar", "7z", "z", "gz", "bz2", "xz", "lz", "tgz", "tbz", "txz", "arj", "lha", "tar":
		return "&#x1F4E6;"
	case "imd", "img", "raw", "dd", "tap", "dsk", "dmg":
		return "&#x1F4BE;"
	case "txt", "log", "csv", "md", "mhtml", "html", "htm", "cfg", "conf", "ini", "json", "xml":
		return "&#x1F4C4;"
	case "pdf", "ps", "doc", "docx", "xls", "xlsx", "rtf":
		return "&#x1F4DA;"
	case "url", "desktop", "webloc":
		return "&#x1F310;"
	}
	return "&#x1F4D2;"
}
