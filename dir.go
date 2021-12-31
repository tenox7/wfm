package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

func sortFiles(f []os.FileInfo, l *[]string, by string) {
	switch by {
	case "nd":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Name() > f[j].Name()
		})
		*l = []string{"na", "&Delta;Name", "sa", "Size", "ta", "Time"}

	case "sa":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Size() < f[j].Size()
		})
		*l = []string{"na", "Name", "sd", "&nabla;Size", "ta", "Time"}
	case "sd":
		sort.Slice(f, func(i, j int) bool {
			return f[i].Size() > f[j].Size()
		})
		*l = []string{"na", "Name", "sa", "&Delta;Size", "ta", "Time"}

	case "ta":
		sort.Slice(f, func(i, j int) bool {
			return f[i].ModTime().Before(f[j].ModTime())
		})
		*l = []string{"na", "Name", "sa", "Size", "td", "&nabla;Time"}
	case "td":
		sort.Slice(f, func(i, j int) bool {
			return f[i].ModTime().After(f[j].ModTime())
		})
		*l = []string{"na", "Name", "sa", "Size", "ta", "&Delta;Time"}

	default:
		*l = []string{"nd", "&nabla;Name", "sa", "Size", "ta", "Time"}
		return
	}
}

func listFiles(w http.ResponseWriter, dir, sort string) {
	header(w, dir)

	d, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		log.Printf("Error: %v", err)
		return
	}
	sl := []string{}
	sortFiles(d, &sl, sort)

	// Topbar
	fmt.Fprintf(w,
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"+
			"<TR>\n"+
			"<TD NOWRAP  WIDTH=\"100%%\" BGCOLOR=\"#0072c6\" VALIGN=\"MIDDLE\" ALIGN=\"LEFT\" STYLE=\"color:#FFFFFF; font-weight:bold;\">\n"+
			"&nbsp;%v\n"+
			"<TD NOWRAP  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"RIGHT\" STYLE=\"color:#000000; font-weight:bold;  white-space:nowrap\">\n"+
			"<A HREF=\"/?a=about&amp;dir=%s&amp;\">&nbsp;WFM v%s&nbsp;</A>"+
			"</TD>\n"+
			"</TR>\n"+
			"</TABLE>\n",
		dir, dir, "2.0",
	)

	// Toolbar
	fmt.Fprintf(w,
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"+
			"<TR>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\" NAME=\"up\" VALUE=\"&uarr; Up\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\" NAME=\"home\" VALUE=\"&equiv; Home\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\" NAME=\"refresh\" VALUE=\"&prop; Refresh\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\" NAME=\"mdelp\" VALUE=\"&otimes; Delete\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\"  NAME=\"mmovp\" VALUE=\"&ang; Move\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\"  NAME=\"ndirp\" VALUE=\"&lowast; New Folder\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\"  NAME=\"nfilep\" VALUE=\"&oplus; New File\">\n"+
			"</TD>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"file\" NAME=\"filename\">&nbsp;\n"+
			"<INPUT TYPE=\"SUBMIT\"  NAME=\"upload\" VALUE=\"&Theta; Upload\">\n"+
			"</TD>\n"+

			"</TR></TABLE>\n")

	// Sortby Header
	w.Write([]byte(
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=0 CELLSPACING=0 BORDER=0>\n" +
			"<TR BGCOLOR=\"#FFFFFF\" >\n" +
			"<TD NOWRAP  ALIGN=\"left\" WIDTH=\"50%%\" BGCOLOR=\"#A0A0A0\">\n" +
			"<A HREF=\"/?dir=" + dir + "&sort=" + sl[0] + "\"><FONT COLOR=\"#FFFFFF\">" + sl[1] + "</FONT></A>\n" +
			"</TD>\n" +
			"<TD NOWRAP  ALIGN=\"right\" BGCOLOR=\"#A0A0A0\">\n" +
			"<A HREF=\"/?dir=" + dir + "&sort=" + sl[2] + "\"><FONT COLOR=\"#FFFFFF\">" + sl[3] + "</FONT></A>\n" +
			"</TD>\n" +
			"<TD NOWRAP  ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n" +
			"<A HREF=\"/?dir=" + dir + "&sort=" + sl[4] + "\"><FONT COLOR=\"#FFFFFF\">" + sl[5] + "</FONT></A>\n" +
			"</TD>\n" +
			"<TD NOWRAP  ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n" +
			"&nbsp;\n" +
			"</TD>\n" +
			"<TD NOWRAP  ALIGN=\"left\"  BGCOLOR=\"#A0A0A0\">\n" +
			"&nbsp;\n" +
			"</TD>\n" +
			"</TR>\n",
	))

	// List Directories First
	for _, f := range d {
		if !f.IsDir() {
			continue
		}
		fmt.Fprintf(w, "<TR><TD NOWRAP  ALIGN=\"LEFT\">&raquo; <A HREF=\"/?dir=%v\">%v&frasl;</A></TD>"+
			"<TD NOWRAP ALIGN=\"right\"></TD>"+
			"<TD NOWRAP ALIGN=\"right\">(%s) %s</TD>"+
			"<TD NOWRAP ALIGN=\"right\">&hellip; &ang; &otimes; &crarr;</TD>"+
			"</TR>\n",
			html.EscapeString(dir+"/"+f.Name()),
			html.EscapeString(f.Name()),
			humanize.Time(f.ModTime()),
			f.ModTime().Format(time.Stamp),
		)
	}

	// List Files
	for _, f := range d {
		if f.IsDir() {
			continue
		}
		fmt.Fprintf(w, "<TR><TD NOWRAP  ALIGN=\"LEFT\">&bull; %v</A></TD>"+
			"<TD NOWRAP ALIGN=\"right\">%v</TD>"+
			"<TD NOWRAP ALIGN=\"right\">(%s) %s</TD>"+
			"<TD NOWRAP ALIGN=\"right\">&hellip; &ang; &otimes; &crarr;</TD>"+
			"</TR>\n",
			html.EscapeString(f.Name()),
			humanize.Bytes(uint64(f.Size())),
			humanize.Time(f.ModTime()),
			f.ModTime().Format(time.Stamp),
		)
	}

	fmt.Fprintf(w, "</FORM></TABLE></BODY></HTML>\n")
}
