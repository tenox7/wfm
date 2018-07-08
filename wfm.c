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
        rt.iconsurl, cfg.favicon, cfg.tagline, msg); // (strlen(virt_dirname)>0) ? ' ' : '/', TAGLINE, virt_dirname
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
char *md5hash(char *str, ...) {
    va_list ap;
    char buff[1024]={0};
    md5_state_t state;
    md5_byte_t digest[16]={0};
    char *outstr;
    int i;

    if(str) {
        va_start(ap, str);
        vsnprintf(buff, sizeof(buff), str, ap);
        va_end(ap);

        outstr=(char*) malloc((sizeof(digest)*2)+2);
        memset(outstr, 0, (sizeof(digest)*2)+2);

        md5_init(&state);
        md5_append(&state, (const md5_byte_t *)buff, strlen(buff));
        md5_finish(&state, digest);

        for (i = 0; i < sizeof(digest); i++)
                sprintf(outstr + i * 2, "%02x", digest[i]);
        
        if(strlen(outstr))
            return outstr;
        else
            return NULL;
    }
    else {
        return NULL;
    }
}


//
// WFM Login Procedure
// Called from WFM main procedure if no sufficient access permission available and no
// JavaScript is available in the browser. Normally client side genrates the auth token.
//
void login(void) {
    char username[64]={0};
    char password[64]={0};

    cgiFormStringNoNewlines("username", username, sizeof(username));
    cgiFormStringNoNewlines("password", password, sizeof(password));
    
    if(strlen(username) && strlen(password)) 
        redirect("%s?directory=%s&login=server&token=%s", cgiScriptName, virt_dirname_urlencoded, md5hash("%s:%s", cgiRemoteAddr, md5hash("%s:%s", username, password)));  // generate MD5 as if it was the client
    else
        login_ui(); // display actual login page, which normally generates token in JavaScript
        
}

//
// Access_check 
// Called by cfg read routine during initialization
//
void access_check(char *access_string) {
    char ipaddr[32]={0};
    char user[64]={0};
    char pass[64]={0};
    char type[4]={0};

    if(sscanf(access_string, "access-ip=%2s:%30s", type, ipaddr)==2) {

        if(ipaddr[0]=='*' || strcmp(cgiRemoteAddr, ipaddr)==0) {
            if(strcmp(type, "ro")==0) 
                rt.access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                rt.access_level=PERM_RW;
        }

    }
    else if(sscanf(access_string, "access-md5pw=%2[^':']:%30[^':']:%63s", type, user, pass)==3) {
        cfg.users_defined=1;

        // perform user auth by comparing user supplied token with system generated token
        if(strcmp(md5hash("%s:%s", cgiRemoteAddr, pass), rt.token)==0) {
            if(strcmp(type, "ro")==0) 
                rt.access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                rt.access_level=PERM_RW;

            rt.access_as_user=1;
            strncpy(rt.loggedinuser, user, sizeof(rt.loggedinuser));
        }
    }
    else if(sscanf(access_string, "access-htauth=%2[^':']:%30s", type, user)==2) {
        cfg.users_defined=1;

        if(user[0]=='*' || (getenv("REMOTE_USER") && strcmp(user, getenv("REMOTE_USER"))==0)) {
            if(strcmp(type, "ro")==0) 
                rt.access_level=PERM_RO;
            else if(strcmp(type, "rw")==0) 
                rt.access_level=PERM_RW;

            rt.access_as_user=1;
            strncpy(rt.loggedinuser, getenv("REMOTE_USER"), sizeof(rt.loggedinuser));
        }
    }
}


//
// Check filename
// Should be called by every function that uses filename
// Function can be passed implicit filename or use the global variable
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

    // Do checks
    if(!strlen(phys_filename) || strlen(phys_filename)>(PHYS_FILENAME_SIZE-2)) error("Invalid phys_filename lenght [%d]", strlen(phys_filename));
    if(!strlen(virt_filename) || strlen(virt_filename)>(VIRT_FILENAME_SIZE-2)) error("Invalid virt_filename lenght [%d]", strlen(virt_filename));
    if(regexec(&dotdot, phys_filename, 0, 0, 0)==0) error("Double dots in pfilename");
    if(regexec(&dotdot, virt_filename, 0, 0, 0)==0) error("Double dots in vfilename");

    strncpy(temp_dirname, phys_filename, PHYS_FILENAME_SIZE);
    if(strlen(dirname(temp_dirname)) < strlen(cfg.homedir)) error("Invalid directory name.");

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
        snprintf(phys_destination, PHYS_DESTINATION_SIZE, "%s/%s", cfg.homedir, virt_destination);
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
    snprintf(phys_dirname, PHYS_DIRNAME_SIZE, "%s/%s", cfg.homedir, virt_dirname);

    if(strlen(phys_dirname)<2 || strlen(phys_dirname)>(PHYS_DIRNAME_SIZE-2)) 
        error("Invalid directory name.");

    if(regexec(&dotdot, phys_dirname, 0, 0, 0)==0) error("Invalid directory name.");
    if(strlen(phys_dirname) < strlen(cfg.homedir)) error("Invalid directory name.");

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


// split string in to array of char[] with separators
// Warning this function modifies src!
// Written by Tomasz Nowak
int strsplit(char *src, char ***dst, char *sep) {
    char **arr;
    char defsep[]=" \t\n\r\f";

    if(!sep)
        sep=defsep;

    char *src_org = src;
    char *c;
    int n = 0;

    while ((c = strpbrk(src, sep))) {
        while (c == src) {
            src++;
            c = strpbrk(src, sep);
        }
        if (c == NULL) 
            break;

        src = c + 1;
        n++;
    }

    int n_elem = n + 1;
    arr = (char **)malloc(sizeof(char *) * n_elem);
    memset(arr, 0, sizeof(char *) * n_elem);

    src = src_org;
    n = 0;

    while ((c = strpbrk(src, sep))) {
        while (c == src) {
            src++;
            c = strpbrk(src, sep);
        }
        if (c == NULL) 
            break;

        *c = 0;
        arr[n] = src;

        src = c + 1;
        n++;
    }

    arr[n] = src;

    *dst = arr;

    if(!*arr[n])
        return n_elem-1;
    else
        return n_elem;
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
// Debug print to a file
//
void dbgprintf(char *msg, ...) {
    va_list ap;
    char buff[1024]={0};
    FILE *f;

    if(msg) {
        va_start(ap, msg);
        vsnprintf(buff, sizeof(buff), msg, ap);
        va_end(ap);

        f=fopen("/tmp/wfmdbg.log", "a");
        if(!f)
            error("Unable to open debug file");

        fprintf(f, "DEBUG: %s\n", buff);

        fflush(f);
        fclose(f);
    }

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
// Load and process config file
// Invoke Access Check
//
void cfgload(void) {
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
    char c_largeset[]="large-file-set=true";
    char c_access[]="access";

    memset(&cfg, 0, sizeof(cfg));
    memset(&rt, 0, sizeof(rt));

    cgiFormStringNoNewlines("token", rt.token, sizeof(rt.token));
    snprintf(rt.iconsurl, sizeof(rt.iconsurl), "%s?ea=icon&amp;name=", cgiScriptName);

    snprintf(cfgname, sizeof(cfgname), "%s.cfg", basename(cgiScriptName));
    cfgfile=fopen(cfgname, "r");
    if(!cfgfile)
        error("Unable to open configuration file %s.<BR>%s", cfgname, strerror(errno));

    while(fgets(cfgline, sizeof(cfgline), cfgfile)) {
        if((*cfgline==';')||(*cfgline=='/')||(*cfgline=='#')||(*cfgline=='\n')) continue;
        else if(strncmp(cfgline, c_homedir, strlen(c_homedir))==0)              strncpy(cfg.homedir, cfgline+strlen(c_homedir), sizeof(cfg.homedir));
        else if(strncmp(cfgline, c_homeurl, strlen(c_homeurl))==0)              strncpy(cfg.homeurl, cfgline+strlen(c_homeurl), sizeof(cfg.homeurl));
        else if(strncmp(cfgline, c_tagline, strlen(c_tagline))==0)              strncpy(cfg.tagline, cfgline+strlen(c_tagline), sizeof(cfg.tagline));
        else if(strncmp(cfgline, c_favicon, strlen(c_favicon))==0)              strncpy(cfg.favicon, cfgline+strlen(c_favicon), sizeof(cfg.favicon));
        else if(strncmp(cfgline, c_editdef, strlen(c_editdef))==0)              cfg.edit_by_default=1;
        else if(strncmp(cfgline, c_editany, strlen(c_editany))==0)              cfg.edit_any_file=1;
        else if(strncmp(cfgline, c_largeset, strlen(c_largeset))==0)            cfg.largeset=1;
        else if(strncmp(cfgline, c_du, strlen(c_du))==0)                        cfg.recursive_du=1;
        else if(strncmp(cfgline, c_access, strlen(c_access))==0)                access_check(cfgline);
    }
    fclose(cfgfile);
    
    // remove newlines
    if(strlen(cfg.homedir)>2) cfg.homedir[strlen(cfg.homedir)-1]='\0';
    if(strlen(cfg.homeurl)>2) cfg.homeurl[strlen(cfg.homeurl)-1]='\0';
    if(strlen(cfg.tagline)>2) cfg.tagline[strlen(cfg.tagline)-1]='\0';
    if(strlen(cfg.favicon)>2) cfg.favicon[strlen(cfg.favicon)-1]='\0';
    
    // do checks
    if(strlen(cfg.homedir) < 4)
        error("Home directory not defined.");

    if(cfg.homedir[0]!='/')
        error("Home directory must be absolute path.");

    if(!strlen(cfg.tagline))
        strcpy(cfg.tagline, "Web File Manager");

    if(!strlen(cfg.favicon))
        strcpy(cfg.favicon, "wfmicon.gif");

    checkdirectory();


    // JavaScript check
         if(strncmp(cgiUserAgent, "Mozilla/5", 9)==0)                               rt.js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 6", 31)==0)        rt.js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 7", 31)==0)        rt.js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4.0 (compatible; MSIE 8", 31)==0)        rt.js=2;
    else if(strncmp(cgiUserAgent, "Mozilla/4", 9)==0)                               rt.js=1;
    else                                                                            rt.js=0;

}

//
// WFM Entry
//
int cgiMain(void) {
    char action[32]={0};
    char ea[8]={0};

    // early action - simple actions before cfg is read or access check performed (no authentication!)
    cgiFormStringNoNewlines("ea", ea, sizeof(ea));
    if(strcmp(ea, "icon")==0) icon();
    if(strcmp(ea, "upstat")==0) upload_status();

    // normal initialization
    tstart();

    fprintf(cgiOut, "Cache-Control: max-age=0, private\r\nExpires: -1\r\n");

    memset(virt_dirname, 0, VIRT_DIRNAME_SIZE);
    memset(phys_dirname, 0, PHYS_DIRNAME_SIZE); 
    memset(virt_filename, 0, VIRT_FILENAME_SIZE); 
    memset(phys_filename, 0, PHYS_FILENAME_SIZE);
    memset(virt_destination, 0, VIRT_DESTINATION_SIZE); 
    memset(phys_destination, 0, PHYS_DESTINATION_SIZE);
    memset(final_destination, 0, PHYS_DESTINATION_SIZE);
    memset(virt_parent, 0, VIRT_DIRNAME_SIZE);

    cfgload();



    // main routine - regular actions
    cgiFormStringNoNewlines("action", action, sizeof(action));
    if(cgiFormSubmitClicked("noop")==cgiFormSuccess                       && rt.access_level >= PERM_RO)         dirlist();
    else if(cgiFormSubmitClicked("multi_delete_prompt")==cgiFormSuccess   && rt.access_level >= PERM_RO)         multiprompt_ui("delete");
    else if(cgiFormSubmitClicked("multi_delete_prompt.x")==cgiFormSuccess && rt.access_level >= PERM_RO)         multiprompt_ui("delete");
    else if(cgiFormSubmitClicked("multi_move_prompt")==cgiFormSuccess     && rt.access_level >= PERM_RO)         multiprompt_ui("move");
    else if(cgiFormSubmitClicked("multi_move_prompt.x")==cgiFormSuccess   && rt.access_level >= PERM_RO)         multiprompt_ui("move");
    else if(cgiFormSubmitClicked("upload")==cgiFormSuccess                && rt.access_level >= PERM_RW)         receivefile();
    else if(strcmp(action, "sendfile")==0                                 && rt.access_level >= PERM_RO)         sendfile();
    else if(strcmp(action, "delete")==0                                   && rt.access_level >= PERM_RW)         delete();
    else if(strcmp(action, "delete_prompt")==0                            && rt.access_level >= PERM_RW)         multiprompt_ui("delete");
    else if(strcmp(action, "move_prompt")==0                              && rt.access_level >= PERM_RW)         multiprompt_ui("move");
    else if(strcmp(action, "rename_prompt")==0                            && rt.access_level >= PERM_RW)         singleprompt_ui("move");
    else if(strcmp(action, "move")==0                                     && rt.access_level >= PERM_RW)         move();
    else if(strcmp(action, "edit")==0                                     && rt.access_level >= PERM_RO)         edit_ui();
    else if(strcmp(action, "edit_save")==0                                && rt.access_level >= PERM_RW)         edit_save();
    else if(strcmp(action, "mkfile")==0                                   && rt.access_level >= PERM_RW)         mkfile();
    else if(strcmp(action, "mkfile_prompt")==0                            && rt.access_level >= PERM_RW)         singleprompt_ui("mkfile");
    else if(strcmp(action, "mkdir")==0                                    && rt.access_level >= PERM_RW)         newdir();
    else if(strcmp(action, "mkdir_prompt")==0                             && rt.access_level >= PERM_RW)         singleprompt_ui("mkdir");
    else if(strcmp(action, "about")==0                                    && rt.access_level >= PERM_RO)         about();
    else if(strcmp(action, "login")==0      )                                                                    login();
    else if(                                                                 rt.access_level >= PERM_RO)         dirlist();
    else 
        if(cfg.users_defined) // if users present but supplied credentials didn't match, or credentials not specified
            redirect("%s?action=login", cgiScriptName);
        else
            error("Access Denied.");

    return 0;
}


