// WFM File I/O Routines

#include "wfm.h"

/*
// Debug dump vars
//void debugdumpvars(void) {

    cgiHeaderContentType("text/plain");

    printf(
        "virt_dirname=%s\n"
        "wp.phys_dirname=%s\n"
        "wp.virt_filename=%s\n"
        "wp.phys_filename=%s\n"
        "wp.virt_destination=%s\n"
        "wp.phys_destination=%s\n"
  //      "wp.final_destination=%s\n"
        "virt_parent=%s\n",
        virt_dirname,
        wp.phys_dirname,
        wp.virt_filename,
        wp.phys_filename,
        wp.virt_destination,
        wp.phys_destination,
    //    wp.final_destination,
        virt_parent
   );

    exit(1);
}
*/

//
// Send file to client browser 
// Called by cgiMain action=sendfile
//
void sendfile(void) {
    char buff[1024]={0};
    FILE *in;
    int rd=0, tot=0, size=0, pos=0, blk=0;

    checkfilename(NULL);

    // TODO: 2gb file limit?
    in=fopen(wp.phys_filename, "rb");
    if(!in)
        error("Unable to open file.<BR>%s", strerror(errno));

    fseek(in, 0, SEEK_END);
    size=ftell(in);
    fseek(in, 0, SEEK_SET);
    
    fprintf(cgiOut, 
        "Content-Type: application/octet-stream\r\n"
        "Content-Disposition: attachment; filename=\"%s\"; size=%d\r\n"
        "Content-Length: %d\r\n\r\n", 
        wp.virt_filename, size, size
    );

    blk=sizeof(buff);

    reread:    
    pos=ftell(in);
    rd=fread(buff, blk, 1, in);
        
    if(rd) {
        tot+=rd*blk;
      //fprintf(cgiOut, "rw=%u size=%u total=%u remaining=%u\n", rd*blk, size, tot, size-pos-(rd*blk));
        fwrite(buff, blk, 1, cgiOut);
        goto reread;
    }

    if(pos<size) {
        blk=size-pos;
        fseek(in, pos, SEEK_SET);
        goto reread;
    }

    fclose(in);

}

//
// Receive file from client browser via upload form
// Called by cgiMain action=upload
//
void receivefile(void) {
    cgiFilePtr input;
    FILE *output;
    char buff[1024]={0};
    int got=0; //,size=0;

    checkfilename(NULL);

    //if(cgiFormFileSize("filename", &size) != cgiFormSuccess)
    //    error("No file size specified.");

    if(cgiFormFileOpen("filename", &input) != cgiFormSuccess) 
        error("Unable to access uploaded file.");

    output=fopen(wp.phys_filename, "wb");
    if(!output) 
        error("Unable to open file %s for writing.<BR>%s", wp.virt_filename, strerror(errno));

    if(flock(fileno(output), LOCK_EX) == -1)
        error("Unable to lock file %s.<BR>%s", wp.virt_filename, strerror(errno));

    while(cgiFormFileRead(input, buff, sizeof(buff), &got) == cgiFormSuccess) 
        if(got)
            if(fwrite(buff, got, 1, output) != 1) 
                error("While writing file.<BR>%s", strerror(errno));

    cgiFormFileClose(input);
    fclose(output);

    wfm_commit(CHANGE, NULL);
    
    redirect("%s?highlight=%s&directory=%s&rt.token=%s", cgiScriptName, wp.virt_filename_urlencoded, wp.virt_dirname_urlencoded, rt.token);

}

//
// Create a new new empty file
// Called by cgiMain action=mkfile
//
void mkfile(void) {
    FILE *output;

    checkfilename(NULL);

    output=fopen(wp.phys_filename, "a"); //TODO: should probably give error if file already exists...

    if(!output) 
        error("Unable to create file.<BR>%s", strerror(errno));

    fclose(output);

    wfm_commit(CHANGE, NULL);

    redirect("%s?highlight=%s&directory=%s&rt.token=%s", cgiScriptName, wp.virt_filename_urlencoded, wp.virt_dirname_urlencoded, rt.token);

}

//
// Create a new empty folder
// Called by cgiMain action=newdir
//
void newdir(void) {

    checkfilename(NULL);

    if(mkdir(wp.phys_filename, S_IRWXU | S_IRGRP | S_IXGRP | S_IROTH | S_IXOTH )!=0)
        error("Unable to create directory.<BR>%s", strerror(errno));

    redirect("%s?highlight=%s&directory=%s&rt.token=%s", cgiScriptName, wp.virt_filename_urlencoded, wp.virt_dirname_urlencoded, rt.token);

}

//
// Save file from textarea editor
// Called by cgiMain action=edit_save
//
void edit_save(void) {
    int size=0;
    int tmpfd;
    char *buff;
    char tempname[64]={0};
    //FILE *output;
    FILE *tempf;
#ifndef WFMGIT
    char backup[4]={0};
    char backup_filename[sizeof(wp.phys_dirname)]={0};
    regex_t re;
    regmatch_t pmatch;
#endif
    struct stat tmpstat;

    checkfilename(NULL);

    cgiFormStringSpaceNeeded("content", &size);

    if(size>=5*1024*1024)
        error("The file is too large for online editing.<BR>");

    buff=(char *) malloc(size); 
    if(buff==NULL)
        error("Unable to allocate memory.");

    memset(buff, 0, size);
        
    cgiFormString("content", buff, size);

#ifndef WFMGIT
    // rename to .bak if requested
    cgiFormStringNoNewlines("backup", backup, sizeof(backup));

    if(strcmp(backup, "yes")==0) {
        regcomp(&re, "\\.(.+)$", REG_EXTENDED|REG_ICASE);
        if(regexec(&re, wp.phys_filename, 1, &pmatch, 0)==0) {
            if(pmatch.rm_so+4 < sizeof(wp.phys_dirname)) {
                strcpy(backup_filename, wp.phys_filename);
                strcpy(backup_filename+pmatch.rm_so+1, "bak\0");
                if(rename(wp.phys_filename, backup_filename)!=0)
                    error("Unable to create .bak file.<BR>%s was not modified.<BR>%s", wp.virt_filename, strerror(errno));
            }
        }
    }
#endif
    // write to temporary file
    snprintf(tempname, sizeof(tempname), "%s/.wfmXXXXXX", wp.phys_dirname);
    
    tmpfd=mkstemp(tempname);

    if(!tmpfd)
        error("Unable to create temporary file %s.<BR>%s", basename(tempname), strerror(errno));

    if(chmod(tempname, 00644)!=0)
        error("Unable to set file permissions.<BR>%s", strerror(errno));

    tempf=fdopen(tmpfd, "w");
    
    if(!tempf)
        error("Unable to open temporary file %s.<BR>%s", basename(tempname), strerror(errno));

    if(flock(fileno(tempf), LOCK_EX) == -1)
        error("Unable to lock file %s.<BR>%s", basename(tempname), strerror(errno));

    if(fwrite(buff, strlen(buff), 1, tempf) != 1)
        error("Unable to write to temporary file %s.<BR>%s", basename(tempname), strerror(errno));

    fclose(tempf);

    if(stat(tempname, &tmpstat)!=0)
        error("Unable to check temporary file size.<BR>%s<BR>%s", basename(tempname), strerror(errno));

    if(tmpstat.st_size != strlen(buff))
        error("Temprary file has a wrong length. Giving up.<BR>%s size=%d, buff len=%d", wp.virt_filename, tmpstat.st_size);

    // finally rename to desination file
    if(rename(tempname, wp.phys_filename)!=0)
        error("Unable to rename temp file.<BR>%s - %s<BR>%s<BR>", basename(tempname), wp.virt_filename, strerror(errno));

    free(buff);

    wfm_commit(CHANGE, NULL);

    redirect("%s?highlight=%s&directory=%s&rt.token=%s", cgiScriptName, wp.virt_filename_urlencoded, wp.virt_dirname_urlencoded, rt.token);
}

//
// Recursively Delete Folders - Internal Routine
// Called by fileio_delete when directory is encountered
//
void fileio_re_rmdir(char *dirname) {
    DIR *dir;
    struct dirent *direntry;
    struct stat fileinfo;
    char tempfullpath[sizeof(wp.phys_filename)]={0};

    dir=opendir(dirname);
    if(!dir) 
        error("Unable to remove directory..");


    direntry=readdir(dir);
    while(direntry!=0) {
        if(strncmp(direntry->d_name, ".", 1) && strncmp(direntry->d_name, "..", 2)) {
            snprintf(tempfullpath, sizeof(wp.phys_filename), "%s/%s", dirname, direntry->d_name);

            if(lstat(tempfullpath, &fileinfo)!=0)  
                error("Unable to get file status.<BR>%s", strerror(errno));

            if(S_ISDIR(fileinfo.st_mode)) {
                fileio_re_rmdir(tempfullpath); 
                if(rmdir(tempfullpath)!=0)
                    error("Unable to remove directory...<BR>%s", strerror(errno));
            } else {
                if(unlink(tempfullpath)!=0)
                    error("Unable to remove file....<BR>%s", strerror(errno));
                wfm_commit(DELETE, tempfullpath);
            }
            
        }
        direntry=readdir(dir);
    }

    closedir(dir);

}

//
// Delete Files Internal Routine
// Called by delete()
//
void fileio_delete(void) {
        struct stat fileinfo;

        if(lstat(wp.phys_filename, &fileinfo)==0) {
            if(S_ISDIR(fileinfo.st_mode)) {
                fileio_re_rmdir(wp.phys_filename);
                if(rmdir(wp.phys_filename)!=0)
                    error("Unable to remove directory.<BR>%s", strerror(errno));
            }
            else {
                if(unlink(wp.phys_filename)!=0) 
                    error("Unable to remove file.<BR>%s", strerror(errno));

                wfm_commit(DELETE, NULL);
            }
        }

}

//
// Delete File or Directory Handler - Allows Multiselect 
// Called by cgiMain action=delete
//
void delete(void) {
    int i;
    char **responses; 

    // Single
    if(cgiFormStringMultiple("multiselect_filename", &responses) == cgiFormNotFound) {  
        checkfilename(NULL);
        fileio_delete(); 
    } 
    // Multi
    else {
        for(i=0; responses[i]; i++) {
            checkfilename(responses[i]);
            fileio_delete();
        }
    }           

    redirect("%s?directory=%s&rt.token=%s", cgiScriptName, wp.virt_dirname_urlencoded, rt.token);
}

//
// Move File/Directory Internal Routine
// Called by move()
//
void fileio_move(void) {
    struct stat fileinfo;

    // If moving file to a different directory we need to append the original file name to destination
    if( stat(wp.phys_destination, &fileinfo)==0 && S_ISDIR(fileinfo.st_mode) )  
        snprintf(wp.final_destination, sizeof(wp.final_destination), "%s/%s", wp.phys_destination, wp.virt_filename);
    else 
        strncpy(wp.final_destination, wp.phys_destination, sizeof(wp.final_destination));
    
    if(rename(wp.phys_filename, wp.final_destination)!=0) 
        error("Unable to move file. <BR>[%d: %s]<BR>[SRC=%s] [DST=%s]", errno, strerror(errno), wp.phys_filename, wp.final_destination);

    wfm_commit(MOVE, NULL);

}

// 
// Move File/Directory - Handler
// Called by cgiMain action=move
//
void move(void) {
    int i;
    char **responses; 

    checkdestination();
   
    // Single
    if(cgiFormStringMultiple("multiselect_filename", &responses) == cgiFormNotFound) {  
        checkfilename(NULL);
        fileio_move();
    } 
    // Multi
    else {
        for(i=0; responses[i]; i++) {
            checkfilename(responses[i]);
            fileio_move();
        }
    }           

    redirect("%s?highlight=%s&directory=%s&rt.token=%s", cgiScriptName, url_encode(wp.virt_destination), wp.virt_dirname_urlencoded, rt.token);
}


//
// Recursive Dir Size
//
off_t du(char *pdir) {
    DIR *dir;
    struct dirent *direntry;
    struct stat fileinfo;
    char child[sizeof(wp.phys_dirname)]={0};
    off_t tot=0;

    if(lstat(pdir, &fileinfo)==0)
        if(S_ISLNK(fileinfo.st_mode))
            return -1;

    dir=opendir(pdir);
    if(dir) {
        direntry=readdir(dir);
        while(direntry) {
            snprintf(child, sizeof(wp.phys_dirname), "%s/%s", pdir, direntry->d_name);
            if(lstat(child, &fileinfo)==0) {
                if(S_ISDIR(fileinfo.st_mode)) {
                    if(direntry->d_name[0]=='.' && direntry->d_name[1]=='\0') 
                        ;
                    else if(direntry->d_name[0]=='.' && direntry->d_name[1]=='.' && direntry->d_name[2]=='\0') 
                        ;
                    else
                        tot+=du(child);
                }
                else {
                    tot+=fileinfo.st_size;
                }
            }
            direntry=readdir(dir);
        }
        closedir(dir);
    }   

    return tot;
}

//
// Recursive folder list
// Called by for move_ui()
//
void re_dir_ui(char *vdir, int level) {
    struct dirent **direntry;
    struct stat fileinfo;
    char child[sizeof(wp.virt_dirname)]={0};
    char phy_child[sizeof(wp.phys_dirname)]={0};
    char re_phys_dirname[sizeof(wp.phys_dirname)]={0};
    int n;
    int nentr, e;
    
    snprintf(re_phys_dirname, sizeof(re_phys_dirname), "%s/%s", cfg.homedir, vdir);

    if(strlen(re_phys_dirname)<2 || strlen(re_phys_dirname)>(sizeof(wp.phys_dirname)-2)) 
        error("Invalid directory name.");

    if(regexec(&dotdot, re_phys_dirname, 0, 0, 0)==0) error("Invalid directory name.");
    if(strlen(re_phys_dirname) < strlen(cfg.homedir)) error("Invalid directory name.");

    nentr=scandir(re_phys_dirname, &direntry, 0, alphasort);

    for(e=0; e<nentr; e++) {
        snprintf(phy_child, sizeof(phy_child), "%s/%s/%s", cfg.homedir, vdir, direntry[e]->d_name);
        if((direntry[e]->d_name[0]!='.') && (lstat(phy_child, &fileinfo)==0) && S_ISDIR(fileinfo.st_mode))  {


            snprintf(child, sizeof(wp.virt_dirname), "%s/%s", vdir, direntry[e]->d_name); 

            fprintf(cgiOut, "<OPTION VALUE=\"%s\">", child);

            for (n=0; n<(level-1); n++)
                fprintf(cgiOut, "&nbsp;&nbsp;&nbsp;");

            fprintf(cgiOut, "%s&nbsp;%s</OPTION>\n", (rt.js) ? "&boxvr;" : "-", direntry[e]->d_name);

            // recurse
            if(!cfg.largeset)
                re_dir_ui(child,level+1);
        }
        free(direntry[e]);
    }

}

//
// Display directory up tree used for file move with large file set
//
int re_dir_up(char *vdir) {
    int n,nn,m,len;
    char **dirs;
    char tmp[sizeof(wp.virt_dirname)]={0};

    strcpy(tmp, vdir);
    len=strsplit(tmp, &dirs, "/.");
    for(n=0; n<len; n++) {
        fprintf(cgiOut, "<OPTION VALUE=\"/");
        
        for(nn=0; nn<n+1; nn++)
            fprintf(cgiOut, "%s/", dirs[nn]);

        fprintf(cgiOut, "\">");

        for(m=0; m<n; m++)
            fprintf(cgiOut, "&nbsp;&nbsp;&nbsp;");

        fprintf(cgiOut, "&boxvr; %s</OPTION>\n", dirs[n]);

    }

    
    return n+1;
}

//
// Scandir replacement function
//
int namesort(const void *d1, const void *d2) {
    return(strcasecmp(((ASDIR*)d1)->name, ((ASDIR*)d2)->name));
}

int rnamesort(const void *d1, const void *d2) {
    return(strcasecmp(((ASDIR*)d2)->name, ((ASDIR*)d1)->name));
}

int sizesort(const void *d1, const void *d2) {
         if(((ASDIR*)d1)->size < ((ASDIR*)d2)->size) return -1;
    else if(((ASDIR*)d1)->size > ((ASDIR*)d2)->size) return 1;
    else return 0;
}

int rsizesort(const void *d1, const void *d2) {
         if(((ASDIR*)d1)->size > ((ASDIR*)d2)->size) return -1;
    else if(((ASDIR*)d1)->size < ((ASDIR*)d2)->size) return 1;
    else return 0;
}

int timesort(const void *d1, const void *d2) {
         if(((ASDIR*)d1)->mtime < ((ASDIR*)d2)->mtime) return -1;
    else if(((ASDIR*)d1)->mtime > ((ASDIR*)d2)->mtime) return 1;
    else return 0;
}

int rtimesort(const void *d1, const void *d2) {
         if(((ASDIR*)d1)->mtime > ((ASDIR*)d2)->mtime) return -1;
    else if(((ASDIR*)d1)->mtime < ((ASDIR*)d2)->mtime) return 1;
    else return 0;
}

int asscandir(const char *dir, ASDIR **namelist, int (*compar)(const void *, const void *)) {
    DIR *dirh;
    ASDIR *names;
    struct dirent *entry;
    struct stat fileinfo;
    char filename[PATH_MAX]={0};
    int entries=0;

    dirh=opendir(dir);
    if(dirh==NULL)
        return -1;
        
    names=(ASDIR*)malloc(sizeof(ASDIR));
    if(names==NULL)
        return -1;

    entry=readdir(dirh);
    while(entry!=NULL) {
        if(entry->d_name[0]!='.') {
            snprintf(filename, sizeof(filename), "%s/%s", dir, entry->d_name);
            if(stat(filename, &fileinfo)!=0) {
                entry=readdir(dirh);
                continue;
            }

            memset(&names[entries], 0, sizeof(ASDIR));
            strcpy(names[entries].name, entry->d_name);
            names[entries].type=fileinfo.st_mode;
            if(S_ISDIR(fileinfo.st_mode) && cfg.recursive_du)
                names[entries].size=du(filename);
            else            
                names[entries].size=fileinfo.st_size;
            names[entries].atime=fileinfo.st_atime;
            names[entries].mtime=fileinfo.st_mtime;
            names[entries].rtime=fileinfo.st_ctime;
            
            names=(ASDIR*)realloc((ASDIR*)names, sizeof(ASDIR)*(entries+2));
            if(names==NULL)
                return -1;
            entries++;
        }
        entry=readdir(dirh);
    }
    closedir(dirh);

    if(entries)
        qsort(&names[0], entries, sizeof(ASDIR), compar);

    *namelist=names;
    return entries;
}
