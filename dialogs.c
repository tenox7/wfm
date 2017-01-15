// WFM HTML Dialog Routines

#include "wfm.h"

//
// Prompt for delete / move operation
//
void multiprompt_ui(char *m_action) {
    int i;
    int res;
    char **responses; 
    struct stat fileinfo;
    char M_action[64]={0};

    res=cgiFormStringMultiple("multiselect_filename", &responses);

    // pre-check for filenames so, that if there is an error, HTML is not yet out, allowing error dialog to be rendered
    if(res == cgiFormNotFound) {  
        checkfilename(NULL);
    } else {
        for(i=0; responses[i]; i++) 
            checkfilename(responses[i]);
    }

    cgiHeaderContentType("text/html");
    snprintf(M_action, sizeof(M_action), "%c%s Confirmation", toupper(m_action[0]), m_action+1);
    html_title(M_action);

    fprintf(cgiOut, 
        "</HEAD>\n"
        "<!-- Multi Prompt -->\n"
        "<BODY BGCOLOR=\"#FFFFFF\">\n"
        "<TABLE BORDER=0 CELLSPACING=0 CELLPADDING=0 CLASS=\"twh\"><TR><TD VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
        "<FORM NAME=\"wfm\" ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n\n", cgiScriptName);

    fprintf(cgiOut, 
        "<TABLE WIDTH=\"500\" BGCOLOR=\"#F0F0F0\" BORDER=0 CELLSPACING=0 CELLPADDING=1 CLASS=\"tbr\">\n"
        "  <TR><TD COLSPAN=2 BGCOLOR=\"#004080\"><FONT COLOR=\"#FFFFFF\">&nbsp; %s</FONT></TD></TR>\n"
        "  <TR><TD WIDTH=30>&nbsp;</TD><TD>\n"
        "    &nbsp;<BR>\n"
        "    About to %s following items:<P><UL>\n",
         M_action, m_action);


    if(res == cgiFormNotFound) {  
        checkfilename(NULL);
        if(stat(phys_filename, &fileinfo)==0) {
            fprintf(cgiOut, "<INPUT TYPE=\"HIDDEN\" NAME=\"filename\" VALUE=\"%s\">\n", virt_filename);
            fprintf(cgiOut, "<LI TYPE=\"square\"><B>%s</B>", virt_filename);
            if(S_ISDIR(fileinfo.st_mode))
                fprintf(cgiOut, " [directory %s]\n", buprintf(du(phys_filename), FALSE));
            else
                fprintf(cgiOut, " [file %s]\n", buprintf(fileinfo.st_size, FALSE));
        }
    } else {
        for(i=0; responses[i]; i++) {
            checkfilename(responses[i]);
            if(stat(phys_filename, &fileinfo)==0) {
                fprintf(cgiOut, "<INPUT TYPE=\"HIDDEN\" NAME=\"multiselect_filename\" VALUE=\"%s\">\n", virt_filename);
                fprintf(cgiOut, "<LI TYPE=\"square\"><B>%s</B>", virt_filename);
                if(S_ISDIR(fileinfo.st_mode))
                    fprintf(cgiOut, "/ [directory %s]\n", buprintf(du(phys_filename), FALSE)); 
                else
                    fprintf(cgiOut, " [file %s]\n", buprintf(fileinfo.st_size, FALSE));
            }
        }
    }           

    fprintf(cgiOut, "</UL>");

    // move needs a destination...
    if(strcmp(m_action, "move")==0) {
        fprintf(cgiOut, "<P>Source: %s<P>Destination: <SELECT NAME=\"destination\"><OPTION VALUE=\"/\">/ - Root Directory</OPTION>\n", virt_dirname);
        re_dir_ui("/", 1);
        fprintf(cgiOut, "</SELECT>\n<INPUT TYPE=\"HIDDEN\" NAME=\"absdst\" VALUE=\"1\">\n<P>\n");
    }

    fprintf(cgiOut, 
        "   </TD></TR>\n"
        "   <TR><TD COLSPAN=2>\n"
        "    <P><CENTER>\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"action\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"directory\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"token\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"SUBMIT\" VALUE=\" OK \" NAME=\"OK\">\n"
        "    <INPUT TYPE=\"SUBMIT\" VALUE=\" Cancel \" NAME=\"noop\">\n"
        "    </CENTER><BR>\n"
        "   </TD></TR>\n"
        "</TABLE></FORM>\n\n"
        "</TD></TR></TABLE>\n"
        "</BODY>\n</HTML>\n", m_action, virt_dirname, token);

    cgiStringArrayFree(responses);

}

//
// Single Prompt
// Used for rename, mkfile, mkdir
//
void singleprompt_ui(char *m_action) {
    char M_action[64]={0};

    snprintf(M_action, sizeof(M_action), "%c%s", toupper(m_action[0]), m_action+1);

    if(strcmp(m_action, "move")==0) {
        checkfilename(NULL);
        snprintf(M_action, sizeof(M_action), "Rename");
    }

    cgiHeaderContentType("text/html");
    html_title(M_action);

    fprintf(cgiOut, 
        "</HEAD>\n"
        "<!-- Single Prompt -->\n"
        "<BODY %s BGCOLOR=\"#FFFFFF\">\n"
        "<TABLE BORDER=0 CELLSPACING=0 CELLPADDING=0 CLASS=\"twh\"><TR><TD VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
        "<FORM NAME=\"wfm\" ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n\n"
        "<TABLE WIDTH=\"400\" BGCOLOR=\"#F0F0F0\" BORDER=0 CELLSPACING=0 CELLPADDING=1 CLASS=\"tbr\">\n"
        "  <TR><TD COLSPAN=2 BGCOLOR=\"#004080\"><FONT COLOR=\"#FFFFFF\">&nbsp; %s</FONT></TD></TR>\n"
        "  <TR><TD WIDTH=30>&nbsp;</TD><TD>\n"
        "    &nbsp;<BR>\n",
        (js) ? "ONLOAD=\"document.wfm.inp1.focus();\"" : "", cgiScriptName, M_action);


    if(strcmp(m_action, "move")==0)
        fprintf(cgiOut,
        "    Current Name: <B>%s</B><P>\n"
        "    Enter new name:<P>\n"
        "    <INPUT TYPE=\"TEXT\" ID=\"inp1\" NAME=\"destination\" SIZE=\"40\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"filename\" VALUE=\"%s\">\n",
            virt_filename, virt_filename, virt_filename);

    else if(strcmp(m_action, "mkfile")==0)
        fprintf(cgiOut,
        "    Enter name of the new text file:<P>\n"
        "    <INPUT TYPE=\"TEXT\" ID=\"inp1\" NAME=\"filename\" SIZE=\"40\" VALUE=\"\">\n");

    else if(strcmp(m_action, "mkdir")==0)
        fprintf(cgiOut,
        "    &nbsp;<BR>Enter name of the new directory:<P>\n"
        "    <INPUT TYPE=\"TEXT\" ID=\"inp1\" NAME=\"filename\" SIZE=\"40\" VALUE=\"\">\n");

    fprintf(cgiOut,
        "   </TD></TR>\n"
        "   <TR><TD COLSPAN=2>\n"
        "    <P><CENTER>\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"action\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"directory\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"HIDDEN\" NAME=\"token\" VALUE=\"%s\">\n"
        "    <INPUT TYPE=\"SUBMIT\" VALUE=\" OK \" NAME=\"OK\">\n"
        "    <INPUT TYPE=\"SUBMIT\" VALUE=\" Cancel \" NAME=\"noop\">\n"
        "    </CENTER><BR>\n"
        "   </TD></TR>\n"
        "</TABLE></FORM>\n\n"
        "</TD></TR></TABLE>\n"
        "</BODY>\n</HTML>\n", m_action, virt_dirname, token);

}


//
// Error message - note that strerror() is already passed by the caller
//
void error(char *msg, ...) {
    va_list ap;
    char buff[1024]={0};

    if(msg) {
        va_start(ap, msg);
        vsnprintf(buff, sizeof(buff), msg, ap);
        va_end(ap);

        cgiHeaderContentType("text/html");
        html_title("ERROR");
        fprintf(cgiOut, 
            "</HEAD>\n"\
            "<BODY BGCOLOR=\"#FFFFFF\">\n"\
            "<TABLE BORDER=0 CELLSPACING=0 CELLPADDING=0 CLASS=\"twh\"><TR><TD VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
            "<TABLE WIDTH=\"400\" BGCOLOR=\"#D4D0C8\" BORDER=0 CELLSPACING=0 CELLPADDING=1 CLASS=\"tbr\">\n"
            "<TR BGCOLOR=\"#FF0000\">\n"
             "<TD COLSPAN=3 ALIGN=\"LEFT\">\n"
                "<SPAN STYLE=\"color:#FFFFFF; font-weight:bold;\">&nbsp;ERROR:</SPAN>\n"
             "</TD>\n"
            "</TR>\n"
            "<TR BGCOLOR=\"#EEEEEE\">\n"
                "<TD WIDTH=\"50\" VALIGN=\"top\">\n"
                "&nbsp;<BR>\n"
                "</TD>\n"
                "<TD ALIGN=\"LEFT\">\n"
                "&nbsp;<BR>\n"
                "%s<BR>\n"
                "&nbsp;<P>\n"
                "&nbsp;<P>\n"
                "</TD>\n"
                "<TD WIDTH=\"20\">\n"
                "&nbsp;\n"
                "</TD>\n"
            "</TR>\n"
            "<TR><TD COLSPAN=3 ALIGN=\"CENTER\" BGCOLOR=\"#EEEEEE\">\n"
            "<FORM ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n"
            "<INPUT TYPE=\"SUBMIT\" VALUE=\" OK \" NAME=\"noop\">\n"
            "<INPUT TYPE=\"HIDDEN\" NAME=\"directory\" VALUE=\"%s\">\n"
            "<INPUT TYPE=\"HIDDEN\" NAME=\"token\" VALUE=\"%s\">\n"
            "</FORM>\n</TD></TR>\n"
            "<TR><TD COLSPAN=3 BGCOLOR=\"#EEEEEE\">&nbsp;</TD></TR>\n"
            "</TABLE>\n"
            "</TD></TR></TABLE>\n</BODY></HTML>\n",        
        buff, cgiScriptName, virt_dirname, token);
    }
    else {
        cgiHeaderContentType("text/plain");
        fprintf(cgiOut, "FATAL ERROR\n");
    }

    exit(0);
}


//
// About message 
//
void about(void) {
    cgiHeaderContentType("text/html");
    html_title("About");
    fprintf(cgiOut,
        "</HEAD>\n"
        "<BODY BGCOLOR=\"#FFFFFF\">\n"
        "<TABLE BORDER=0 CELLSPACING=0 CELLPADDING=0 CLASS=\"twh\"><TR><TD VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
        "<TABLE WIDTH=\"500\" BGCOLOR=\"#D4D0C8\" BORDER=0 CELLSPACING=0 CELLPADDING=1 CLASS=\"tbr\">\n"
        "<TR BGCOLOR=\"#0040A0\">\n"
         "<TD COLSPAN=3 ALIGN=\"LEFT\">\n"
            "<IMG SRC=\"%swfmicon.gif\" ALT=\"wfm Icon\"><SPAN STYLE=\"color:#FFFFFF\"> About:</SPAN>\n"
         "</TD>\n"
        "</TR>\n"
        "<TR BGCOLOR=\"#EEEEEE\">\n"
            "<TD WIDTH=\"50\" VALIGN=\"top\">\n"
            "&nbsp;<BR>\n"
            "</TD>\n"
            "<TD ALIGN=\"LEFT\">\n"
            "&nbsp;<BR>\n"
            "<U>%s</U><BR>\n"
            "WFM Implemented by Antoni Sawicki<BR>\n"
            "CGIC Library by Thomas Boutell<BR>\n"
            "Server Side RFC 1321 implementation by L. Peter Deutsch<BR>\n"
            "Client Side RFC 1321 implementation by Paul Johnston<BR>\n"
            "Icons by Yusuke Kamiyamane<BR>\n"
#ifdef WFMGIT
            "Uses libgit2 library<BR>\n"
#endif
            "URL Encoding routines by Fred Bulback<BR>\n"
            "Copyright &copy; 1994-2017 by Antoni Sawicki<BR>\n"
            "Copyright &copy; 1996-2011 by Thomas Boutell and Boutell.Com, Inc.<BR>\n"
            "Copyright &copy; 2002 by Aladdin Enterprises<BR>\n"
            "Copyright &copy; 1999-2009 by Paul Johnston<BR>\n"
            "Copyright &copy; 2010 by Yusuke Kamiyamane<BR>\n"
            "<HR>\n"
            "WFM: %s (build %s / %s)<BR>\n"
            "GCC: %s<BR>\n"
            "Server: %s<BR>\n"
            "User Agent: %s<BR>\n"
            "NAME_MAX: %d<BR>\n"
            "JavaScript Level: %d<BR>\n"
            "Git Support: %s (%s)<BR>\n"
            "&nbsp;<P>\n"
            "&nbsp;<P>\n"
            "</TD>\n"
            "<TD WIDTH=\"20\">\n"
            "&nbsp;\n"
            "</TD>\n"
        "</TR>\n"
        "<TR><TD COLSPAN=3 ALIGN=\"CENTER\" BGCOLOR=\"#EEEEEE\">\n"
        "<FORM ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n"
        "<INPUT TYPE=\"SUBMIT\" VALUE=\" OK \" NAME=\"noop\">\n"
        "<INPUT TYPE=\"HIDDEN\" NAME=\"directory\" VALUE=\"%s\">\n"
        "<INPUT TYPE=\"HIDDEN\" NAME=\"token\" VALUE=\"%s\">\n"
        "</FORM>\n</TD></TR>\n"
        "<TR><TD COLSPAN=3 BGCOLOR=\"#EEEEEE\">&nbsp;</TD></TR>\n"
        "</TABLE>\n"
        "</TD></TR></TABLE>\n</BODY></HTML>\n",
        ICONSURL, TAGLINE, VERSION, __DATE__, __TIME__, __VERSION__, cgiServerSoftware, cgiUserAgent, NAME_MAX, js,
#ifdef WFMGIT
        "Yes"
#else
        "No"
#endif
        , (repo_check()) ? "Oo Repo Present" : "Repo OK",    
        cgiScriptName, virt_dirname, token);

}


//
// Prompt for username and password
//
void login_ui(void) {
    cgiHeaderContentType("text/html");
    html_title("Login");

    if(js>=2) fputs( 
        "<SCRIPT LANGUAGE=\"JavaScript\" TYPE=\"text/javascript\">\n<!--\n"
        "var hexcase=0;function hex_md5(a){return rstr2hex(rstr_md5(str2rstr_utf8(a)))}function hex_hmac_md5(a,b){return rstr2hex(rstr_hmac_md5(str2rstr_utf8(a),str2rstr_utf8(b)))}function md5_vm_test(){return hex_md5(\"abc\").toLowerCase()==\"900150983cd24fb0d6963f7d28e17f72\"}function rstr_md5(a){return binl2rstr(binl_md5(rstr2binl(a),a.length*8))}function rstr_hmac_md5(c,f){var e=rstr2binl(c);if(e.length>16){e=binl_md5(e,c.length*8)}var a=Array(16),d=Array(16);for(var b=0;b<16;b++){a[b]=e[b]^909522486;d[b]=e[b]^1549556828}var g=binl_md5(a.concat(rstr2binl(f)),512+f.length*8);return binl2rstr(binl_md5(d.concat(g),512+128))}function rstr2hex(c){try{hexcase}catch(g){hexcase=0}var f=hexcase?\"0123456789ABCDEF\":\"0123456789abcdef\";var b=\"\";var a;for(var d=0;d<c.length;d++){a=c.charCodeAt(d);b+=f.charAt((a>>>4)&15)+f.charAt(a&15)}return b}function str2rstr_utf8(c){var b=\"\";var d=-1;var a,e;while(++d<c.length){a=c.charCodeAt(d);e=d+1<c.length?c.charCodeAt(d+1):0;if(55296<=a&&a<=56319&&56320<=e&&e<=57343){a=65536+((a&1023)<<10)+(e&1023);d++}if(a<=127){b+=String.fromCharCode(a)}else{if(a<=2047){b+=String.fromCharCode(192|((a>>>6)&31),128|(a&63))}else{if(a<=65535){b+=String.fromCharCode(224|((a>>>12)&15),128|((a>>>6)&63),128|(a&63))}else{if(a<=2097151){b+=String.fromCharCode(240|((a>>>18)&7),128|((a>>>12)&63),128|((a>>>6)&63),128|(a&63))}}}}}return b}function rstr2binl(b){var a=Array(b.length>>2);for(var c=0;c<a.length;c++){a[c]=0}for(var c=0;c<b.length*8;c+=8){a[c>>5]|=(b.charCodeAt(c/8)&255)<<(c%%32)}return a}function binl2rstr(b){var a=\"\";for(var c=0;c<b.length*32;c+=8){a+=String.fromCharCode((b[c>>5]>>>(c%%32))&255)}return a}function binl_md5(p,k){p[k>>5]|=128<<((k)%%32);p[(((k+64)>>>9)<<4)+14]=k;var o=1732584193;var n=-271733879;var m=-1732584194;var l=271733878;for(var g=0;g<p.length;g+=16){var j=o;var h=n;var f=m;var e=l;o=md5_ff(o,n,m,l,p[g+0],7,-680876936);l=md5_ff(l,o,n,m,p[g+1],12,-389564586);m=md5_ff(m,l,o,n,p[g+2],17,606105819);n=md5_ff(n,m,l,o,p[g+3],22,-1044525330);o=md5_ff(o,n,m,l,p[g+4],7,-176418897);l=md5_ff(l,o,n,m,p[g+5],12,1200080426);m=md5_ff(m,l,o,n,p[g+6],17,-1473231341);n=md5_ff(n,m,l,o,p[g+7],22,-45705983);o=md5_ff(o,n,m,l,p[g+8],7,1770035416);l=md5_ff(l,o,n,m,p[g+9],12,-1958414417);m=md5_ff(m,l,o,n,p[g+10],17,-42063);n=md5_ff(n,m,l,o,p[g+11],22,-1990404162);o=md5_ff(o,n,m,l,p[g+12],7,1804603682);l=md5_ff(l,o,n,m,p[g+13],12,-40341101);m=md5_ff(m,l,o,n,p[g+14],17,-1502002290);n=md5_ff(n,m,l,o,p[g+15],22,1236535329);o=md5_gg(o,n,m,l,p[g+1],5,-165796510);l=md5_gg(l,o,n,m,p[g+6],9,-1069501632);m=md5_gg(m,l,o,n,p[g+11],14,643717713);n=md5_gg(n,m,l,o,p[g+0],20,-373897302);o=md5_gg(o,n,m,l,p[g+5],5,-701558691);l=md5_gg(l,o,n,m,p[g+10],9,38016083);m=md5_gg(m,l,o,n,p[g+15],14,-660478335);n=md5_gg(n,m,l,o,p[g+4],20,-405537848);o=md5_gg(o,n,m,l,p[g+9],5,568446438);l=md5_gg(l,o,n,m,p[g+14],9,-1019803690);m=md5_gg(m,l,o,n,p[g+3],14,-187363961);n=md5_gg(n,m,l,o,p[g+8],20,1163531501);o=md5_gg(o,n,m,l,p[g+13],5,-1444681467);l=md5_gg(l,o,n,m,p[g+2],9,-51403784);m=md5_gg(m,l,o,n,p[g+7],14,1735328473);n=md5_gg(n,m,l,o,p[g+12],20,-1926607734);o=md5_hh(o,n,m,l,p[g+5],4,-378558);l=md5_hh(l,o,n,m,p[g+8],11,-2022574463);m=md5_hh(m,l,o,n,p[g+11],16,1839030562);n=md5_hh(n,m,l,o,p[g+14],23,-35309556);o=md5_hh(o,n,m,l,p[g+1],4,-1530992060);l=md5_hh(l,o,n,m,p[g+4],11,1272893353);m=md5_hh(m,l,o,n,p[g+7],16,-155497632);n=md5_hh(n,m,l,o,p[g+10],23,-1094730640);o=md5_hh(o,n,m,l,p[g+13],4,681279174);l=md5_hh(l,o,n,m,p[g+0],11,-358537222);m=md5_hh(m,l,o,n,p[g+3],16,-722521979);n=md5_hh(n,m,l,o,p[g+6],23,76029189);o=md5_hh(o,n,m,l,p[g+9],4,-640364487);l=md5_hh(l,o,n,m,p[g+12],11,-421815835);m=md5_hh(m,l,o,n,p[g+15],16,530742520);n=md5_hh(n,m,l,o,p[g+2],23,-995338651);o=md5_ii(o,n,m,l,p[g+0],6,-198630844);l=md5_ii(l,o,n,m,p[g+7],10,1126891415);m=md5_ii(m,l,o,n,p[g+14],15,-1416354905);n=md5_ii(n,m,l,o,p[g+5],21,-57434055);o=md5_ii(o,n,m,l,p[g+12],6,1700485571);l=md5_ii(l,o,n,m,p[g+3],10,-1894986606);m=md5_ii(m,l,o,n,p[g+10],15,-1051523);n=md5_ii(n,m,l,o,p[g+1],21,-2054922799);o=md5_ii(o,n,m,l,p[g+8],6,1873313359);l=md5_ii(l,o,n,m,p[g+15],10,-30611744);m=md5_ii(m,l,o,n,p[g+6],15,-1560198380);n=md5_ii(n,m,l,o,p[g+13],21,1309151649);o=md5_ii(o,n,m,l,p[g+4],6,-145523070);l=md5_ii(l,o,n,m,p[g+11],10,-1120210379);m=md5_ii(m,l,o,n,p[g+2],15,718787259);n=md5_ii(n,m,l,o,p[g+9],21,-343485551);o=safe_add(o,j);n=safe_add(n,h);m=safe_add(m,f);l=safe_add(l,e)}return Array(o,n,m,l)}function md5_cmn(h,e,d,c,g,f){return safe_add(bit_rol(safe_add(safe_add(e,h),safe_add(c,f)),g),d)}function md5_ff(g,f,k,j,e,i,h){return md5_cmn((f&k)|((~f)&j),g,f,e,i,h)}function md5_gg(g,f,k,j,e,i,h){return md5_cmn((f&j)|(k&(~j)),g,f,e,i,h)}function md5_hh(g,f,k,j,e,i,h){return md5_cmn(f^k^j,g,f,e,i,h)}function md5_ii(g,f,k,j,e,i,h){return md5_cmn(k^(f|(~j)),g,f,e,i,h)}function safe_add(a,d){var c=(a&65535)+(d&65535);var b=(a>>16)+(d>>16)+(c>>16);return(b<<16)|(c&65535)}function bit_rol(a,b){return(a<<b)|(a>>>(32-b))};"
        "\n//-->\n</SCRIPT>\n", cgiOut);

    fputs("</HEAD>\n", cgiOut);

    if(js>=2) 
        fputs("<BODY ONLOAD=\"document.wfm.username.focus(); document.wfm.Login.value='MD5 Login';\" BGCOLOR=\"#FFFFFF\">\n", cgiOut);
    else 
        fputs("<BODY BGCOLOR=\"#FFFFFF\">\n", cgiOut);

    fprintf(cgiOut,
        "<TABLE BORDER=0 CELLSPACING=0 CELLPADDING=0 CLASS=\"twh\"><TR><TD VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
        "<FORM NAME=\"wfm\" ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\">\n"
        "<TABLE WIDTH=\"400\" BGCOLOR=\"#F0F0F0\" BORDER=0 CELLSPACING=0 CELLPADDING=1 CLASS=\"tbr\">\n"
        "  <TR><TD COLSPAN=2 BGCOLOR=\"#004080\"><FONT COLOR=\"#FFFFFF\">&nbsp; Authentication Required</FONT></TD></TR>\n"
        "  <TR><TD WIDTH=30>&nbsp;</TD><TD>\n"
        "    &nbsp;<BR>Enter your login credentials:<P>\n"
        "    <TABLE>\n"
        "      <TR><TD>Username:</TD><TD><INPUT TYPE=\"TEXT\" NAME=\"username\" SIZE=\"32\" VALUE=\"\"></TD></TR>\n"
        "      <TR><TD>Password:</TD><TD><INPUT TYPE=\"PASSWORD\" NAME=\"password\" SIZE=\"32\" VALUE=\"\"></TD></TR>\n"
        "    </TABLE><P>\n"
        "   </TD></TR>\n"
        "   <TR><TD COLSPAN=2>\n"
        "    <P><CENTER>\n"
        "    <INPUT TYPE=\"HIDDEN\" VALUE=\"login\" NAME=\"action\">\n"
        "    <INPUT TYPE=\"HIDDEN\" VALUE=\"%s\" NAME=\"directory\">\n"
        "    <INPUT TYPE=\"SUBMIT\" VALUE=\" %s Login \" NAME=\"Login\" ",
        cgiScriptName, virt_dirname, (getenv("HTTPS")) ? "SSL" : "Plaintext");

    if(js>=2) fprintf(cgiOut,
        "onClick=\"self.location='%s?directory=%s&amp;login=client&amp;token=' + hex_md5('%s:' + document.wfm.username.value + ':' + document.wfm.password.value); return false;\"",
        cgiScriptName, virt_dirname_urlencoded, cgiRemoteAddr);

    fputs(
        ">\n"
        "    </CENTER><BR>\n"
        "   </TD></TR>\n"
        "</TABLE></FORM>\n\n"
        "</TD></TR></TABLE>\n"
        "</BODY>\n</HTML>\n", cgiOut);

}


//
// Text Area File Editor
//
void edit_ui(void) {
    FILE *input;
    char *buff;
    char backup[4]={0};
    char *bkcolor;
    int size;

    checkfilename(NULL);

    cgiFormString("backup", backup, sizeof(backup));

    if(strcmp("yes", backup)==0) 
        bkcolor="background-color:#404040; color:#FFFFFF;";
    else
        bkcolor="background-color:#EEEEEE; color:#000000;";

    input=fopen(phys_filename, "r");
    if(input==NULL) 
        error("Unable to open file.<BR>%s", strerror(errno));

    fseek(input, 0, SEEK_END);
    size=ftell(input);
    fseek(input, 0, SEEK_SET);

    if(size>=5*1024*1024)
        error("The file is too large for online editing.<BR>");

    buff=(char *) malloc(size+1);
    if(buff==NULL)
        error("Unable to allocate memory.");
        
    memset(buff, 0, size+1);

    fread(buff, size, 1, input);
    fclose(input);

    cgiHeaderContentType("text/html");
    html_title("Editor");

    if (js) fprintf(cgiOut, 
    "<SCRIPT LANGUAGE=\"JavaScript\" TYPE=\"text/javascript\">\n"
        "<!--\n"
        "function chwrap() {               \n"
        "    if(document.EDITOR.content.wrap=='off') {                \n"
        "        document.EDITOR.content.wrap='soft';\n"
        "        document.EDITOR.wrapbtn.style.backgroundColor='#404040';\n"
        "        document.EDITOR.wrapbtn.style.color='#FFFFFF';\n"
        "    } else {     \n"
        "        document.EDITOR.content.wrap='off';\n"
        "        document.EDITOR.wrapbtn.style.backgroundColor='#EEEEEE';\n"
        "        document.EDITOR.wrapbtn.style.color='#000000';\n"
        "    }    \n"
        "}     \n"
        "function chbak() {               \n"
        "    if(document.EDITOR.backup.value=='yes') {                \n"
        "        document.EDITOR.backup.value='no';                \n"
        "        document.EDITOR.bakbtn.style.backgroundColor='#EEEEEE';\n"
        "        document.EDITOR.bakbtn.style.color='#000000';\n"
        "    } else {     \n"
        "        document.EDITOR.backup.value='yes';               \n"
        "        document.EDITOR.bakbtn.style.backgroundColor='#404040';\n"
        "        document.EDITOR.bakbtn.style.color='#FFFFFF';\n"
        "    } \n"
        "}     \n"
        "//-->\n"
    "</SCRIPT>\n");

    fprintf(cgiOut,
    "<STYLE TYPE=\"text/css\"><!-- \n"
        "html, body, table, textarea, form { box-sizing: border-box; width:100%%; height:100%%; margin:0px; padding:0px; } \n"
    "--></STYLE>\n"
    "</HEAD>\n"
    "<BODY>\n"
    "<FORM NAME=\"EDITOR\" ACTION=\"%s?action=edit_save\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\" >\n"
        "<TABLE BGCOLOR=\"#EEEEEE\" BORDER=0 CELLSPACING=0 CELLPADDING=5 STYLE=\"height:%s%%;\">\n"
         "<TR STYLE=\"height:1%%;\">\n"
            "<TD ALIGN=\"LEFT\" VALIGN=\"MIDDLE\" BGCOLOR=\"#CCCCCC\">\n"
            "<IMG SRC=\"%sedit.gif\" BORDER=0 ALIGN=\"MIDDLE\" ALT=\"EDIT\">\n"
            "File Editor: %s\n"
            "</TD>\n"
            "<TD  BGCOLOR=\"#CCCCCC\" ALIGN=\"RIGHT\">",
            cgiScriptName, (strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 6", 31)==0) ? "80" : "100", ICONSURL, virt_filename);

    if(js) fprintf(cgiOut, 
                "<INPUT TYPE=\"button\" ID=\"bakbtn\" onClick=\"chbak()\" VALUE=\"Backup\" STYLE=\"border:none; %s \"> \n"
                "<INPUT TYPE=\"button\" ID=\"wrapbtn\" onClick=\"chwrap()\" VALUE=\"Wrap\" STYLE=\"border:none; background-color:#404040; color:#FFFFFF;\">\n",
                 bkcolor);

    fprintf(cgiOut,
            "</TD>\n"
         "</TR>\n"
         "<TR STYLE=\"height:99%%;\">\n"
            "<TD COLSPAN=2 ALIGN=\"CENTER\" VALIGN=\"MIDDLE\" STYLE=\"height:100%%;\">\n"
                 "<TEXTAREA COLS=\"120\" ROWS=\"24\" NAME=\"content\" STYLE=\"resize:none;\">");


    cgiHtmlEscapeData(buff, size);

    fprintf(cgiOut, "</TEXTAREA>\n"
            "</TD>\n"
         "</TR>\n"
         "<TR>\n"
            "<TD COLSPAN=2 ALIGN=\"RIGHT\" VALIGN=\"MIDDLE\">\n"
            "<INPUT TYPE=\"SUBMIT\" VALUE=\"Save\" >\n"
            "<INPUT TYPE=\"SUBMIT\" VALUE=\"Cancel\" NAME=\"noop\">\n"
            "</TD>\n"
         "</TR>\n"
        "</TABLE>\n"
    "<INPUT TYPE=\"hidden\" NAME=\"action\" VALUE=\"edit_save\">\n"
    "<INPUT TYPE=\"hidden\" NAME=\"filename\" VALUE=\"%s\">\n"
    "<INPUT TYPE=\"hidden\" NAME=\"directory\" VALUE=\"%s\">\n"
    "<INPUT TYPE=\"hidden\" NAME=\"token\" VALUE=\"%s\">\n"
    "<INPUT TYPE=\"hidden\" NAME=\"backup\" VALUE=\"%s\">\n"
    "</FORM></BODY></HTML>\n",
    virt_filename,  virt_dirname, token, backup);

    free(buff);

}