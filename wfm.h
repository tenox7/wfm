#define VERSION "1.2.3"
#define copyright "<!-- WFM Version " VERSION ", Mountain View, CA, " __DATE__ " [" __TIME__ "] -->\n" \
                  "<!-- Copyright (c) 1994-2017 by Antoni Sawicki -->\n"

#define CSS_STYLE         \
            "  <STYLE TYPE=\"text/css\"><!-- \n" \
            "    A:link {text-decoration: none; color:#0000CE; } \n" \
            "    A:visited {text-decoration: none; color:#0000CE; } \n" \
            "    A:active {text-decoration: none; color:#FF0000; } \n" \
            "    A:hover {text-decoration: none; color:#FF0000; } \n" \
            "    body, td, th, input { font-family: Tahoma, Sans-Serif; font-size:11px;  } \n" \
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
        copyright \
        "<HEAD>\n" \
        "  <META HTTP-EQUIV=\"Content-Type\" CONTENT=\"text/html;charset=US-ASCII\">\n" \
        "  <META HTTP-EQUIV=\"Content-Language\" CONTENT=\"en-US\">\n" \
        "  <META HTTP-EQUIV=\"google\" CONTENT=\"notranslate\">\n" \
        "  <META NAME=\"viewport\" CONTENT=\"width=device-width\">\n" \
        CSS_STYLE


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
//#include <sys/dir.h>
#include "md5.h"
#include "cgic.h"
#include "wfmiconres.h"

#define VALIDCHRS "an ()[]{}-_.,!@#$%^&=+;"
char VALIDCHRS_DIR[256]; // above + /


#define P1024_1 1024.0f
#define P1024_2 1048576.0f
#define P1024_3 1073741824.0f
#define P1024_4 1099511627776.0f

#define SHM_SIZE 16

#define VIRT_DIRNAME_SIZE NAME_MAX  // around 255
#define PHYS_DIRNAME_SIZE 1024
#define VIRT_FILENAME_SIZE NAME_MAX
#define PHYS_FILENAME_SIZE 1280
#define VIRT_DESTINATION_SIZE NAME_MAX
#define PHYS_DESTINATION_SIZE 1280

char virt_dirname[VIRT_DIRNAME_SIZE];
char *virt_dirname_urlencoded;
char phys_dirname[PHYS_DIRNAME_SIZE]; 
char virt_filename[VIRT_FILENAME_SIZE]; 
char *virt_filename_urlencoded; 
char phys_filename[PHYS_FILENAME_SIZE];
char virt_destination[VIRT_DESTINATION_SIZE]; 
char phys_destination[PHYS_DESTINATION_SIZE];
char final_destination[PHYS_DESTINATION_SIZE];
char virt_parent[VIRT_DIRNAME_SIZE];
char *virt_parent_urlencoded;

char ICONSURL[1024];
char HOMEDIR[1024];
char HOMEURL[1024];
char TAGLINE[1024];
char FAVICON[1024];

char token[256];
char loggedinuser[64];

regex_t dotdot;
int access_level;
int access_as_user;
int users_defined;
int edit_by_default;
int edit_any_file;
int recursive_du;

int js;

double t1, t2;
struct timeval mt;

enum { FALSE, TRUE };
enum { PERM_NO, PERM_RO, PERM_RW };
enum { CHANGE, DELETE, MOVE };

typedef struct asdir_ {
    char name[NAME_MAX];
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

void error(char *, ...);
void redirect(char *, ...);
char *buprintf(float, int);
int strip(char *, int, char *);
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
void sendfile(void);
void receivefile(void);
off_t du(char *);
void re_dir_ui(char *, int);
void login_ui(void);
void tstop(void);
void html_title(char *);
void singleprompt_ui(char *);
char *url_encode(char *);
char *url_decode(char *);
int wfm_commit(int, char *);
int repo_check(void);
