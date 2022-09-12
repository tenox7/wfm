package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/dustin/go-humanize"
)

func (r *wfmRequest) prompt(action string, mul []string) {
	header(r.w, r.uDir, r.eSort, "")

	r.w.Write([]byte(`
    <TABLE WIDTH="100%" HEIGHT="90%" BORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD VALIGN="MIDDLE" ALIGN="CENTER">
    <BR>&nbsp;<BR><P>
    <TABLE WIDTH="400" BGCOLOR="#F0F0F0" BORDER="0" CELLSPACING="0" CELLPADDING="1" CLASS="tbr">
      <TR><TD COLSPAN="2" BGCOLOR="#004080"><FONT COLOR="#FFFFFF">&nbsp; ` + action + `</FONT></TD></TR>
      <TR><TD WIDTH="30">&nbsp;</TD><TD>
    `))

	switch action {
	case "mkdir":
		r.w.Write([]byte(`
        &nbsp;<BR>Enter name for the new directory:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        `))
	case "mkfile":
		r.w.Write([]byte(`
        &nbsp;<BR>Enter name for the new file:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        `))
	case "mkurl":
		r.w.Write([]byte(`
        &nbsp;<BR>Enter name for the new url file:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        &nbsp;<BR>Destination URL:<P>
        <INPUT TYPE="TEXT" NAME="url" SIZE="40" VALUE="">
        `))
	case "rename":
		eBn := html.EscapeString(r.uFbn)
		r.w.Write([]byte(`
        &nbsp;<BR>Enter new name for the file <B>` + eBn + `</B>:<P>
        <INPUT TYPE="TEXT" NAME="dst" SIZE="40" VALUE="` + eBn + `">
        <INPUT TYPE="HIDDEN" NAME="file" VALUE="` + eBn + `">
        `))
	case "move":
		eBn := html.EscapeString(r.uFbn)
		r.w.Write([]byte(`
		&nbsp;<BR>Select destination folder for <B>` + eBn + `</B>:<P>
		<SELECT NAME="dst">
		` + upDnDir(r.uDir, "") + `</SELECT>
		<INPUT TYPE="HIDDEN" NAME="file" VALUE="` + eBn + `">
		`))
	case "delete":
		var a string
		fi, _ := os.Stat(r.uDir + "/" + r.uFbn)
		if fi.IsDir() {
			a = "directory - recursively"
		} else {
			a = "file, size " + humanize.Bytes(uint64(fi.Size()))
		}
		eBn := html.EscapeString(r.uFbn)
		r.w.Write([]byte(`
        &nbsp;<BR>Are you sure you want to delete:<BR><B>` + eBn + `</B>
        (` + a + `)<P>
        <INPUT TYPE="HIDDEN" NAME="file" VALUE="` + eBn + `">
        `))
	case "multi_delete":
		fmt.Fprintf(r.w, "&nbsp;<BR>Are you sure you want to delete from <B>%v</B>:<P><UL>\n", html.EscapeString(r.uDir))
		for _, f := range mul {
			fE := html.EscapeString(f)
			fmt.Fprintf(r.w, "<INPUT TYPE=\"HIDDEN\" NAME=\"mulf\" VALUE=\"%s\">\n"+
				"<LI TYPE=\"square\">%v</LI>\n", fE, fE)
		}
		fmt.Fprintln(r.w, "</UL><P>")
	case "multi_move":
		fmt.Fprintf(r.w, "&nbsp;<BR>Move from: <B>%v</B><P>\n"+
			"To: <SELECT NAME=\"dst\">%v</SELECT><P>\n<UL>Items:<P>\n",
			html.EscapeString(r.uDir),
			upDnDir(r.uDir, r.uFbn),
		)
		for _, f := range mul {
			fE := html.EscapeString(f)
			fmt.Fprintf(r.w, "<INPUT TYPE=\"HIDDEN\" NAME=\"mulf\" VALUE=\"%s\">\n"+
				"<LI TYPE=\"square\">%v</LI>\n", fE, fE)
		}
		fmt.Fprintln(r.w, "</UL><P>")
	}

	r.w.Write([]byte(`
    </TD></TR>
    <TR><TD COLSPAN="2">
    <P><CENTER>
    <INPUT TYPE="SUBMIT" VALUE=" OK " NAME="OK" ` + disTag[r.rwAccess] + `>&nbsp;
    <INPUT TYPE="SUBMIT" VALUE=" Cancel " NAME="cancel">
    <INPUT TYPE="HIDDEN" NAME="fn" VALUE="` + action + `">
    </CENTER>
    </TD></TR><TR><TD COLSPAN="2">&nbsp;</TD></TR>
    </TABLE>
    </TD></TR></TABLE>
    `))

	footer(r.w)
}

func (r *wfmRequest) editText() {
	fi, err := os.Stat(r.uDir + "/" + r.uFbn)
	if err != nil {
		htErr(r.w, "Unable to get file attributes", err)
		return
	}
	if fi.Size() > 1<<20 {
		htErr(r.w, "edit", fmt.Errorf("the file is too large for editing"))
		return
	}
	f, err := ioutil.ReadFile(r.uDir + "/" + r.uFbn)
	if err != nil {
		htErr(r.w, "Unable to read file", err)
		return
	}
	header(r.w, r.uDir, r.eSort, `html, body, table, textarea, form { box-sizing: border-box; height:98%; }`)
	r.w.Write([]byte(`
    <TABLE BGCOLOR="#EEEEEE" BORDER="0" CELLSPACING="0" CELLPADDING="5" STYLE="width: 100%; height: 100%;">
    <TR STYLE="height:1%;">
    <TD ALIGN="LEFT" VALIGN="MIDDLE" BGCOLOR="#CCCCCC">File Editor: ` + html.EscapeString(r.uFbn) + `</TD>
    <TD  BGCOLOR="#CCCCCC" ALIGN="RIGHT">&nbsp;</TD>
    </TR>
    <TR STYLE="height:98%;">
    <TD COLSPAN="2" ALIGN="CENTER" VALIGN="MIDDLE" STYLE="height:100%;">
    <TEXTAREA NAME="text" SPELLCHECK="false" COLS="80" ROWS="24" STYLE="width: 99%; height: 99%;">` + html.EscapeString(string(f)) + `</TEXTAREA><P>
    </TD></TR><TR STYLE="height:1%;"><TD ALIGN="RIGHT">
	<INPUT TYPE="SUBMIT" NAME="save" VALUE="Save" ` + disTag[r.rwAccess] + `>&nbsp;
	<INPUT TYPE="SUBMIT" NAME="cancel" VALUE="Cancel">
    <INPUT TYPE="HIDDEN" NAME="dir" VALUE="` + html.EscapeString(r.uDir) + `">
    <INPUT TYPE="HIDDEN" NAME="file" VALUE="` + html.EscapeString(r.uFbn) + `">
    </TD></TR></TABLE>
    `))
	footer(r.w)
}

func (r *wfmRequest) about(ua string) {
	header(r.w, r.uDir, r.eSort, "")

	r.w.Write([]byte(`
    <TABLE WIDTH="100%" HEIGHT="90%" BORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD VALIGN="MIDDLE" ALIGN="CENTER">
    <BR>&nbsp;<BR><P>
    <TABLE WIDTH="400" BGCOLOR="#F0F0F0" BORDER="0" CELLSPACING="0" CELLPADDING="1" CLASS="tbr">
      <TR><TD COLSPAN="2" BGCOLOR="#004080"><FONT COLOR="#FFFFFF">&nbsp; Web File Manager</FONT></TD></TR>
      <TR><TD WIDTH="30">&nbsp;</TD><TD><BR>
	  WFM Version v` + vers + `<BR>
	  Developed by Antoni Sawicki Et Al.<BR>
	  <A HREF="https://github.com/tenox7/wfm/">https://github.com/tenox7/wfm/</A><BR>
	  Copyright &copy; 1994-2018 by Antoni Sawicki<BR>
	  Copyright &copy; 2018-2022 by Google LLC<BR>
	  This is not an official Google product.<P>
	`))

	if *aboutRnt {
		fmt.Fprintf(r.w, "Build: %v %v-%v<BR>Agent: %v<P>",
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
			ua)
	}

	r.w.Write([]byte(`
      </TD></TR>
    <TR><TD COLSPAN="2">
    <P><CENTER>
    <INPUT TYPE="SUBMIT" VALUE=" OK " NAME="OK">&nbsp;
    </CENTER>
    </TD></TR><TR><TD COLSPAN="2">&nbsp;</TD></TR>
    </TABLE>
    </TD></TR></TABLE>
    `))

	footer(r.w)
}
