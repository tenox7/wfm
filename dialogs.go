package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func prompt(w http.ResponseWriter, eDir, sort, action string) {
	header(w, eDir, sort)

	w.Write([]byte(`
    <TABLE WIDTH="100%" HEIGHT="100%" BORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD VALIGN="MIDDLE" ALIGN="CENTER">
    <BR>&nbsp;<BR><P>
    <TABLE WIDTH="400" BGCOLOR="#F0F0F0" BORDER="0" CELLSPACING="0" CELLPADDING="1" CLASS="tbr">
      <TR><TD COLSPAN="2" BGCOLOR="#004080"><FONT COLOR="#FFFFFF">&nbsp; ` + action + `</FONT></TD></TR>
      <TR><TD WIDTH="30">&nbsp;</TD><TD>
    `))

	switch action {
	case "mkdir":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new directory:<P>
        <INPUT TYPE="TEXT" NAME="newd" SIZE="40" VALUE="">
        `))
	case "mkfile":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new file:<P>
        <INPUT TYPE="TEXT" NAME="newf" SIZE="40" VALUE="">
        `))
	case "mkurl":
		w.Write([]byte(`
        &nbsp;<BR>Enter name for the new url file:<P>
        <INPUT TYPE="TEXT" NAME="newu" SIZE="40" VALUE="">
        &nbsp;<BR>Destination URL:<P>
        <INPUT TYPE="TEXT" NAME="url" SIZE="40" VALUE="">
        `))
	}

	w.Write([]byte(`
    </TD></TR>
    <TR><TD COLSPAN="2">
    <P><CENTER>
    <INPUT TYPE="SUBMIT" VALUE=" OK " NAME="OK">
    <INPUT TYPE="SUBMIT" VALUE=" Cancel " NAME="cancel">
    <INPUT TYPE="HIDDEN" NAME="fn" VALUE="` + action + `">
    </CENTER>
    </TD></TR>
    </TABLE>
    </TD></TR></TABLE>
    `))

	footer(w)
}

func editText(w http.ResponseWriter, fp, sort string) {
	fi, err := os.Stat(fp)
	if err != nil {
		htErr(w, "Unable to get file attributes", err)
		return
	}
	if fi.Size() > 5<<20 {
		htErr(w, "edit", fmt.Errorf("the file is too large for editing"))
		return
	}
	f, err := ioutil.ReadFile(fp)
	if err != nil {
		htErr(w, "Unable to read file", err)
		return
	}
	header(w, html.EscapeString(filepath.Dir(fp)), sort)
	w.Write([]byte(`
    <TABLE BGCOLOR="#EEEEEE" BORDER="0" CELLSPACING="0" CELLPADDING="5" STYLE="width: 100%; height: 100%;">
    <TR STYLE="height:1%;">
    <TD ALIGN="LEFT" VALIGN="MIDDLE" BGCOLOR="#CCCCCC">File Editor: ` + html.EscapeString(filepath.Base(fp)) + `</TD>
    <TD  BGCOLOR="#CCCCCC" ALIGN="RIGHT"></TD>
    </TR>
    <TR STYLE="height:99%;">
    <TD COLSPAN="2" ALIGN="CENTER" VALIGN="MIDDLE" STYLE="height:100%;">
    <TEXTAREA NAME="edit" SPELLCHECK="false" COLS="120" ROWS="24" STYLE="width: 99%; height: 99%;">` + html.EscapeString(string(f)) + `</TEXTAREA><P>
    <INPUT TYPE="submit" VALUE="save" STYLE="float: left;">
	<INPUT TYPE="submit" value="cancel" STYLE="float: left; margin-left: 10px">
    </TD></TR></TABLE>
    `))
	footer(w)
}
