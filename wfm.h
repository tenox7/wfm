#define VERSION "1.4.1"
#define COPYRIGHT "<!-- WFM Version " VERSION ", Mountain View, CA, " __DATE__ " [" __TIME__ "] -->\n" \
                  "<!-- Copyright (c) 1994-2018 by Antoni Sawicki -->\n" \
                  "<!-- Copyright (c) 2018-2019 by Google LLC -->\n"

#define FONT_SIZE "12px"

#define CSS_STYLE         \
            "  <STYLE TYPE=\"text/css\"><!-- \n" \
            "    A:link {text-decoration: none; color:#0000CE; } \n" \
            "    A:visited {text-decoration: none; color:#0000CE; } \n" \
            "    A:active {text-decoration: none; color:#FF0000; } \n" \
            "    A:hover {text-decoration: none; color:#FF0000; } \n" \
            "    body, td, th, input { font-family: Tahoma, Sans-Serif; font-size:" FONT_SIZE ";  } \n" \
            "    html, body { box-sizing: border-box; width:100%%; height:100%%; margin:0px; padding:0px; } \n" \
            "    input  { border-color:#000000; border-style:solid; }\n" \
            "    img { vertical-align: middle; }\n" \
            "    .tbr { border-width: 1px; border-style: solid solid solid solid; border-color: #AAAAAA #555555 #555555 #AAAAAA; }\n" \
            "    .twh { width:100%%; height:100%%; }\n" \
            "  --></STYLE>\n"

#define HTML_HEADER \
        "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01 Transitional//EN\"\n" \
        "   \"http://www.w3.org/TR/html4/loose.dtd\">\n" \
        "<HTML LANG=\"en\">\n" \
        COPYRIGHT \
        "<HEAD>\n" \
        "  <META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html;charset=US-ASCII\">\n" \
        "  <META HTTP-EQUIV=\"Content-Language\" CONTENT=\"en-US\">\n" \
        "  <META HTTP-EQUIV=\"google\" CONTENT=\"notranslate\">\n" \
        "  <META NAME=\"viewport\" CONTENT=\"width=device-width\">\n" \
        CSS_STYLE


#define _STRINGIFY(s) #s
#define STRINGIFY(s) _STRINGIFY(s)

#define _FILE_OFFSET_BITS 64

#ifdef __sun__
#define _POSIX_PTHREAD_SEMANTICS
#endif

#include <stdio.h>
#include <string.h>
#include <strings.h>
#include <stdlib.h>
#include <unistd.h>
#include <libgen.h>
#include <ctype.h>
#include <dirent.h>
#include <regex.h>
#include <stdarg.h>
#include <errno.h>
#include <time.h>
#include <limits.h>
#include <sys/time.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/shm.h>    
#include <sys/file.h>
#include <sys/utsname.h>

#include "md5.h"
#include "cgic.h"
#include "wfmiconres.h"

#define VALIDCHRS "an ()[]{}-_.,!@#$%^&=+;"
#define VALIDCHRS_DIR VALIDCHRS "/"

//#define SHM_SIZE 16

struct  wfm_paths {
    char virt_dirname[NAME_MAX];
    char *virt_dirname_urlencoded;
    char virt_filename[NAME_MAX]; 
    char *virt_filename_urlencoded; 
    char virt_destination[NAME_MAX]; 
    char virt_parent[1024];
    char *virt_parent_urlencoded;
    char phys_dirname[2048]; 
    char phys_filename[4096];
    char phys_destination[4096];
    char final_destination[8192];
} wp;

struct config_struct {
    int users_defined;
    int edit_by_default;
    int edit_any_file;
    int recursive_du;
    int largeset;
    char homedir[PATH_MAX];
    char homeurl[1024];
    char tagline[1024];
    char favicon[1024];
} cfg;

struct runtime_struct {
    char token[256];
    char iconsurl[64];
    char loggedinuser[64];
    int access_level;
    int access_as_user;
	int auth_method;
    int js;
} rt;

double t1, t2;
struct timeval mt;

enum { FALSE, TRUE };
enum { PERM_NO, PERM_RO, PERM_RW };
enum { CHANGE, DELETE, MOVE };
enum { AUTH_NONE, AUTH_IP, AUTH_MD5, AUTH_HT };

typedef struct asdir_ {
    char name[1024];
    mode_t type;
    off_t size;
    time_t atime, mtime, rtime;
} ASDIR;

int namesort(const void *, const void *);
int rnamesort(const void *, const void *);
int sizesort(const void *, const void *);
int rsizesort(const void *, const void *);
int timesort(const void *, const void *);
int rtimesort(const void *, const void *);
int asscandir(const char *, ASDIR **, int (*compar)(const void *, const void *));

void dbgprintf(char *, ...);
void error(char *, ...);
void redirect(char *, ...);
char *buprintf(float, int);
int strip(char *, int, char *);
int strsplit(char *, char ***, char *);
void checkfilename(char *);
void checkdestination(void); 
void mkfile(void);
void newdir(void);
void edit_save(void);
void delete(void);
void move(void);
void dirlist(void);
void edit_ui(void);
void rename_ui(void);
void mkdir_ui(void);
void mkfile_ui(void);
void multiprompt_ui(char *);
void about(void);
void save(void);
void receivefile(void);
void mkurl(void);
void goto_url(void);
off_t du(char *);
void re_dir_ui(char *, int);
int re_dir_up(char *);
void login_ui(void);
void tstop(void);
void html_title(char *);
void singleprompt_ui(char *);
char *url_encode(char *);
char *url_decode(char *);
int wfm_commit(int, char *);
int repo_check(void);
