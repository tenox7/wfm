package main

import (
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

func (r *wfmRequest) listFiles(hi string) {
	i := icons(r.modern)
	d, err := ioutil.ReadDir(r.uDir)
	if err != nil {
		htErr(r.w, "Unable to read directory", err)
		return
	}
	sl := []string{}
	sortFiles(d, &sl, r.eSort)

	header(r.w, r.uDir, r.eSort, "")
	toolbars(r.w, r.uDir, r.userName, sl, i, r.rwAccess)
	qeDir := url.PathEscape(r.uDir)

	z := 0
	var total uint64

	// List Directories First
	for _, f := range d {
		var ldir bool
		var li string
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ls, err := os.Stat(r.uDir + "/" + f.Name())
			if err != nil {
				continue
			}
			ldir = ls.IsDir()
			li = i["li"]
		}
		if !f.IsDir() && !ldir {
			continue
		}
		if !*showDot && f.Name()[0:1] == "." {
			continue
		}
		if f.Name() == hi {
			r.w.Write([]byte(`<TR BGCOLOR="#33CC33">`))
		} else if z%2 == 0 {
			r.w.Write([]byte(`<TR BGCOLOR="#FFFFFF">`))
		} else {
			r.w.Write([]byte(`<TR BGCOLOR="#F0F0F0">`))
		}
		z++
		qeFile := url.PathEscape(f.Name())
		heFile := html.EscapeString(f.Name())
		nUrl, err := url.JoinPath(*wfmPfx, qeDir, qeFile)
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		if r.eSort != "" {
			nUrl += `?sort=` + r.eSort
		}
		r.w.Write([]byte(`
			<TD NOWRAP ALIGN="left">
				<INPUT TYPE="CHECKBOX" NAME="mulf" VALUE="` + heFile + `">
				<A HREF="` + nUrl + `">` + i["di"] + heFile + `/</A>` + li + `
			</TD>
			<TD NOWRAP>&nbsp;</TD>
			<TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
			<TD NOWRAP ALIGN="right">
		`))
		if r.rwAccess {
			r.w.Write([]byte(`
				<A HREF="` + *wfmPfx + `?fn=renp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["re"] + `</A>&nbsp;
				<A HREF="` + *wfmPfx + `?fn=movp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["mv"] + `</A>&nbsp;
				<A HREF="` + *wfmPfx + `?fn=delp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["rm"] + `</A>&nbsp;
		`))
		}
		r.w.Write([]byte(`
					</TD>
		</TR>
        `))
	}

	// List Files
	for _, f := range d {
		var ldir bool
		var li string
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ls, err := os.Stat(r.uDir + "/" + f.Name())
			if err != nil {
				continue
			}
			ldir = ls.IsDir()
			li = i["li"]
		}
		if f.IsDir() || ldir {
			continue
		}
		if !*showDot && f.Name()[0:1] == "." {
			continue
		}
		if f.Name() == hi {
			r.w.Write([]byte(`<TR BGCOLOR="#33CC33">`))
		} else if z%2 == 0 {
			r.w.Write([]byte(`<TR BGCOLOR="#FFFFFF">`))
		} else {
			r.w.Write([]byte(`<TR BGCOLOR="#F0F0F0">`))
		}
		z++
		qeFile := url.PathEscape(f.Name())
		heFile := html.EscapeString(f.Name())
		nUrl, err := url.JoinPath(*wfmPfx, qeDir, qeFile)
		if err != nil {
			log.Printf("Unable to parse url: %v", err)
		}
		r.w.Write([]byte(`
			<TD NOWRAP ALIGN="LEFT">
				<INPUT TYPE="CHECKBOX" NAME="mulf" VALUE="` + heFile + `">
				<A HREF="` + nUrl + `">` + fileIcon(qeFile, r.modern) + ` ` + heFile + `</A>` + li + `
			</TD>
			<TD NOWRAP ALIGN="right">` + humanize.Bytes(uint64(f.Size())) + `</TD>
			<TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
			<TD NOWRAP ALIGN="right">
				<A HREF="` + *wfmPfx + `?fn=down&amp;dir=` + qeDir + `&amp;file=` + qeFile + `">` + i["dn"] + `</A>&nbsp;
			`))
		if r.rwAccess {
			r.w.Write([]byte(`
				<A HREF="` + *wfmPfx + `?fn=edit&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["ed"] + `</A>&nbsp;
				<A HREF="` + *wfmPfx + `?fn=renp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["re"] + `</A>&nbsp;
				<A HREF="` + *wfmPfx + `?fn=movp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["mv"] + `</A>&nbsp;
				<A HREF="` + *wfmPfx + `?fn=delp&amp;dir=` + qeDir + `&amp;file=` + qeFile + `&amp;sort=` + r.eSort + `">` + i["rm"] + `</A>&nbsp;
			`))
		}
		r.w.Write([]byte(`
				</TD>
        </TR>
        `))
		total = total + uint64(f.Size())
	}

	// Footer
	r.w.Write([]byte(`<TR><TD></TD><TD ALIGN="right" STYLE="border-top:1px solid grey">Total ` +
		humanize.Bytes(total) + `</TD><TD></TD><TD></TD></TR>` + "\n\t</TABLE>\n"))
	footer(r.w)
}

func toolbars(w http.ResponseWriter, uDir, user string, sl []string, i map[string]string, rw bool) {
	eDir := html.EscapeString(uDir)
	qeDir := url.PathEscape(uDir)
	// Topbar
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#FFFFFF" CELLPADDING="0" CELLSPACING="0" BORDER="0" STYLE="height:28px;"><TR>
            <TD NOWRAP  WIDTH="100%" BGCOLOR="#0072c6" VALIGN="MIDDLE" ALIGN="LEFT" STYLE="color:#FFFFFF; font-weight:bold;">
                <FONT COLOR="#FFFFFF">&nbsp;` + *siteName + `&nbsp;:&nbsp;` + eDir + `</FONT>
            </TD>
            <TD NOWRAP  BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="RIGHT" STYLE="color:#000000; white-space:nowrap">
				` + rorw[rw] + `&nbsp;
				<A HREF="` + *wfmPfx + `?fn=logout">` + i["tid"] + user + `</A>&nbsp;
                <A HREF="` + *wfmPfx + `?fn=about&amp;dir=` + qeDir + `&amp;sort=">&nbsp;` + i["tve"] + ` v` + vers + `&nbsp;</A>
            </TD>
        </TR></TABLE>
        `))

	// Toolbar
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#EEEEEE" CELLPADDING="0" CELLSPACING="0" BORDER="0" STYLE="height:28px;"><TR>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="up" VALUE="` + i["tup"] + `Up" CLASS="nb">
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="home" VALUE="` + i["tho"] + `Home" CLASS="nb">
        </TD>
        <TD NOWRAP  VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="refresh" VALUE="` + i["tre"] + `Refresh" CLASS="nb">
        </TD>
            <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER" >
        <INPUT TYPE="SUBMIT" NAME="mdelp" VALUE="` + i["trm"] + `Delete" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mmovp" VALUE="` + i["tmv"] + `Move" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkd" VALUE="` + i["tdi"] + `New Dir" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkf" VALUE="` + i["tfi"] + `New File" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkb" VALUE="` + i["tln"] + `New Link" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        <TD NOWRAP VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="FILE" NAME="filename" CLASS="nb">&nbsp;
            <INPUT TYPE="SUBMIT" NAME="upload" VALUE="` + i["tul"] + `Upload" CLASS="nb" ` + disTag[rw] + `>
        </TD>
        </TR></TABLE>
        `))

	// Sortby and File List Header
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#FFFFFF" CELLPADDING="0" CELLSPACING="0" BORDER="0" CLASS="thov"><TR>
        <TD NOWRAP ALIGN="left" WIDTH="50%" BGCOLOR="#A0A0A0">
            <A HREF="` + *wfmPfx + `/` + qeDir + `?sort=` + sl[0] + `"><FONT COLOR="#FFFFFF">` + sl[1] + `</FONT></A>
        </TD>
        <TD NOWRAP ALIGN="right" BGCOLOR="#A0A0A0">
            <A HREF="` + *wfmPfx + `/` + qeDir + `?sort=` + sl[2] + `"><FONT COLOR="#FFFFFF">` + sl[3] + `</FONT></A>
        </TD>
        <TD NOWRAP ALIGN="right"  BGCOLOR="#A0A0A0">
            <A HREF="` + *wfmPfx + `/` + qeDir + `?sort=` + sl[4] + `"><FONT COLOR="#FFFFFF">` + sl[5] + `</FONT></A>
        </TD>
        <TD NOWRAP  ALIGN="right" BGCOLOR="#A0A0A0">
            &nbsp;
        </TD>
        <TD NOWRAP ALIGN="left" BGCOLOR="#A0A0A0">
            &nbsp;
        </TD>
        </TR>
        `))

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

func icons(b bool) map[string]string {
	if b {
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
		}
	}

	return map[string]string{
		"fi": " ",
		"di": " ",
		"li": " (link);",

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
	}
}

func fileIcon(f string, m bool) string {
	if !m {
		return "&#183;"
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
