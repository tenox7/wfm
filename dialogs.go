package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dustin/go-humanize"
)

func prompt(w http.ResponseWriter, uDir, uBaseName, sort, action string) {
	header(w, uDir, sort)

	w.Write([]byte(`
    <TABLE WIDTH="100%" HEIGHT="90%" BORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD VALIGN="MIDDLE" ALIGN="CENTER">
    <BR>&nbsp;<BR><P>
    <TABLE WIDTH="400" BGCOLOR="#F0F0F0" BORDER="0" CELLSPACING="0" CELLPADDING="1" CLASS="tbr">
      <TR><TD COLSPAN="2" BGCOLOR="#004080"><FONT COLOR="#FFFFFF">&nbsp; ` + action + `</FONT></TD></TR>
      <TR><TD WIDTH="30">&nbsp;</TD><TD>
    `))

	switch action {
	case "mkdir":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new directory:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        `))
	case "mkfile":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new file:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        `))
	case "mkurl":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new url file:<P>
        <INPUT TYPE="TEXT" NAME="file" SIZE="40" VALUE="">
        &nbsp;<BR>Destination URL:<P>
        <INPUT TYPE="TEXT" NAME="url" SIZE="40" VALUE="">
        `))
	case "rename":
		eBn := html.EscapeString(uBaseName)
		w.Write([]byte(`
        &nbsp;<BR>Enter new name for the file:<P>
        <INPUT TYPE="TEXT" NAME="newf" SIZE="40" VALUE="` + eBn + `">
        <INPUT TYPE="HIDDEN" NAME="oldf" VALUE="` + eBn + `">
        `))
	case "delete":
		var a string
		fi, _ := os.Stat(uDir + "/" + uBaseName)
		if fi.IsDir() {
			a = "directory - recursively"
		} else {
			a = "file, size " + humanize.Bytes(uint64(fi.Size()))
		}
		eBn := html.EscapeString(uBaseName)
		w.Write([]byte(`
        &nbsp;<BR>Are you sure you want to delete:<BR><B>` + eBn + `</B>
        (` + a + `)<P>
        <INPUT TYPE="HIDDEN" NAME="file" VALUE="` + eBn + `">
        `))
	}

	w.Write([]byte(`
    </TD></TR>
    <TR><TD COLSPAN="2">
    <P><CENTER>
    <INPUT TYPE="SUBMIT" VALUE=" OK " NAME="OK">&nbsp;
    <INPUT TYPE="SUBMIT" VALUE=" Cancel " NAME="cancel">
    <INPUT TYPE="HIDDEN" NAME="fn" VALUE="` + action + `">
    </CENTER>
    </TD></TR><TR><TD COLSPAN="2">&nbsp;</TD></TR>
    </TABLE>
    </TD></TR></TABLE>
    `))

	footer(w)
}

func editText(w http.ResponseWriter, uFilePath, sort string) {
	fi, err := os.Stat(uFilePath)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}
	if fi.Size() > 1<<20 {
		htErr(w, "edit", fmt.Errorf("the file is too large for editing"))
		return
	}
	f, err := ioutil.ReadFile(uFilePath)
	if err != nil {
		htErr(w, "Unable to read file", err)
		return
	}
	header(w, filepath.Dir(uFilePath), sort)
	w.Write([]byte(`
    <TABLE BGCOLOR="#EEEEEE" BORDER="0" CELLSPACING="0" CELLPADDING="5" STYLE="width: 100%; height: 100%;">
    <TR STYLE="height:1%;">
    <TD ALIGN="LEFT" VALIGN="MIDDLE" BGCOLOR="#CCCCCC">File Editor: ` + html.EscapeString(filepath.Base(uFilePath)) + `</TD>
    <TD  BGCOLOR="#CCCCCC" ALIGN="RIGHT"></TD>
    </TR>
    <TR STYLE="height:99%;">
    <TD COLSPAN="2" ALIGN="CENTER" VALIGN="MIDDLE" STYLE="height:100%;">
    <TEXTAREA NAME="text" SPELLCHECK="false" COLS="120" ROWS="24" STYLE="width: 99%; height: 99%;">` + html.EscapeString(string(f)) + `</TEXTAREA><P>
    <INPUT TYPE="SUBMIT" NAME="save" VALUE="Save" STYLE="float: left;">
	<INPUT TYPE="SUBMIT" NAME="cancel" VALUE="Cancel" STYLE="float: left; margin-left: 10px">
    <INPUT TYPE="HIDDEN" NAME="fp" VALUE="` + html.EscapeString(uFilePath) + `">
    </TD></TR></TABLE>
    `))
	footer(w)
}

func about(w http.ResponseWriter, uDir, sort, ua string) {
	header(w, uDir, sort)

	w.Write([]byte(`
    <TABLE WIDTH="100%" HEIGHT="90%" BORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD VALIGN="MIDDLE" ALIGN="CENTER">
    <BR>&nbsp;<BR><P>
    <TABLE WIDTH="400" BGCOLOR="#F0F0F0" BORDER="0" CELLSPACING="0" CELLPADDING="1" CLASS="tbr">
      <TR><TD COLSPAN="2" BGCOLOR="#004080"><FONT COLOR="#FFFFFF">&nbsp; About WFM</FONT></TD></TR>
      <TR><TD WIDTH="30">&nbsp;</TD><TD><BR>
	  WFM Version v` + vers + `<BR>
	  Developed by Antoni Sawicki Et Al.<BR>
	  <A HREF="https://github.com/tenox7/wfm/">https://github.com/tenox7/wfm/</A><BR>
	  Copyright &copy; 1994-2018 by Antoni Sawicki<BR>
	  Copyright &copy; 2018-2022 by Google LLC<BR>
	  This is not an official Google product.<P>
	`))

	if *aboutRnt {
		fmt.Fprintf(w, "Go=%v<BR>OS=%v<BR>ARCH=%v<BR>Agent=%v<P>", runtime.Version(), runtime.GOOS, runtime.GOARCH, ua)
	}

	w.Write([]byte(`
      </TD></TR>
    <TR><TD COLSPAN="2">
    <P><CENTER>
    <INPUT TYPE="SUBMIT" VALUE=" OK " NAME="OK">&nbsp;
    </CENTER>
    </TD></TR><TR><TD COLSPAN="2">&nbsp;</TD></TR>
    </TABLE>
    </TD></TR></TABLE>
    `))

	footer(w)
}
