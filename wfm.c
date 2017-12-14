#include "wfm.h"

// 
// Dispense Common HTML Header
// Used by all (?) functions that display HTML pages
//
void html_title(char *msg) {
    fprintf(cgiOut, 
        HTML_HEADER
        "<LINK REL=\"icon\" TYPE=\"image/gif\" HREF=\"%s%s\">\n"
        "<TITLE>%s : %s</TITLE>\n",
        ICONSURL, FAVICON, TAGLINE, msg); // (strlen(virt_dirname)>0) ? ' ' : '/', TAGLINE, virt_dirname
}


//
// Dispense embedded icons
// Called on early action=icon
//
int icon(void) {
    char icon_name[32]={0};

    cgiFormStringNoNewlines("name", icon_name, sizeof(icon_name));

    fprintf(cgiOut, "Cache-Control: max-age=29030400, public\r\n");
    cgiHeaderContentType("image/gif");

#include "wfmicondis.h"

    exit(0);        
}


//
// Display upload status in SHM via key_id
// Called by early action=upstat
//
void upload_status(void) {
    int shm_key=-1;
    int shm_id=-1;
    char *shm_val=NULL;
    time_t t;
    char *spin[]={"|", "/", "--", "\\"};

    if(cgiFormInteger("upload_id", &shm_key, 0) == cgiFormSuccess && shm_key) {
            shm_id = shmget(shm_key, SHM_SIZE, 0666);
            if(shm_id >= 0) 
                shm_val = shmat(shm_id, NULL, 0);
    }
 
    fprintf(cgiOut, "Cache-Control: no-cache\r\n");
    cgiHeaderContentType("text/plain");
    time(&t);

    if(shm_val)
        fprintf(cgiOut, "%s %s\r\n", spin[(int)t % 4], shm_val);
    else 
        fprintf(cgiOut, "%s\r\n", spin[(int)t % 4]);

    if (shm_val)
        shmdt(shm_val);

    exit(0);        
}


//
// Generate auth token
// Used by access_check() to compare tokens and login() to generate token from a web form
//
char *mktoken(char *str) {
    md5_state_t state;
    md5_byte_t digest[16]={0};
    char *outstr;
    int i;

    outstr=(char*) malloc((sizeof(digest)*2)+2);
    memset(outstr, 0, (sizeof(digest)*2)+2);

    md5_init(&state);
    md5_append(&state, (const md5_byte_t *)str, strlen(str));
    md5_finish(&state, digest);

    for (i = 0; i < sizeof(digest); i++)
            sprintf(outstr + i * 2, "%02x", digest[i]);

    return outstr;
}

//
// WFM Login Procedure
// Called from WFM main procedure if no sufficient access permission available
//
void login(void) {
    char username[64]={0};
    char password[64]={0};
    char token_inp[256]={0};

    cgiFormStringNoNewlines("username", username, sizeof(username)); // only used if JavaScript not 
    cgiFormStringNoNewlines("password", password, sizeof(password)); // available in the browser
    
    if(strlen(username)) {
        snprintf(token_inp, sizeof(token_inp), "%s:%s:%s", cgiRemoteAddr, username, password);
        redirect("%s?directory=%s&login=server&token=%s", cgiScriptName, virt_dirname_urlencoded, mktoken(token_inp));  // generate MD5 as if it was the client
    }
    else
        login_ui(); // display actual login page, which normally generates token in JavaScript
        
}

//
// Access_check 
// Called by cfg read routine during initialization
//
void access_check(char *access_string) {
    char ipaddr[32]={0};
    char user[32]={0};
    char pass[32]={0};
    char type[4]={0};
    char token_inp[64]={0};

    if(sscanf(access_string, "access-ip=%2s:%30s", type, ipaddr)==2) {

        if(ipaddr[0]=='*' || strcmp(cgiRemoteAddr, ipaddr)==0) {
            if(strcmp(type, "ro")==0) 
                access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                access_level=PERM_RW;
        }

    }
    else if(sscanf(access_string, "access-user=%2[^':']:%30[^':']:%30s", type, user, pass)==3) {
        users_defined=1;

        snprintf(token_inp, sizeof(token_inp), "%s:%s:%s", cgiRemoteAddr, user, pass);
        // perform user auth by comparing user supplied token with system generated token
        if(strcmp(mktoken(token_inp), token)==0) {
            if(strcmp(type, "ro")==0) 
                access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                access_level=PERM_RW;

            access_as_user=1;
            strncpy(loggedinuser, user, sizeof(loggedinuser));
        }
    }
    else if(sscanf(access_string, "access-htauth=%2[^':']:%30s", type, user)==2) {
        users_defined=1;

        if(user[0]=='*' || strcmp(user, getenv("REMOTE_USER"))==0) {
            if(strcmp(type, "ro")==0) 
                access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                access_level=PERM_RW;

            access_as_user=1;
            strncpy(loggedinuser, getenv("REMOTE_USER"), sizeof(loggedinuser));
        }
    }
}


//
// Check filename
// Should be called by every function that uses filename
//
void checkfilename(char *inp_filename) {
    char temp_dirname[PHYS_FILENAME_SIZE]={0};
    char temp_filename[VIRT_FILENAME_SIZE]={0};
    char *bname;

    if(inp_filename && strlen(inp_filename)) {
        strncpy(temp_filename, inp_filename, VIRT_FILENAME_SIZE);
    }
    else if(cgiFormFileName("filename", temp_filename, VIRT_FILENAME_SIZE) == cgiFormSuccess) {
        
    }
    else if(cgiFormStringNoNewlines("filename", temp_filename, VIRT_FILENAME_SIZE) == cgiFormSuccess) {
        
    }
    else
        error("No filename specified.");

    // We only want basename from the client!
    bname=strrchr(temp_filename, '/');
    if(!bname)
        bname=strrchr(temp_filename, '\\');
            
    if(!bname)
        bname=temp_filename;
    else
        (void) *bname++;

    strip(bname, VIRT_FILENAME_SIZE, VALIDCHRS);
    strncpy(virt_filename, bname, VIRT_FILENAME_SIZE);
    snprintf(phys_filename, PHYS_FILENAME_SIZE, "%s/%s", phys_dirname, virt_filename);

    if(!strlen(phys_filename) || strlen(phys_filename)>(PHYS_FILENAME_SIZE-2)) error("Invalid phys_filename lenght [%d]", strlen(phys_filename));
    if(!strlen(virt_filename) || strlen(virt_filename)>(VIRT_FILENAME_SIZE-2)) error("Invalid virt_filename lenght [%d]", strlen(virt_filename));
    if(regexec(&dotdot, phys_filename, 0, 0, 0)==0) error("Double dots in pfilename");
    if(regexec(&dotdot, virt_filename, 0, 0, 0)==0) error("Double dots in vfilename");

    strncpy(temp_dirname, phys_filename, PHYS_FILENAME_SIZE);
    if(strlen(dirname(temp_dirname)) < strlen(HOMEDIR)) error("Invalid directory name.");

    virt_filename_urlencoded=url_encode(virt_filename);
}

//
// Check destination
// Only called by move()
//
void checkdestination(void) {
    int absolute_destination;
    
    cgiFormStringNoNewlines("destination", virt_destination, VIRT_DESTINATION_SIZE);
    strip(virt_destination, VIRT_DESTINATION_SIZE, VALIDCHRS_DIR);
    cgiFormInteger("absdst", &absolute_destination, 0);  // move operation relies on absolute paths
    if(absolute_destination)
        snprintf(phys_destination, PHYS_DESTINATION_SIZE, "%s/%s", HOMEDIR, virt_destination);
    else
        snprintf(phys_destination, PHYS_DESTINATION_SIZE, "%s/%s", phys_dirname, virt_destination);

    if(strlen(phys_destination)<1 || strlen(phys_destination)>(PHYS_DESTINATION_SIZE-2)) error("Invalid phys_destination lenght [%d]", strlen(phys_destination));
    if(strlen(virt_destination)<1 || strlen(virt_destination)>(VIRT_DESTINATION_SIZE-2)) error("Invalid virt_destination lenght [%d]", strlen(virt_destination));
    if(regexec(&dotdot, phys_destination, 0, 0, 0)==0) error("Double dots in pfilename");
    if(regexec(&dotdot, virt_destination, 0, 0, 0)==0) error("Double dots in vfilename");
}

//
// Check directory
// Only called by cgiMain during initialization
//
void checkdirectory(void) {
    char temp[VIRT_DIRNAME_SIZE]={0};
    
    cgiFormStringNoNewlines("directory", virt_dirname, VIRT_DIRNAME_SIZE);
    strip(virt_dirname, VIRT_DIRNAME_SIZE, VALIDCHRS_DIR);
    snprintf(phys_dirname, PHYS_DIRNAME_SIZE, "%s/%s", HOMEDIR, virt_dirname);

    if(strlen(phys_dirname)<2 || strlen(phys_dirname)>(PHYS_DIRNAME_SIZE-2)) 
        error("Invalid directory name.");

    if(regexec(&dotdot, phys_dirname, 0, 0, 0)==0) error("Invalid directory name.");
    if(strlen(phys_dirname) < strlen(HOMEDIR)) error("Invalid directory name.");

    if(!strlen(virt_dirname)) strcpy(virt_dirname, "/");

    virt_dirname_urlencoded=url_encode(virt_dirname);

    // parent
    strncpy(temp, virt_dirname, VIRT_DIRNAME_SIZE);
    strncpy(virt_parent, dirname(temp), VIRT_DIRNAME_SIZE);
    virt_parent_urlencoded=url_encode(virt_parent);
}


void tstart(void) {
        gettimeofday(&mt, 0);
        t1=mt.tv_sec+(mt.tv_usec/1000000.0);
}

void tstop(void) {
        gettimeofday(&mt, 0);
        t2=mt.tv_sec+(mt.tv_usec/1000000.0);
}

//  strip unwanted characters from string
//  deny all by default
//  allow a=alpha n=numeric u=spaces as underscore
//  use: allow="an.-@" for email
//  use: allow="anu.-_!" for filenames
//  ! is a cmd termination char
//  use !! to allow only ! char
//  use !a to allow only letter a
int strip(char *str, int len, char *allow) {
    int n,a;
    int alpha=0, number=0, spctou=0;
    char *dst;

    if(!str || !strlen(str) || !allow || !strlen(allow))
        return -1;

    if(*allow == 'a') {
        alpha=1;
        allow++;
    }
    if(*allow == 'n') {
        number=1;
        allow++;
    }
    if(*allow == 'u') {
        spctou=1;
        allow++;
    }
    if(*allow == '!') {
        allow++;
    }

    dst=str;

    for(n=0; n<len && *str!='\0'; n++, str++) {
        if(alpha && isalpha(*str))
            *(dst++)=*str;
        else if(number && isdigit(*str))
            *(dst++)=*str;
        else if(spctou && *str==' ')
            *(dst++)='_';
        else if(strlen(allow))
            for(a=0; a<strlen(allow); a++) 
                if(*str==allow[a])
                    *(dst++)=*str;
    }

    *dst='\0';

    return 0; //strlen(dst);
}

//
// Byte unit printf
//
char *buprintf(float v, int bold) {
    float size;
    char unit;
    char *buffer;

    if(v == -1)
        return "&nbsp;";

         if(v >= P1024_1 && v < P1024_2 )   { size = v / P1024_1; unit = 'K'; }
    else if(v >= P1024_2 && v < P1024_3 )   { size = v / P1024_2; unit = 'M'; }
    else if(v >= P1024_3 && v < P1024_4 )   { size = v / P1024_3; unit = 'G'; }
    else { size = v; unit = ' '; }

    buffer=(char *)calloc(128, sizeof(char));
    
    if(unit == 'G' && bold)
        snprintf(buffer, 128, "<B> %5.1f %cB </B>", size, unit);
    else if(unit == 'M' && bold)
        snprintf(buffer, 128, "%5.1f %cB", size, unit);
    else if(bold)
        snprintf(buffer, 128, "<FONT COLOR=\"#909090\"> %5.1f %cB </FONT>", size, unit);
    else
        snprintf(buffer, 128, " %5.1f %cB", size, unit);

    return (char *)buffer;
}

//
// redirect browser
//
void redirect(char *location, ...) {
    va_list ap;
    char buff[1024]={0};

    va_start(ap, location);
    vsnprintf(buff, sizeof(buff), location, ap);
    va_end(ap);

    cgiHeaderLocation(buff);
}

//
// CGI entry
//
int cgiMain(void) {
    char action[32]={0};
    char ea[8]={0};
    FILE *cfgfile;
    char cfgname[128]={0};
    char cfgline[256]={0};
    char c_tagline[]="tagline=";
    char c_favicon[]="favicon=";
    char c_homeurl[]="browser-url=";
    char c_homedir[]="directory=";
    char c_editdef[]="txt-default-edit=true";
    char c_editany[]="edit-any-file=true";
    char c_du[]="recursive-du=true";
    char c_access[]="access";

    // early action - simple actions before cfg is read or access check performed (no security!)
    cgiFormStringNoNewlines("ea", ea, sizeof(ea));
    if(strcmp(ea, "icon")==0) icon();
    if(strcmp(ea, "upstat")==0) upload_status();

    // normal initialization
    tstart();

    if(regcomp(&dotdot, "\\.\\.", REG_EXTENDED | REG_ICASE)!=0)
        error("Regex compilation problem.<BR>%s", strerror(errno));

    fprintf(cgiOut, "Cache-Control: max-age=0, private\r\nExpires: -1\r\n");

    memset(virt_dirname, 0, VIRT_DIRNAME_SIZE);
    memset(phys_dirname, 0, PHYS_DIRNAME_SIZE); 
    memset(virt_filename, 0, VIRT_FILENAME_SIZE); 
    memset(phys_filename, 0, PHYS_FILENAME_SIZE);
    memset(virt_destination, 0, VIRT_DESTINATION_SIZE); 
    memset(phys_destination, 0, PHYS_DESTINATION_SIZE);
    memset(final_destination, 0, PHYS_DESTINATION_SIZE);
    memset(virt_parent, 0, VIRT_DIRNAME_SIZE);

    snprintf(ICONSURL, sizeof(ICONSURL), "%s?ea=icon&amp;name=", cgiScriptName);
    

    // config file defaults
    access_level=PERM_NO; // no access by default
    access_as_user=0;
    users_defined=0;
    edit_by_default=0; // for .txt files
    edit_any_file=0; 
    recursive_du=0;

    memset(HOMEDIR, 0, sizeof(HOMEDIR));
    memset(HOMEURL, 0, sizeof(HOMEURL));    
    memset(TAGLINE, 0, sizeof(TAGLINE));
    memset(FAVICON, 0, sizeof(FAVICON));

    // process config file
    cgiFormStringNoNewlines("token", token, sizeof(token));
    snprintf(cfgname, sizeof(cfgname), "%s.cfg", basename(cgiScriptName));
    cfgfile=fopen(cfgname, "r");
    if(!cfgfile)
        error("Unable to open configuration file %s.<BR>%s", cfgname, strerror(errno));
            
    while(fgets(cfgline, sizeof(cfgline), cfgfile)) {
        if((*cfgline==';')||(*cfgline=='/')||(*cfgline=='#')||(*cfgline=='\n')) continue;
        else if(strncmp(cfgline, c_homedir, strlen(c_homedir))==0)              strncpy(HOMEDIR, cfgline+strlen(c_homedir), sizeof(HOMEDIR));
        else if(strncmp(cfgline, c_homeurl, strlen(c_homeurl))==0)              strncpy(HOMEURL, cfgline+strlen(c_homeurl), sizeof(HOMEURL));
        else if(strncmp(cfgline, c_tagline, strlen(c_tagline))==0)              strncpy(TAGLINE, cfgline+strlen(c_tagline), sizeof(TAGLINE));
        else if(strncmp(cfgline, c_favicon, strlen(c_favicon))==0)              strncpy(FAVICON, cfgline+strlen(c_favicon), sizeof(FAVICON));
        else if(strncmp(cfgline, c_editdef, strlen(c_editdef))==0)              edit_by_default=1;
        else if(strncmp(cfgline, c_editany, strlen(c_editany))==0)              edit_any_file=1;
        else if(strncmp(cfgline, c_du, strlen(c_du))==0)                        recursive_du=1;
        else if(strncmp(cfgline, c_access, strlen(c_access))==0)                access_check(cfgline);
    }
    fclose(cfgfile);
    
    // remove newlines
    if(strlen(HOMEDIR)>2) HOMEDIR[strlen(HOMEDIR)-1]='\0';
    if(strlen(HOMEURL)>2) HOMEURL[strlen(HOMEURL)-1]='\0';
    if(strlen(TAGLINE)>2) TAGLINE[strlen(TAGLINE)-1]='\0';
    if(strlen(FAVICON)>2) FAVICON[strlen(FAVICON)-1]='\0';
    
    // do checks
    if(strlen(HOMEDIR) < 4 || *HOMEDIR!='/')
        error("Home directory not defined.");

    if(!strlen(TAGLINE))
        strcpy(TAGLINE, "Web File Manager");

    if(!strlen(FAVICON))
        strcpy(FAVICON, "wfmicon.gif");

    snprintf(VALIDCHRS_DIR, sizeof(VALIDCHRS_DIR), "%s/", VALIDCHRS);
    checkdirectory();

    // JavaScript check
         if(strncmp(cgiUserAgent, "Mozilla/5", 9)==0)                               js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 6", 31)==0)        js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 7", 31)==0)        js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 8", 31)==0)        js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4", 9)==0)                               js=1;
    else                                                                            js=0;

    // main routine - regular actions
    cgiFormStringNoNewlines("action", action, sizeof(action));
    if(cgiFormSubmitClicked("noop")==cgiFormSuccess                       && access_level >= PERM_RO)         dirlist();
    else if(cgiFormSubmitClicked("multi_delete_prompt")==cgiFormSuccess   && access_level >= PERM_RO)         multiprompt_ui("delete");
    else if(cgiFormSubmitClicked("multi_delete_prompt.x")==cgiFormSuccess && access_level >= PERM_RO)         multiprompt_ui("delete");
    else if(cgiFormSubmitClicked("multi_move_prompt")==cgiFormSuccess     && access_level >= PERM_RO)         multiprompt_ui("move");
    else if(cgiFormSubmitClicked("multi_move_prompt.x")==cgiFormSuccess   && access_level >= PERM_RO)         multiprompt_ui("move");
    else if(cgiFormSubmitClicked("upload")==cgiFormSuccess                && access_level >= PERM_RW)         receivefile();
    else if(strcmp(action, "sendfile")==0                                 && access_level >= PERM_RO)         sendfile();
    else if(strcmp(action, "delete")==0                                   && access_level >= PERM_RW)         delete();
    else if(strcmp(action, "delete_prompt")==0                            && access_level >= PERM_RW)         multiprompt_ui("delete");
    else if(strcmp(action, "move_prompt")==0                              && access_level >= PERM_RW)         multiprompt_ui("move");
    else if(strcmp(action, "rename_prompt")==0                            && access_level >= PERM_RW)         singleprompt_ui("move");
    else if(strcmp(action, "move")==0                                     && access_level >= PERM_RW)         move();
    else if(strcmp(action, "edit")==0                                     && access_level >= PERM_RO)         edit_ui();
    else if(strcmp(action, "edit_save")==0                                && access_level >= PERM_RW)         edit_save();
    else if(strcmp(action, "mkfile")==0                                   && access_level >= PERM_RW)         mkfile();
    else if(strcmp(action, "mkfile_prompt")==0                            && access_level >= PERM_RW)         singleprompt_ui("mkfile");
    else if(strcmp(action, "mkdir")==0                                    && access_level >= PERM_RW)         newdir();
    else if(strcmp(action, "mkdir_prompt")==0                             && access_level >= PERM_RW)         singleprompt_ui("mkdir");
    else if(strcmp(action, "about")==0                                    && access_level >= PERM_RO)         about();
    else if(strcmp(action, "login")==0      )                                                                 login();
    else if(                                                                 access_level >= PERM_RO)         dirlist();
    else 
        if(users_defined) // if users present but supplied credentials didn't match, or credentials not specified
            redirect("%s?action=login", cgiScriptName);
        else
            error("Access Denied.");

    return 0;
}


