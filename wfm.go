// Web File Manager
//
// TODO:
// * dilist sorting
// * dirlist alternate line colors
// * file routines
// * authentication
// * setuid/setgid
// * https/certbot
// * git client
// * docker support (no chroot) - mount dir as / ?
// * drivers for different storage, like cloud/smb/ftp
// * html charset, currently US-ASCII ?!
// * generate icons on fly with encoding/gid
//   also for input type=image, or  least for favicon?
// time/date format as flag?

package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
)

var (
	addr = flag.String("addr", ":8080", "Listen address and port")
	base = flag.String("base_dir", "", "Base directory path")
)

func header(w http.ResponseWriter, dir string) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(
		"<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n\"http://www.w3.org/TR/html4/loose.dtd\">\n" +
			"<HTML LANG=\"en\">\n" +
			"<HEAD>\n" +
			"<TITLE>WFM " + dir + "</TITLE>\n" +
			"<STYLE TYPE=\"text/css\">\n" +
			"<!--\n" +
			"A:link {text-decoration: none; color:#0000CE; } \n" +
			"A:visited {text-decoration: none; color:#0000CE; } \n" +
			"A:active {text-decoration: none; color:#FF0000; } \n" +
			"A:hover {text-decoration: none; color:#FF0000; } \n" +
			"html, body, table { width:100%%; margin:0px; padding:0px; border:none; } \n" +
			"td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:12px; margin:0px; padding:2px; border:none; } \n" +
			"input { border-color:#000000; border-style:solid; font-family: Tahoma, Arial, Geneva, sans-serif; font-size:12px; }\n" +
			".hovout { border: none; padding: 0px; background-color: transparent; color: #0000CE; }\n" +
			".hovin  { border: none; padding: 0px; background-color: transparent; color: #FF0000; }\n" +
			"-->\n" +
			"</STYLE>\n" +
			"<META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html;charset=US-ASCII\">\n" +
			"<META HTTP-EQUIV=\"Content-Language\" CONTENT=\"en-US\">\n" +
			"<META HTTP-EQUIV=\"google\" CONTENT=\"notranslate\">\n" +
			"<META NAME=\"viewport\" CONTENT=\"width=device-width\">\n" +
			/*"<LINK REL=\"icon\" TYPE=\"image/gif\" HREF=\"ICONGOESHERE\">\n" +*/
			"</HEAD>\n" +
			"<BODY BGCOLOR=\"#FFFFFF\">\n" +
			"<FORM ACTION=\"/\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n" +
			"<INPUT TYPE=\"hidden\" NAME=\"dir\" VALUE=\"" + dir + "\">\n",
	))
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	dir := filepath.Clean(r.FormValue("dir"))
	if r.FormValue("home") != "" {
		dir = "/"
	}
	if r.FormValue("up") != "" {
		dir = filepath.Dir(dir)
	}
	log.Printf("req from=%q uri=%q reqdir=%q dir=%q", r.RemoteAddr, r.RequestURI, r.FormValue("dir"), dir)
	header(w, dir)

	d, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		log.Printf("Error: %v", err)
		return
	}

	// Topbar
	fmt.Fprintf(w,
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"+
			"<TR>\n"+
			"<TD NOWRAP  WIDTH=\"100%%\" BGCOLOR=\"#0072c6\" VALIGN=\"MIDDLE\" ALIGN=\"LEFT\" STYLE=\"color:#FFFFFF; font-weight:bold;\">\n"+
			"&nbsp;&Xi; WFM %v \n"+
			"<TD NOWRAP  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"RIGHT\" STYLE=\"color:#000000; font-weight:bold;  white-space:nowrap\">\n"+
			"&nbsp;&nbsp;%s&nbsp;"+
			"<A HREF=\"/?a=about&amp;dir=%s&amp;\">&nbsp;v%s&nbsp;</A>"+
			"</TD>\n"+
			"</TR>\n"+
			"</TABLE>\n",
		dir, r.RemoteAddr, dir, "2.0",
	)

	// Toolbar
	fmt.Fprintf(w,
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"+
			"<TR>\n"+
			"<TD NOWRAP BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"+
			"<INPUT TYPE=\"SUBMIT\" NAME=\"up\" VALUE=\"&uArr; Up\">\n"+
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
	fmt.Fprintf(w,
		"<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=0 CELLSPACING=0 BORDER=0>\n"+
			"<TR BGCOLOR=\"#FFFFFF\" >\n"+
			"<TD NOWRAP  ALIGN=\"left\" WIDTH=\"50%%\" BGCOLOR=\"#A0A0A0\">\n"+
			"<FONT COLOR=\"#FFFFFF\">%s</FONT>\n"+
			"</TD>\n"+
			"<TD NOWRAP  ALIGN=\"right\" BGCOLOR=\"#A0A0A0\">\n"+
			"<FONT COLOR=\"#FFFFFF\">%s</FONT>\n"+
			"</TD>\n"+
			"<TD NOWRAP  ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n"+
			"<FONT COLOR=\"#FFFFFF\">%s</FONT>\n"+
			"</TD>\n"+
			"<TD NOWRAP  ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n"+
			"&nbsp;\n"+
			"</TD>\n"+
			"<TD NOWRAP  ALIGN=\"left\"  BGCOLOR=\"#A0A0A0\">\n"+
			"&nbsp;\n"+
			"</TD>\n"+
			"</TR>\n",
		"&nabla;namepfx", "sizepfx", "datepfx")

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

func main() {
	flag.Parse()
	var err error
	if *base != "" {
		err = syscall.Chroot(*base)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Starting WFM on %q for directory %q", *addr, *base)

	http.HandleFunc("/", listFiles)
	http.HandleFunc("/favicon.ico", http.NotFound)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
