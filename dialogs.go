package main

import "net/http"

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
