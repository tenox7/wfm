package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
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

func rowTag(name, hi string, z int, modern bool) string {
	switch {
	case name == hi:
		return `<TR CLASS="f" BGCOLOR="#33CC33">`
	case z%2 == 0:
		return `<TR CLASS="f" BGCOLOR="#FFFFFF">`
	default:
		return `<TR CLASS="f" BGCOLOR="` + panelGrey[modern] + `">`
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

	header(r.w, r.pfx, r.uDir, r.eSort, "", r.modern)
	bw := bufio.NewWriterSize(r.w, 1<<15)
	toolbars(bw, r.pfx, r.uDir, r.userName, sl, i, r.rwAccess, r.modern)
	qeDir := strings.ReplaceAll(url.PathEscape(r.uDir), `%2F`, `/`)

	z := 0
	var total uint64

	for _, e := range dirs {
		f := e.fi
		var li string
		if e.link {
			li = i["li"]
		}
		bw.WriteString(rowTag(f.Name(), hi, z, r.modern))
		z++
		qeFile := url.PathEscape(f.Name())
		heFile := html.EscapeString(f.Name())
		nUrl, err := url.JoinPath(r.pfx, qeDir, qeFile)
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		if r.eSort != "" {
			nUrl = wfmURL(nUrl, url.Values{"sort": {r.eSort}})
		}
		bw.WriteString(`
			<TD NOWRAP ALIGN="left">
				<INPUT TYPE="CHECKBOX" NAME="mulf" VALUE="` + heFile + `">
				<A HREF="` + html.EscapeString(nUrl) + `">` + i["di"] + heFile + `/</A>` + li + `
			</TD>
			<TD NOWRAP>&nbsp;</TD>
			<TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
			<TD NOWRAP ALIGN="right">`)
		if r.rwAccess {
			q := url.Values{"dir": {r.uDir}, "file": {f.Name()}, "sort": {r.eSort}}
			q.Set("fn", "renp")
			renp := wfmHref(r.pfx, q)
			q.Set("fn", "movp")
			movp := wfmHref(r.pfx, q)
			q.Set("fn", "delp")
			delp := wfmHref(r.pfx, q)
			bw.WriteString(`
				<A HREF="` + renp + `">` + i["re"] + `</A>&nbsp;
				<A HREF="` + movp + `">` + i["mv"] + `</A>&nbsp;
				<A HREF="` + delp + `">` + i["rm"] + `</A>&nbsp;`)
		}
		bw.WriteString(`
			</TD>
		</TR>
		`)
	}

	for _, e := range files {
		f := e.fi
		var li string
		if e.link {
			li = i["li"]
		}
		bw.WriteString(rowTag(f.Name(), hi, z, r.modern))
		z++
		qeFile := url.PathEscape(f.Name())
		heFile := html.EscapeString(f.Name())
		nUrl, err := url.JoinPath(r.pfx, qeDir, qeFile)
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		q := url.Values{"dir": {r.uDir}, "file": {f.Name()}}
		q.Set("fn", "down")
		down := wfmHref(r.pfx, q)
		bw.WriteString(`
			<TD NOWRAP ALIGN="LEFT">
				<INPUT TYPE="CHECKBOX" NAME="mulf" VALUE="` + heFile + `">
				<A HREF="` + html.EscapeString(nUrl) + `">` + fileIcon(qeFile, r.modern) + ` ` + heFile + `</A>` + li + `
			</TD>
			<TD NOWRAP ALIGN="right">` + humanize.Bytes(uint64(f.Size())) + `</TD>
			<TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
			<TD NOWRAP ALIGN="right">
				<A HREF="` + down + `">` + i["dn"] + `</A>&nbsp;`)
		if r.rwAccess {
			q.Set("sort", r.eSort)
			q.Set("fn", "edit")
			edit := wfmHref(r.pfx, q)
			q.Set("fn", "renp")
			renp := wfmHref(r.pfx, q)
			q.Set("fn", "movp")
			movp := wfmHref(r.pfx, q)
			q.Set("fn", "delp")
			delp := wfmHref(r.pfx, q)
			bw.WriteString(`
				<A HREF="` + edit + `">` + i["ed"] + `</A>&nbsp;
				<A HREF="` + renp + `">` + i["re"] + `</A>&nbsp;
				<A HREF="` + movp + `">` + i["mv"] + `</A>&nbsp;
				<A HREF="` + delp + `">` + i["rm"] + `</A>&nbsp;`)
		}
		bw.WriteString(`
			</TD>
		</TR>
		`)
		total = total + uint64(f.Size())
	}

	bw.WriteString(`<TR><TD COLSPAN="4" NOWRAP ALIGN="right" STYLE="border-top:1px solid #999999">` + fmt.Sprint(len(dirs)+len(files)) + ` items, ` +
		humanize.Bytes(total) + ` total </TD></TR>`)
	if r.modern {
		bw.WriteString("\n\t</TBODY>")
	}
	bw.WriteString("\n\t</TABLE>\n")
	bw.Flush()
	footer(r.w)
}

func toolbars(w io.Writer, pfx, uDir, user string, sl []string, i map[string]string, rw, modern bool) {
	eDir := html.EscapeString(uDir)
	qeDir := strings.ReplaceAll(url.PathEscape(uDir), `%2F`, `/`)
	sortBase, err := url.JoinPath(pfx, qeDir)
	if err != nil {
		log.Printf("Unable to build sort url: %v", err)
	}

	theadO, theadC := "", ""
	if modern {
		theadO, theadC = "<THEAD>", "</THEAD><TBODY>"
	}

	io.WriteString(w, `
	<TABLE WIDTH="100%" CELLPADDING="0" CELLSPACING="0" BORDER="0" CLASS="thov">
	`+theadO+`
	<TR>
		<TD BGCOLOR="#0066CC" VALIGN="MIDDLE" ALIGN="LEFT" STYLE="height:28px; color:#FFFFFF; font-weight:bold;">
			<FONT COLOR="#FFFFFF"><B>&nbsp;`+*siteName+`&nbsp;:&nbsp;`+eDir+`</B></FONT>
		</TD>
		<TD BGCOLOR="#0066CC">&nbsp;</TD>
		<TD COLSPAN="2" NOWRAP BGCOLOR="`+panelGrey[modern]+`" VALIGN="MIDDLE" ALIGN="RIGHT" STYLE="height:28px; color:#000000;">
			&nbsp;`+i[rorw[rw]]+`&nbsp;
			<A HREF="`+wfmHref(pfx, url.Values{"fn": {"logout"}})+`">`+i["tid"]+user+`</A>&nbsp;
			<A HREF="`+wfmHref(pfx, url.Values{"fn": {"about"}, "dir": {uDir}})+`">&nbsp;`+i["tve"]+` v`+vers+`&nbsp;</A>
		</TD>
	</TR>
	`)

	btns := []string{
		`<INPUT TYPE="SUBMIT" NAME="up" VALUE="` + i["tup"] + `Up" CLASS="nb">`,
		`<INPUT TYPE="SUBMIT" NAME="home" VALUE="` + i["tho"] + `Home" CLASS="nb">`,
		`<INPUT TYPE="SUBMIT" NAME="refresh" VALUE="` + i["tre"] + `Refresh" CLASS="nb">`,
		`<INPUT TYPE="SUBMIT" NAME="mdelp" VALUE="` + i["trm"] + `Delete" CLASS="nb" ` + disTag[rw] + `>`,
		`<INPUT TYPE="SUBMIT" NAME="mmovp" VALUE="` + i["tmv"] + `Move" CLASS="nb" ` + disTag[rw] + `>`,
		`<INPUT TYPE="SUBMIT" NAME="mkd" VALUE="` + i["tdi"] + `New Dir" CLASS="nb" ` + disTag[rw] + `>`,
		`<INPUT TYPE="SUBMIT" NAME="mkf" VALUE="` + i["tfi"] + `New File" CLASS="nb" ` + disTag[rw] + `>`,
		`<INPUT TYPE="SUBMIT" NAME="mkb" VALUE="` + i["tln"] + `New Link" CLASS="nb" ` + disTag[rw] + `>`,
		`<INPUT TYPE="FILE" NAME="filename" CLASS="nb">`,
		`<INPUT TYPE="SUBMIT" NAME="upload" VALUE="` + i["tul"] + `Upload" CLASS="nb" ` + disTag[rw] + `>`,
	}
	io.WriteString(w, `
	<TR><TD COLSPAN="4" BGCOLOR="`+panelGrey[modern]+`" VALIGN="MIDDLE" ALIGN="LEFT" STYLE="height:28px;">`)
	if modern {
		io.WriteString(w, strings.Join(btns, "\n"))
	} else {
		io.WriteString(w, `<TABLE BORDER="0" CELLPADDING="0" CELLSPACING="2"><TR>`)
		for _, b := range btns {
			io.WriteString(w, `<TD NOWRAP VALIGN="MIDDLE">`+b+`</TD>`)
		}
		io.WriteString(w, `</TR></TABLE>`)
	}
	io.WriteString(w, `</TD></TR>
	`)

	io.WriteString(w, `
	<TR>
	<TD NOWRAP ALIGN="left" WIDTH="50%" BGCOLOR="#666666">
		<A HREF="`+wfmHref(sortBase, url.Values{"sort": {sl[0]}})+`"><FONT COLOR="#FFFFFF">`+sl[1]+`</FONT></A>
	</TD>
	<TD NOWRAP ALIGN="right" BGCOLOR="#666666">
		<A HREF="`+wfmHref(sortBase, url.Values{"sort": {sl[2]}})+`"><FONT COLOR="#FFFFFF">`+sl[3]+`</FONT></A>
	</TD>
	<TD NOWRAP ALIGN="right" BGCOLOR="#666666">
		<A HREF="`+wfmHref(sortBase, url.Values{"sort": {sl[4]}})+`"><FONT COLOR="#FFFFFF">`+sl[5]+`</FONT></A>
	</TD>
	<TD NOWRAP ALIGN="right" BGCOLOR="#666666">&nbsp;</TD>
	</TR>
	`+theadC+`
	`)
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
