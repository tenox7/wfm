#include "wfm.h"

char DIRIMG[256], AUPIMG[256], ADNIMG[256], GENIMG[256], NEWIMG[256], ZIPIMG[256];
char IMGIMG[256], OFFIMG[256], PDFIMG[256];
char TXTIMG[256], EXEIMG[256], MEDIMG[257], ISOIMG[256], LNKIMG[256];
regex_t reg_zip, reg_img, reg_pdf, reg_exe, reg_txt, reg_off, reg_med, reg_iso;

char M_HR[]="<FONT COLOR=\"#000000\" STYLE=\"font-weight:bold;\">(Last Hour)";
char M_DAY[]="<FONT COLOR=\"#505050\" STYLE=\"font-weight:bold;\">(Last Day)";
char M_WK[]="<FONT COLOR=\"#000000\">(Last Week)";
//char M_2WK[]="<FONT COLOR=\"#000000\">(Last 2 Weeks)";
char M_MO[]="<FONT COLOR=\"#505050\">(Last Month)";
//char M_2MO[]="<FONT COLOR=\"#505050\">(Last 2 Months)";
//char M_6MO[]="<FONT COLOR=\"#707070\">(Last 6 Months)";
char M_YR[]="<FONT COLOR=\"#909090\">(Last Year)";
char M_OLD[]="<FONT COLOR=\"#C0C0C0\">(Old)";

char tNORMAL_COLOR[]="FFFFFF";
char tHIGH_COLOR[]="33CC33";
char tHL_COLOR[]="FFD700";

static const char *access_string[]={ "none", "readonly", "readwrite" };

void dir_icoinita(void) {
    snprintf(DIRIMG, sizeof(DIRIMG), "<IMG SRC=\"%sdir.gif\" ALT=\"Dir\" ALIGN=\"MIDDLE\" BORDER=\"0\">", ICONSURL);
    snprintf(LNKIMG, sizeof(LNKIMG), "<IMG SRC=\"%slnk.gif\" ALT=\"Symlink\" ALIGN=\"MIDDLE\" BORDER=\"0\">", ICONSURL);
    snprintf(AUPIMG, sizeof(AUPIMG), "<IMG SRC=\"%saup.gif\" ALT=\"Up\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"7\" HEIGHT=\"4\">", ICONSURL);
    snprintf(ADNIMG, sizeof(ADNIMG), "<IMG SRC=\"%sadn.gif\" ALT=\"Down\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"7\" HEIGHT=\"4\">", ICONSURL);
    snprintf(GENIMG, sizeof(GENIMG), "<IMG SRC=\"%sgen.gif\" ALT=\"Unknown\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(NEWIMG, sizeof(NEWIMG), "<IMG SRC=\"%sarr.gif\" ALT=\"New\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(ZIPIMG, sizeof(ZIPIMG), "<IMG SRC=\"%szip.gif\" ALT=\"Archive\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(IMGIMG, sizeof(IMGIMG), "<IMG SRC=\"%simg.gif\" ALT=\"Image\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(OFFIMG, sizeof(OFFIMG), "<IMG SRC=\"%soff.gif\" ALT=\"Office File\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(PDFIMG, sizeof(PDFIMG), "<IMG SRC=\"%spdf.gif\" ALT=\"PDF\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(TXTIMG, sizeof(TXTIMG), "<IMG SRC=\"%stxt.gif\" ALT=\"Text\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(EXEIMG, sizeof(EXEIMG), "<IMG SRC=\"%sexe.gif\" ALT=\"Exec\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(MEDIMG, sizeof(MEDIMG), "<IMG SRC=\"%smed.gif\" ALT=\"Multimedia\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);
    snprintf(ISOIMG, sizeof(ISOIMG), "<IMG SRC=\"%siso.gif\" ALT=\"Disk Image\" ALIGN=\"MIDDLE\" BORDER=\"0\" WIDTH=\"16\" HEIGHT=\"16\">", ICONSURL);

    if(
        regcomp(&reg_zip, "\\.(zip|rar|tar|gz|tgz|z|arj|bz|tbz|7z|xz)$",            REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_img, "\\.(gif|png|tif|tiff|jpg|jpeg)$",                        REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_off, "\\.(doc|docx|rtf|dot|xls|xlsx|ppt|pptx|off)$",           REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_iso, "\\.(iso|flp|img|nrg|dmg)$",                              REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_med, "\\.(mp3|mp4|vaw|mov|avi|ivr|mkv)$",                      REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_pdf, "\\.(pdf|ps|eps|ai)$",                                    REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_txt, "\\.(txt|asc|nfo|me|log|htm|html|shtml|js|jsp|php|xml|dtd|css|bas|c|h|cpp|cmd|bat|sh|ksh|awk|reg|log|bak|cfg|py)$", REG_EXTENDED | REG_ICASE)!=0 ||
        regcomp(&reg_exe, "\\.(exe|cmd|vbs|bat|com|pif)$",                          REG_EXTENDED | REG_ICASE)!=0 
    ) error("Unable to compile regex.");

}


//
// Display directory list main panel
//
void dirlist(void) {
    ASDIR *direntry;
    off_t size, totalsize=0;
    char highlight[VIRT_FILENAME_SIZE];
    char namepfx[1024], sizepfx[1024], datepfx[1024];
    char rtime[64], mtime[64], atime[64];
    char *stime;
    char sortby[64];
    char *name, *icon, *linecolor;
	int nentr=0, e;
    int editable;
    time_t now;

    time(&now);

    cgiFormStringNoNewlines("highlight", highlight, VIRT_FILENAME_SIZE-1);
    cgiFormStringNoNewlines("sortby", sortby, 63);
    if(strlen(sortby)<4)
        snprintf(sortby, 63, "name");

    //
    // Get and Print Directory Entries
    //
         if(strcmp(sortby, "name")==0)        nentr=asscandir(phys_dirname, &direntry, namesort);
    else if(strcmp(sortby, "rname")==0)       nentr=asscandir(phys_dirname, &direntry, rnamesort);
    else if(strcmp(sortby, "size")==0)        nentr=asscandir(phys_dirname, &direntry, sizesort);
    else if(strcmp(sortby, "rsize")==0)       nentr=asscandir(phys_dirname, &direntry, rsizesort);
    else if(strcmp(sortby, "date")==0)        nentr=asscandir(phys_dirname, &direntry, timesort);
    else if(strcmp(sortby, "rdate")==0)       nentr=asscandir(phys_dirname, &direntry, rtimesort);
    else                                      nentr=asscandir(phys_dirname, &direntry, namesort);

    dir_icoinita();

    cgiHeaderContentType("text/html");

    //
    // HTML HEADER
    //
    fprintf(cgiOut,
        "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n\"http://www.w3.org/TR/html4/loose.dtd\">\n"
        "\n%s" 
        "<HTML>"
        "<HEAD>\n"
        "<TITLE>%s : %c%s</TITLE>\n"
        "<SCRIPT LANGUAGE=\"JavaScript\" TYPE=\"text/javascript\">\n"
        "<!--\n"
        "    function checkUncheckAll(checkAllState, cbGroup) {\n"
        "        if(cbGroup.length > 0) {\n"
        "            for (i = 0; i < cbGroup.length; i++) {\n"
        "                cbGroup[i].checked = checkAllState.checked;\n"
        "            }\n"
        "        }\n"
        "        else {\n"
        "            cbGroup.checked = checkAllState.checked;\n"
        "        }\n"
        "    }\n"
        "\n"
        "   function xmlhttpPost(strURL) {\n"
        "       var xmlHttpReq = false;\n"
        "       var self = this;\n"
        "       var method = \"GET\";\n"
        "\n"
        "      if (window.XMLHttpRequest) {\n"
        "         self.xmlHttpReq = new XMLHttpRequest();\n"
        "      }\n"
        "      else if (window.ActiveXObject) {\n"
        "         self.xmlHttpReq = new ActiveXObject(\"Microsoft.XMLHTTP\");\n"
        "   }\n"
        "   if (self.xmlHttpReq != null) {\n"
        "         self.xmlHttpReq.open(method, strURL, true);\n"
        "         self.xmlHttpReq.onreadystatechange = function () {\n"
        "            if (self.xmlHttpReq.readyState == 4) {\n"
        "               var result = document.getElementById(\"Upload_Status\");\n"
        "               result.value = self.xmlHttpReq.responseText;\n"
        "            }\n"
        "         }\n"
        "           self.xmlHttpReq.send(null);\n"
        "      }\n"
        "   }\n"
        "\n"
        "   function start() {\n"
        "      setInterval('xmlhttpPost(\"%s?ea=upstat&upload_id=%s\");', \"100\");\n"
        "   }\n"
        "\n"
        "//-->\n"
        "</SCRIPT>\n"
        "<STYLE TYPE=\"text/css\">\n"
        "<!--\n"
        "A:link {text-decoration: none; color:#0000CE; } \n"
        "A:visited {text-decoration: none; color:#0000CE; } \n"
        "A:active {text-decoration: none; color:#FF0000; } \n"
        "A:hover {text-decoration: none; color:#FF0000; } \n"
        "html, body, table { width:100%%; margin:0px; padding:0px; border:none; } \n"
        "img { vertical-align:middle; } \n"
        "td, th { font-family: Tahoma, Arial, Geneva, sans-serif; font-size:11px; margin:0px; padding:2px; border:none;  } \n"
        "input { border-color:#000000;border-style:solid; \n"
        "font-family: Tahoma, Arial, Geneva, sans-serif; font-size:11px; }\n"
        ".hovout { border: none; padding: 0px; background-color: transparent; color: #0000CE; }\n"
        ".hovin  { border: none; padding: 0px; background-color: transparent; color: #FF0000; }\n"
        "//-->\n"
        "</STYLE>\n" 
        "<META HTTP-EQUIV=\"Content-type\" CONTENT=\"text/html;charset=UTF-8\">\n"
        "<LINK REL=\"icon\" TYPE=\"image/gif\" HREF=\"%s%s\">\n"
        "</HEAD>\n"
        "<BODY BGCOLOR=\"#FFFFFF\">\n"
        "<FORM ACTION=\"%s\" METHOD=\"POST\" ENCTYPE=\"multipart/form-data\" onsubmit=\"start()\">\n",
    copyright, TAGLINE, (strlen(virt_dirname)>0) ? ' ' : '/', virt_dirname, cgiScriptName, "1234",  ICONSURL, FAVICON, cgiScriptName);




    //
    // TITLE
    //
    fprintf(cgiOut, 
            "<!-- TITLE --> \n"
            "<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"
                "<TR>\n"
                "<TD WIDTH=\"100%%\" BGCOLOR=\"#0072c6\" VALIGN=\"MIDDLE\" ALIGN=\"LEFT\" STYLE=\"color:#FFFFFF; font-weight:bold;\">\n"
                "&nbsp;<IMG SRC=\"%s%s\" ALIGN=\"MIDDLE\" ALT=\"WFM\">\n"
                "%s : %c%s \n"
                "<TD BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"RIGHT\" STYLE=\"color:#000000; font-weight:bold;  white-space:nowrap\">\n"
                ,
            ICONSURL, FAVICON, TAGLINE, (strlen(virt_dirname)>0) ? ' ' : '/', virt_dirname 
    );


    // lock / unlock
    if(!access_as_user && users_defined)
        fprintf(cgiOut, 
            "<A HREF=\"%s?action=login&amp;directory=%s\">"\
            "&nbsp;<IMG SRC=\"%s%s.gif\" ALIGN=\"MIDDLE\" BORDER=\"0\" ALT=\"Access\"></A>&nbsp;%s\n",  
            cgiScriptName, virt_dirname, ICONSURL, access_string[access_level], access_string[access_level]);
    else
        fprintf(cgiOut, 
            "<A HREF=\"%s?directory=%s\"><IMG SRC=\"%s%s.gif\" BORDER=\"0\" ALIGN=\"MIDDLE\" ALT=\"Access\">"
            "</A>&nbsp;%s&nbsp;<IMG SRC=\"%suser.gif\" ALIGN=\"MIDDLE\" ALT=\"User\">&nbsp;%s&nbsp;\n",
            cgiScriptName, virt_dirname, ICONSURL, access_string[access_level], access_string[access_level], ICONSURL, loggedinuser);

    // about / version
    fprintf(cgiOut, 
            "</TD><TD BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"RIGHT\" STYLE=\"color:#000000; font-weight:bold;  white-space:nowrap\">"
            "&nbsp;<IMG SRC=\"%snet.gif\" ALIGN=\"MIDDLE\" ALT=\"Client IP\">&nbsp;%s&nbsp;</TD>"
            "<TD BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"RIGHT\" STYLE=\"color:#000000; font-weight:bold;  white-space:nowrap\">"
            "<A HREF=\"%s?action=about&amp;directory=%s&amp;token=%s\"><IMG BORDER=\"0\" SRC=\"%sver.gif\" ALIGN=\"MIDDLE\" ALT=\"Version\"></A>&nbsp;v%s&nbsp;"
            "</TD>\n"\
            "</TR>\n"\
            "</TABLE>\n",
            ICONSURL, cgiRemoteAddr, cgiScriptName, virt_dirname, token, ICONSURL, VERSION
    );



    //
    // TOOLBAR
    //
    fprintf(cgiOut, 
            "<!-- TOOLBAR -->\n"\
            "<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=\"0\" CELLSPACING=\"0\" BORDER=\"0\" STYLE=\"height:28px;\">\n"\
                "<TR>\n"\
                "<!-- DIR-UP -->\n"\
                "<TD BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                 "<A HREF=\"%s?sortby=%s&amp;directory=%s&amp;token=%s\">\n"\
                    "<IMG SRC=\"%sdir_up.gif\" BORDER=0 ALIGN=\"MIDDLE\" WIDTH=\"16\" HEIGHT=\"16\" ALT=\"Dir Up\">&nbsp;Up\n"\
                 "</A>\n"\
                "</TD>\n",
                cgiScriptName, sortby, virt_parent, token, ICONSURL);

    fprintf(cgiOut,                 
                "<!-- HOME -->\n"\
                "<TD BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                 "<A HREF=\"%s?sortby=%s&amp;directory=/&amp;token=%s\">\n"\
                    "<IMG SRC=\"%shome.gif\" BORDER=0 ALIGN=\"MIDDLE\" WIDTH=\"16\" HEIGHT=\"16\" ALT=\"Home\">&nbsp;Home\n"\
                 "</A>\n"\
                "</TD>\n",
                cgiScriptName, sortby, token, ICONSURL);

    fprintf(cgiOut,                 
                "<!-- RELOAD -->\n"\
                "<TD  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                     "<A HREF=\"%s?sortby=%s&amp;directory=%s&amp;token=%s\">\n"\
                         "<IMG SRC=\"%sreload.gif\" BORDER=0 ALIGN=\"MIDDLE\" ALT=\"Reload\">&nbsp;Refresh\n"\
                     "</A>\n"\
                "</TD>\n",
                cgiScriptName, sortby, virt_dirname, token, ICONSURL);

    fprintf(cgiOut,                 
                "<!-- MULTI DELETE -->\n"\
                "<TD  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                     "<INPUT TYPE=\"IMAGE\" SRC=\"%sdelete.gif\" STYLE=\"border: none; padding: 0px; vertical-align:middle;\" ALT=\"Delete\" ALIGN=\"MIDDLE\" NAME=\"multi_delete_prompt\" VALUE=\"Delete\">\n"
                     "<INPUT TYPE=\"SUBMIT\" CLASS=\"hovout\" NAME=\"multi_delete_prompt\" VALUE=\"Delete\" onMouseOver=\"this.className='hovin';\" onMouseOut=\"this.className='hovout';\">\n"
                "</TD>\n",
                ICONSURL);

    fprintf(cgiOut,                 
                "<!-- MULTI MOVE -->\n"\
                "<TD  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                     "<INPUT TYPE=\"IMAGE\" SRC=\"%smove.gif\" STYLE=\"border: none; padding: 0px; vertical-align:middle; \" ALT=\"Move\" ALIGN=\"MIDDLE\" NAME=\"multi_move_prompt\" VALUE=\"Move\">\n"
                     "<INPUT TYPE=\"SUBMIT\" CLASS=\"hovout\" NAME=\"multi_move_prompt\" VALUE=\"Move\" onMouseOver=\"this.className='hovin';\" onMouseOut=\"this.className='hovout';\">\n"
                "</TD>\n",
                ICONSURL);

    fprintf(cgiOut,                                 
                "<!-- NEWDIR -->\n"\
                "<TD  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                     "<A HREF=\"%s?action=mkdir_prompt&amp;directory=%s&amp;token=%s\" >\n"\
                            "<IMG SRC=\"%smkdir.gif\" BORDER=0 ALIGN=\"MIDDLE\" ALT=\"New Folder\">&nbsp;New Folder\n"\
                     "</A>\n"\
                "</TD>\n",
                cgiScriptName, virt_dirname, token, ICONSURL);
                

    fprintf(cgiOut,                                 
                "<!-- NEWFILE -->\n"\
                "<TD  BGCOLOR=\"#F1F1F1\" VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"\
                     "<A HREF=\"%s?action=mkfile_prompt&amp;directory=%s&amp;token=%s\" >\n"\
                            "<IMG SRC=\"%smkfile.gif\" BORDER=0 ALIGN=\"MIDDLE\" ALT=\"New File\">&nbsp;New File\n"\
                     "</A>\n"\
                "</TD>\n",
                cgiScriptName, virt_dirname, token, ICONSURL);


                
    fprintf(cgiOut,                 
                "<!-- UPLOAD -->\n"\
                "<TD BGCOLOR=\"#F1F1F1\"  VALIGN=\"MIDDLE\" ALIGN=\"CENTER\">\n"
                    "<INPUT TYPE=\"hidden\" NAME=\"directory\" VALUE=\"%s\">\n"
                    "<INPUT TYPE=\"hidden\" NAME=\"token\" VALUE=\"%s\">\n"
                    "<INPUT TYPE=\"hidden\" NAME=\"upload_id\" VALUE=\"%s\">\n"
                    "<INPUT TYPE=\"file\" NAME=\"filename\">&nbsp;\n"
                    "<INPUT TYPE=\"submit\" NAME=\"upload\" ID=\"Upload_Status\" VALUE=\"Upload\" >\n"
                "</TD>\n"
                "</TR>\n"
            "</TABLE>\n",
            virt_dirname, token, "1234");

    //
    // SORT BY
    //
    if(strcmp(sortby, "size")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>", cgiScriptName, virt_dirname, token);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=rsize\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>&nbsp;%s", cgiScriptName, virt_dirname, token, ADNIMG);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>", cgiScriptName, virt_dirname, token);
    } else if(strcmp(sortby, "rsize")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>", cgiScriptName, virt_dirname, token);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>&nbsp;%s", cgiScriptName, virt_dirname, token, AUPIMG);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>", cgiScriptName, virt_dirname, token);
    } else if(strcmp(sortby, "date")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>", cgiScriptName, virt_dirname, token);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>", cgiScriptName, virt_dirname, token);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=rdate\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>&nbsp;%s", cgiScriptName, virt_dirname, token, ADNIMG);
    } else if(strcmp(sortby, "rdate")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>", cgiScriptName, virt_dirname, token);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>", cgiScriptName, virt_dirname, token);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>&nbsp;%s", cgiScriptName, virt_dirname, token, AUPIMG);
    } else if(strcmp(sortby, "name")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=rname\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>&nbsp;%s", cgiScriptName, virt_dirname, token, ADNIMG);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>", cgiScriptName, virt_dirname, token);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>", cgiScriptName, virt_dirname, token);
    } else if(strcmp(sortby, "rname")==0) {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>&nbsp;%s", cgiScriptName, virt_dirname, token, AUPIMG);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>", cgiScriptName, virt_dirname, token);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>", cgiScriptName, virt_dirname, token);
    } else {
        snprintf(namepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=name\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Filename</A>", cgiScriptName, virt_dirname, token);
        snprintf(sizepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=size\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Size</A>", cgiScriptName, virt_dirname, token);
        snprintf(datepfx, 1024, "&nbsp;<A HREF=\"%s?directory=%s&amp;token=%s&amp;sortby=date\" STYLE=\"text-decoration: none; color:#FFFFFF;\">Modified</A>", cgiScriptName, virt_dirname, token);
    }


    // SORTBY ROW + dir files display
    fprintf(cgiOut, 
            "<!-- MAIN FILE TABLE --> \n"\
            "<TABLE WIDTH=\"100%%\" BGCOLOR=\"#FFFFFF\" CELLPADDING=0 CELLSPACING=0 BORDER=0 >\n"\
            "<!-- SORTBY LINE -->\n"\
            "<TR BGCOLOR=\"#FFFFFF\" >\n"\
                "<TD ALIGN=\"left\" WIDTH=\"50%%\" BGCOLOR=\"#A0A0A0\">\n"\
                "<FONT COLOR=\"#FFFFFF\">\n"\
                "<INPUT TYPE=\"CHECKBOX\" NAME=\"CHECKALL\"  STYLE=\"padding: 0px; border: none;\" ONCLICK=\"checkUncheckAll(this, multiselect_filename);\">\n"
                "%s\n"\
                "</FONT>\n"\
                "</TD>\n"\
                "<TD ALIGN=\"right\" BGCOLOR=\"#A0A0A0\">\n"\
                "<FONT COLOR=\"#FFFFFF\">\n"\
                    "%s\n"\
                "</FONT>\n"\
                "</TD>\n"\
                "<TD ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n"\
                "<FONT COLOR=\"#FFFFFF\">\n"\
                    "%s\n"\
                "</FONT>\n"\
                "</TD>\n"\
                "<TD ALIGN=\"right\"  BGCOLOR=\"#A0A0A0\">\n"\
                    "&nbsp;"\
                "</TD>"\
                "<TD ALIGN=\"left\"  BGCOLOR=\"#A0A0A0\">\n"\
                "<FONT COLOR=\"#FFFFFF\">\n"\
                    "&nbsp;\n"\
                "</FONT>\n"\
                "</TD>\n"\
            "</TR>\n"
            "<!-- End of Header -->\n\n",
            namepfx, sizepfx, datepfx);


    // Directories
    for(e=0; e<nentr; e++) {
        if(!S_ISDIR(direntry[e].type)) 
            continue;
        if(direntry[e].name[0]=='.') 
            continue; 
            
        name=direntry[e].name;
        if(recursive_du) {
            snprintf(phys_filename, PHYS_FILENAME_SIZE, "%s/%s", phys_dirname, direntry[e].name);
            size=du(phys_filename);
        }
        else {
            size=-1;
        }

        ctime_r(&direntry[e].atime, atime);
        ctime_r(&direntry[e].mtime, mtime);
        ctime_r(&direntry[e].rtime, rtime);
        
             if(now-direntry[e].mtime < 3600)        stime=M_HR;
        else if(now-direntry[e].mtime < 24*3600)     stime=M_DAY;
        else if(now-direntry[e].mtime < 7*24*3600)   stime=M_WK;
//      else if(now-direntry[e].mtime < 14*24*3600)  stime=M_2WK;
        else if(now-direntry[e].mtime < 31*24*3600)  stime=M_MO;
//      else if(now-direntry[e].mtime < 62*24*3600)  stime=M_2MO;
//      else if(now-direntry[e].mtime < 182*24*3600) stime=M_6MO;
        else if(now-direntry[e].mtime < 365*24*3600) stime=M_YR;
        else                                         stime=M_OLD;

        if(strcmp(highlight, name)==0)  {
            icon=NEWIMG;
            linecolor=tHIGH_COLOR;
        }
        else {
            icon=DIRIMG;
            linecolor=tNORMAL_COLOR;
        }


        // directory name / date
        fprintf(cgiOut, 
            "<!-- Directory Entry -->\n"\
            "<TR BGCOLOR=\"#%s\" onMouseOver=\"this.bgColor='#%s';\" onMouseOut=\"this.bgColor='#%s';\">\n"\
            "<TD ALIGN=\"LEFT\">\n"
            "<INPUT TYPE=\"CHECKBOX\" NAME=\"multiselect_filename\" STYLE=\"border: none;\" VALUE=\"%s\">", 
        linecolor, HL_COLOR, linecolor, name);
                    
        fprintf(cgiOut, 
            "<A HREF=\"%s?sortby=%s&amp;directory=%s/%s&amp;token=%s\">%s %s</A></TD> \n"\
            "<TD ALIGN=\"RIGHT\">%s</TD>\n"\
            "<TD ALIGN=\"RIGHT\"><SPAN TITLE=\"Created:%s\n Modified:%s\n Accessed:%s\n\">%s&nbsp;%s</FONT></SPAN></TD>\n"\
            "<TD>&nbsp;</TD>"\
            "<TD ALIGN=\"LEFT\">",
        cgiScriptName, sortby, (strcmp(virt_dirname, "/")==0) ? "" : virt_dirname, name,  token, icon, name,  
        buprintf(size, TRUE), rtime, mtime, atime, stime, mtime);

        // rename
        fprintf(cgiOut, "\n"\
            "<A HREF=\"%s?action=rename_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Rename '%s'\">\n"\
            "<IMG SRC=\"%srename.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Rename File\">\n"\
            "</A>\n",
            cgiScriptName, virt_dirname, name, token, name, ICONSURL);

        // move
        fprintf(cgiOut, "\n"\
            "<A HREF=\"%s?action=move_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Move '%s'\">\n"\
            "<IMG SRC=\"%smove.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Move File\">\n"\
            "</A>\n",
        cgiScriptName, virt_dirname, name, token, name, ICONSURL);

        // delete
        fprintf(cgiOut, "\n"\
            "<A HREF=\"%s?action=delete_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Delete '%s'\">\n"\
            "<IMG SRC=\"%sdelete.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Delete Directory\">\n"\
            "</A>\n"\
            "</TD>\n"\
            "</TR>\n\n\n",
        cgiScriptName, virt_dirname, name, token, name, ICONSURL);
                  
        totalsize+=size;
    }

                            
    // regular files
    for(e=0; e<nentr; e++) {
        if(S_ISDIR(direntry[e].type)) 
            continue;
        if(direntry[e].name[0]=='.')
            continue; 

        name=direntry[e].name;
        size=direntry[e].size;

        ctime_r(&direntry[e].atime, atime);
        ctime_r(&direntry[e].mtime, mtime);
        ctime_r(&direntry[e].rtime, rtime);
        
             if(now-direntry[e].mtime < 3600)        stime=M_HR;
        else if(now-direntry[e].mtime < 24*3600)     stime=M_DAY;
        else if(now-direntry[e].mtime < 7*24*3600)   stime=M_WK;
//      else if(now-direntry[e].mtime < 14*24*3600)  stime=M_2WK;
        else if(now-direntry[e].mtime < 31*24*3600)  stime=M_MO;
//      else if(now-direntry[e].mtime < 62*24*3600)  stime=M_2MO;
//      else if(now-direntry[e].mtime < 182*24*3600) stime=M_6MO;
        else if(now-direntry[e].mtime < 365*24*3600) stime=M_YR;
        else                                         stime=M_OLD;

             if(regexec(&reg_zip, name, 0, 0, 0)==0)    { icon=ZIPIMG; editable=0; }
        else if(regexec(&reg_img, name, 0, 0, 0)==0)    { icon=IMGIMG; editable=0; }
        else if(regexec(&reg_off, name, 0, 0, 0)==0)    { icon=OFFIMG; editable=0; }
        else if(regexec(&reg_pdf, name, 0, 0, 0)==0)    { icon=PDFIMG; editable=0; }
        else if(regexec(&reg_txt, name, 0, 0, 0)==0)    { icon=TXTIMG; editable=1; }
        else if(regexec(&reg_exe, name, 0, 0, 0)==0)    { icon=EXEIMG; editable=0; }
        else if(regexec(&reg_med, name, 0, 0, 0)==0)    { icon=MEDIMG; editable=0; }
        else if(regexec(&reg_iso, name, 0, 0, 0)==0)    { icon=ISOIMG; editable=0; }
        else                                            { icon=GENIMG; editable=0; }

        if(edit_any_file)                               { editable=1; }

        if(strcmp(highlight, name)==0)   { 
            icon=NEWIMG; 
            linecolor=tHIGH_COLOR;
        }
        else {
             linecolor=tNORMAL_COLOR;
        }

        // filename 
        fprintf(cgiOut, 
            "<!-- File Entry -->\n"
            "<TR  BGCOLOR=\"#%s\" onMouseOver=\"this.bgColor='#%s';\" onMouseOut=\"this.bgColor='#%s';\">\n"
            "<TD ALIGN=\"LEFT\"><INPUT TYPE=\"CHECKBOX\" NAME=\"multiselect_filename\" STYLE=\"border: none;\" VALUE=\"%s\">"
            "<A HREF=\"%s?action=%s&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Open '%s'\">%s %s</A></TD>\n",
        linecolor, HL_COLOR, linecolor, name, cgiScriptName, (edit_by_default && editable) ? "edit" : "sendfile", 
        virt_dirname, name, token, name, icon, name);


        // size / date
        fprintf(cgiOut, 
            "\n"
            "<TD ALIGN=\"RIGHT\" >%s</TD>\n"
            "<TD ALIGN=\"RIGHT\" ><SPAN TITLE=\"Created:%s\n Modified:%s\n Accessed:%s\n\">%s&nbsp;%s</FONT></SPAN></TD>\n",
        buprintf(size, TRUE), rtime, mtime, atime, stime, mtime);


        // file tools
        fprintf(cgiOut, "\n<TD>&nbsp;</TD><TD ALIGN=\"LEFT\">\n");


        // rename
        fprintf(cgiOut, 
            "<A HREF=\"%s?action=rename_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Rename '%s'\">\n"
            "<IMG SRC=\"%srename.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Rename File\">\n"
            "</A>\n",
            cgiScriptName, virt_dirname, name, token, name, ICONSURL);

        // move
        fprintf(cgiOut, 
            "\n"
            "<A HREF=\"%s?action=move_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Move '%s'\">"
            "<IMG SRC=\"%smove.gif\" BORDER=0 WIDTH=16 HEIGHT=16  ALT=\"Move '%s'\">\n"
            "</A>\n",
            cgiScriptName, virt_dirname, name, token, name,  ICONSURL, name);

        // delete
        fprintf(cgiOut, 
            "\n"
            "<A HREF=\"%s?action=delete_prompt&amp;directory=%s&amp;filename=%s&amp;token=%s\" "
            "TITLE=\"Remove '%s'\"> \n"
            "<IMG SRC=\"%sdelete.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Delete File\">\n"
            "</A>\n",
            cgiScriptName, virt_dirname, name, token, name, ICONSURL);


        // view
        if(strlen(HOMEURL)>4)
            fprintf(cgiOut, 
                "\n"
                "<A HREF=\"%s%s%s/%s\" TITLE=\"Preview '%s' In Browser\">\n"
                "<IMG SRC=\"%sext.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Preview '%s' In Browser\" >\n"
                "</A>\n", 
            HOMEURL, (virt_dirname[0]!='/') ? "/" : "", (strcmp(virt_dirname, "/")==0) ? "" : virt_dirname, name, name, ICONSURL,  name);

        
        // edit for text files..
        if(editable) {
            if(edit_by_default) 
                fprintf(cgiOut, 
                    "\n"
                    "<A HREF=\"%s?action=sendfile&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Download '%s'\">\n"
                    "<IMG SRC=\"%sdisk.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Download File\">\n"
                    "</A>\n"
                    "</TD>\n"
                    "</TR>\n\n",
                cgiScriptName, virt_dirname, name, token, name, ICONSURL);
            else
                fprintf(cgiOut, 
                    "\n"
                    "<A HREF=\"%s?action=edit&amp;directory=%s&amp;filename=%s&amp;token=%s\" TITLE=\"Edit '%s'\">\n"
                    "<IMG SRC=\"%sedit.gif\" BORDER=0 WIDTH=16 HEIGHT=16 ALT=\"Edit File\">\n"
                    "</A>\n"
                    "</TD>\n"
                    "</TR>\n\n",
                cgiScriptName, virt_dirname, name, token, name, ICONSURL);
        }
        else {
            fprintf(cgiOut, 
                "\n"
                "&nbsp;\n"
                "</TD>\n"
                "</TR>\n\n"
            );
       }

       totalsize+=size;

    }

    tstop();

    //
    // footer line
    //
    fprintf(cgiOut, 
        "<!-- FOOTER -->\n"
        "<TR>\n"
            "<TD BGCOLOR=\"#%s\">&nbsp;</TD>\n"
            "<TD BGCOLOR=\"#%s\" ALIGN=\"right\" STYLE=\"border-top:1px solid grey\">total %s </TD>\n"
            "<TD BGCOLOR=\"#%s\" ALIGN=\"right\" STYLE=\"color:#D0D0D0;\">%.1f ms</TD>\n"
            "<TD BGCOLOR=\"#%s\">&nbsp;</TD>\n"
            "<TD BGCOLOR=\"#%s\">&nbsp;</TD>\n"
        "</TR>\n"
        "</TABLE>\n</FORM>\n</BODY>\n<!-- Page generated in %f seconds -->\n</HTML>\n\n",
        NORMAL_COLOR, NORMAL_COLOR, buprintf(totalsize, TRUE), NORMAL_COLOR, (t2-t1)*1000, NORMAL_COLOR, NORMAL_COLOR, t2-t1
    );
    
}

