package main

import (
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

func listFiles(w http.ResponseWriter, uDir, sort, user string, modern bool) {
	i := candy(modern)
	dir := filepath.Clean(uDir)
	d, err := ioutil.ReadDir(dir)
	if err != nil {
		htErr(w, "Unable to read directory", err)
		return
	}
	sl := []string{}
	sortFiles(d, &sl, sort)

	header(w, uDir, sort)
	eDir := html.EscapeString(dir)
	toolbars(w, eDir, user, sl, i)

	r := 0

	// List Directories First
	for _, f := range d {
		var ldir bool
		var li string
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ls, err := os.Stat(uDir + "/" + f.Name())
			if err != nil {
				continue
			}
			ldir = ls.IsDir()
			li = i["li"]
		}
		if !f.IsDir() && !ldir {
			continue
		}
		if !*sdot && f.Name()[0:1] == "." {
			continue
		}
		if r%2 == 0 {
			w.Write([]byte(`<TR BGCOLOR="#FFFFFF">`))
		} else {
			w.Write([]byte(`<TR BGCOLOR="#F0F0F0">`))
		}
		r++
		fE := html.EscapeString(f.Name())
		w.Write([]byte(`
        <TD NOWRAP ALIGN="left">
        <A HREF="` + *wpfx + `?dir=` + eDir + `/` + fE + `&amp;sort=` + sort + `">` + i["di"] + fE + `/</A>` + li + `</TD>
        <TD NOWRAP></TD>
        <TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
        <TD NOWRAP ALIGN="right">
        <A HREF="` + *wpfx + `?fn=renp&amp;dir=` + eDir + `&amp;oldf=` + fE + `&amp;=sort=` + sort + `">` + i["re"] + `</A>&nbsp;
        <A HREF="` + *wpfx + `?fn=delp&amp;dir=` + eDir + `&amp;file=` + fE + `&amp;=sort=` + sort + `">` + i["rm"] + `</A>&nbsp;
		</TD>
        </TR>
        `))
	}

	// List Files
	for _, f := range d {
		var ldir bool
		var li string
		if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ls, err := os.Stat(uDir + "/" + f.Name())
			if err != nil {
				continue
			}
			ldir = ls.IsDir()
			li = i["li"]
		}
		if f.IsDir() || ldir {
			continue
		}
		if !*sdot && f.Name()[0:1] == "." {
			continue
		}
		if r%2 == 0 {
			w.Write([]byte(`<TR BGCOLOR="#FFFFFF">`))
		} else {
			w.Write([]byte(`<TR BGCOLOR="#F0F0F0">`))
		}
		r++
		fE := html.EscapeString(f.Name())
		w.Write([]byte(`
        <TD NOWRAP ALIGN="LEFT">
        <A HREF="` + *wpfx + `?fn=disp&amp;fp=` + eDir + "/" + fE + `">` + i["fi"] + fE + `</A>` + li + `</TD>
        <TD NOWRAP ALIGN="right">` + humanize.Bytes(uint64(f.Size())) + `</TD>
        <TD NOWRAP ALIGN="right">(` + humanize.Time(f.ModTime()) + `) ` + f.ModTime().Format(time.Stamp) + `</TD>
        <TD NOWRAP ALIGN="right">
        <A HREF="` + *wpfx + `?fn=down&amp;fp=` + eDir + "/" + fE + `&amp;=sort=` + sort + `">` + i["dn"] + `</A>&nbsp;
        <A HREF="` + *wpfx + `?fn=edit&amp;fp=` + eDir + "/" + fE + `&amp;=sort=` + sort + `">` + i["ed"] + `</A>
        <A HREF="` + *wpfx + `?fn=renp&amp;dir=` + eDir + `&amp;oldf=` + fE + `&amp;=sort=` + sort + `">` + i["re"] + `</A>&nbsp;
        <A HREF="` + *wpfx + `?fn=delp&amp;dir=` + eDir + `&amp;file=` + fE + `&amp;=sort=` + sort + `">` + i["rm"] + `</A>&nbsp;
        </TD>
        </TR>
        `))
	}

	w.Write([]byte(`</TABLE>`))
	footer(w)
}

func toolbars(w http.ResponseWriter, eDir, user string, sl []string, i map[string]string) {
	// Topbar
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#FFFFFF" CELLPADDING="0" CELLSPACING="0" BORDER="0" STYLE="height:28px;"><TR>
            <TD NOWRAP  WIDTH="100%" BGCOLOR="#0072c6" VALIGN="MIDDLE" ALIGN="LEFT" STYLE="color:#FFFFFF; font-weight:bold;">
                &nbsp;` + i["tcd"] + eDir + `
            </TD>
            <TD NOWRAP  BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="RIGHT" STYLE="color:#000000; white-space:nowrap">
				<A HREF="` + *wpfx + `?fn=logout">` + i["tid"] + user + `</A>
                <A HREF="` + *wpfx + `?fn=about&amp;dir=` + eDir + `&amp;sort=">&nbsp;` + i["tve"] + ` v` + vers + `&nbsp;</A>
            </TD>
        </TR></TABLE>
        `))

	// Toolbar
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#FFFFFF" CELLPADDING="0" CELLSPACING="0" BORDER="0" STYLE="height:28px;"><TR>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="up" VALUE="` + i["tup"] + `Up" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="home" VALUE="` + i["tho"] + `Home" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="refresh" VALUE="` + i["tre"] + `Refresh" CLASS="nb">
        </TD>
            <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER" >
        <INPUT TYPE="SUBMIT" DISABLED NAME="mdelp" VALUE="` + i["trm"] + `Delete" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" DISABLED NAME="mmovp" VALUE="` + i["tmv"] + `Move" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkd" VALUE="` + i["tdi"] + `New Folder" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkf" VALUE="` + i["tfi"] + `New File" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="SUBMIT" NAME="mkb" VALUE="` + i["tln"] + `New Bookmark" CLASS="nb">
        </TD>
        <TD NOWRAP BGCOLOR="#F1F1F1" VALIGN="MIDDLE" ALIGN="CENTER">
            <INPUT TYPE="FILE" NAME="filename" CLASS="nb">&nbsp;
            <INPUT TYPE="SUBMIT" NAME="upload" VALUE="` + i["tul"] + `Upload" CLASS="nb">
        </TD>
        </TR></TABLE>
        `))

	// Sortby and File List Header
	w.Write([]byte(`
        <TABLE WIDTH="100%" BGCOLOR="#FFFFFF" CELLPADDING="0" CELLSPACING="0" BORDER="0" CLASS="thov"><TR>
        <TD NOWRAP ALIGN="left" WIDTH="50%" BGCOLOR="#A0A0A0">
            <A HREF="` + *wpfx + `?dir=` + eDir + `&amp;sort=` + sl[0] + `"><FONT COLOR="#FFFFFF">` + sl[1] + `</FONT></A>
        </TD>
        <TD NOWRAP ALIGN="right" BGCOLOR="#A0A0A0">
            <A HREF="` + *wpfx + `?dir=` + eDir + `&amp;sort=` + sl[2] + `"><FONT COLOR="#FFFFFF">` + sl[3] + `</FONT></A>
        </TD>
        <TD NOWRAP ALIGN="right"  BGCOLOR="#A0A0A0">
            <A HREF="` + *wpfx + `?dir=` + eDir + `&amp;sort=` + sl[4] + `"><FONT COLOR="#FFFFFF">` + sl[5] + `</FONT></A>
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

func candy(b bool) map[string]string {
	c := map[string]string{
		"fi": "&bull; ",
		"di": "&raquo; ",
		"li": " (link);",

		"rm": "&times;",
		"mv": "&ang;",
		"re": "&ne;",
		"ed": "&para;",
		"dn": "&darr;",

		"tup": "&uarr; ",
		"tho": "&equiv; ",
		"tre": "&reg; ",
		"tid": "User: ",
		"tve": "WFM ",
	}
	if b {
		c = map[string]string{
			"fi": "&#x1F4C4; ",
			"di": "&#x1F4C2; ",
			"li": " &#x1F517; ",

			"rm": "&#x274C;",
			"mv": "&#x1F4E6;",
			"re": "&#x1F3F7;",
			"ed": "&#x1F4DD;",
			"dn": "&#x1F4BE;",

			"tcd": "&#x1F69A; ",
			"tup": "&#x21EA; ",
			"tho": "&#x1F3E0; ",
			"tre": "&#x1F300; ",
			"trm": "&#x274C; ",
			"tmv": "&#x1F4E6; ",
			"tln": "&#x1F310; ",
			"tfi": "&#x1F4D2; ",
			"tdi": "&#x1F4C2; ",
			"tul": "&#x1F680; ",

			"tid": "&#x1F3AB; ",
			"tve": "&#x1F9F0; ",
		}
	}

	return c
}
