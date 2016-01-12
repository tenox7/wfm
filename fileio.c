// WFM File I/O Routines

#include "wfm.h"

//
// Send file to client browser 
// Called by cgiMain action=sendfile
//
void sendfile(void) {
    char buff[1024];
    FILE *in;
    int rd, tot, size, pos, blk;

    checkfilename(NULL);

    // TODO: 2gb file limit?
    in=fopen(phys_filename, "rb");
    if(!in)
        error("Unable to open file.<BR>%s", strerror(errno));

    fseek(in, 0, SEEK_END);
    size=ftell(in);
    fseek(in, 0, SEEK_SET);
    
    fprintf(cgiOut, 
        "Content-Type: application/octet-stream\r\n"
        "Content-Disposition: attachment; filename=\"%s\"; size=%d\r\n"
        "Content-Length: %d\r\n\r\n", 
        virt_filename, size, size
    );

    blk=sizeof(buff);

    reread:    
    pos=ftell(in);
    rd=fread(buff, blk, 1, in);
        
    if(rd) {
        tot=tot+rd*blk;
//      printf("rw=%u size=%u total=%u remaining=%u\n", rd*blk, size, tot, size-pos-(rd*blk));
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
    char *buff;
    int size=0, got=0;

    if(cgiFormFileSize("filename", &size) != cgiFormSuccess)
        error("No file size specified.");

    if(cgiFormFileOpen("filename", &input) != cgiFormSuccess) 
        error("Unable to access uploaded file.");

    buff=(char *) malloc(size+1);
    if(buff==NULL)
        error("Unable to allocate memory.");

    // TODO: check for freeze - return value?
    if(cgiFormFileRead(input, buff, size, &got) != cgiFormSuccess)
        error("Reading file.");
        
    cgiFormFileClose(input);

    if(got != size) 
        error("Wrong file size. Size=%d Received=%d.", size, got);

    checkfilename(NULL);

    output=fopen(phys_filename, "wb");
    if(!output) 
        error("Unable to open file %s for writing.<BR>%s", virt_filename, strerror(errno));

    if(fwrite(buff, size, 1, output) != 1) 
        error("While writing file.<BR>%s", strerror(errno));

    fclose(output);
    free(buff);
    
    redirect("?highlight=%s&directory=%s&token=%s", virt_filename, virt_dirname, token);

}

//
// Create a new new empty file
// Called by cgiMain action=mkfile
//
void mkfile(void) {
    FILE *output;

    checkfilename(NULL);

    output=fopen(phys_filename, "a");

    if(!output) 
        error("Unable to create file.<BR>%s", strerror(errno));

    fclose(output);
    
    redirect("?highlight=%s&directory=%s&token=%s", virt_filename, virt_dirname, token);

}

//
// Create a new empty folder
// Called by cgiMain action=newdir
//
void newdir(void) {

    checkfilename(NULL);

    if(mkdir(phys_filename, S_IRWXU | S_IRGRP | S_IXGRP | S_IROTH | S_IXOTH )!=0)
        error("Unable to create directory.<BR>%s", strerror(errno));
    
    redirect("?highlight=%s&directory=%s&token=%s", virt_filename, virt_dirname, token);

}

//
// Save file from textarea editor
// Called by cgiMain action=edit_save
//
void edit_save(void) {
    int size=0;
    int tmpfd;
    char *buff;
    char tempname[64];
    //FILE *output;
    FILE *tempf;
    char backup[4];
    char backup_filename[PHYS_DESTINATION_SIZE];
   	regex_t re;
	regmatch_t pmatch;
    struct stat tmpstat;


    checkfilename(NULL);

    // the size should be updated by onclick from content.value.lenght just before submission
    // it's used to verify that received data length is consistent with editor contents
    cgiFormInteger("size", &size, 0);

    if(size>100*1024*1024)
        error("Input size too large.");

    buff=(char *) malloc(size); 
    if(buff==NULL)
        error("Unable to allocate memory.");

    memset(buff, 0, size);
        
    cgiFormString("content", buff, size);

    if(strlen(buff)+1 != size) // +1 because size was also given +1 via front end
        error("Received wrong size. <BR>ContentLen=%d DataLen=%d. <BR> The file was not changed.", size, strlen(buff));

    // rename to .bak if requested
    cgiFormStringNoNewlines("backup", backup, sizeof(backup));

    if(strcmp(backup, "yes")==0) {
        regcomp(&re, "\\.(.+)$", REG_EXTENDED|REG_ICASE);
        if(regexec(&re, phys_filename, 1, &pmatch, 0)==0) {
            if(pmatch.rm_so+4 < PHYS_DESTINATION_SIZE) {
                strcpy(backup_filename, phys_filename);
                strcpy(backup_filename+pmatch.rm_so+1, "bak\0");
                if(rename(phys_filename, backup_filename)!=0)
                    error("Unable to create .bak file.<BR>%s was not modified.<BR>%s", virt_filename, strerror(errno));
            }
        }
    }

    // write to temporary file
    snprintf(tempname, sizeof(tempname), "%s/.wfmXXXXXX", phys_dirname);
    
    tmpfd=mkstemp(tempname);

    if(!tmpfd)
        error("Unable to create temporary file %s.<BR>%s", basename(tempname), strerror(errno));
    
    tempf=fdopen(tmpfd, "w");
    
    if(!tempf)
        error("Unable to open temporary file %s.<BR>%s", basename(tempname), strerror(errno));

    if(fwrite(buff, strlen(buff), 1, tempf) != 1)
        error("Unable to write to temporary file %s.<BR>%s", basename(tempname), strerror(errno));

    fclose(tempf);

    if(stat(tempname, &tmpstat)!=0)
        error("Unable to check temporary file size.<BR>%s<BR>%s", basename(tempname), strerror(errno));

    if(tmpstat.st_size != strlen(buff))
        error("Temprary file has a wrong length. Giving up.<BR>%s size=%d, buff len=%d", virt_filename, tmpstat.st_size);

    if(chmod(tempname, 00644)!=0)
        error("Unable to set file permissions.<BR>%s", strerror(errno));

    // finally rename to desination file
    if(rename(tempname, phys_filename)!=0)
        error("Unable to rename temp file.<BR>%s - %s<BR>%s<BR>", basename(tempname), virt_filename, strerror(errno));


    free(buff);

    redirect("?highlight=%s&directory=%s&token=%s", virt_filename, virt_dirname, token);

}

//
// Recursively Delete Folders - Internal Routine
// Called by fileio_delete when directory is encountered
//
void fileio_re_rmdir(char *dirname) {
    DIR *dir;
    struct dirent *direntry;
    struct stat fileinfo;
    char tempfullpath[PHYS_FILENAME_SIZE];

    dir=opendir(dirname);
    if(!dir) 
        error("Unable to remove directory..");


    direntry=readdir(dir);
    while(direntry!=0) {
        if(strncmp(direntry->d_name, ".", 1) && strncmp(direntry->d_name, "..", 2)) {
            snprintf(tempfullpath, PHYS_FILENAME_SIZE, "%s/%s", dirname, direntry->d_name);

            if(lstat(tempfullpath, &fileinfo)!=0)  
                error("Unable to get file status.<BR>%s", strerror(errno));

            if(S_ISDIR(fileinfo.st_mode)) {
                fileio_re_rmdir(tempfullpath); 
                if(rmdir(tempfullpath)!=0)
                    error("Unable to remove directory...<BR>%s", strerror(errno));
            } else {
                if(unlink(tempfullpath)!=0)
                    error("Unable to remove directory....<BR>%s", strerror(errno));
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

        if(lstat(phys_filename, &fileinfo)==0) {
            if(S_ISDIR(fileinfo.st_mode)) {
                fileio_re_rmdir(phys_filename);
                if(rmdir(phys_filename)!=0)
                    error("Unable to remove directory.<BR>%s", strerror(errno));
            }
            else {
                if(unlink(phys_filename)!=0) 
                    error("Unable to remove file.<BR>%s", strerror(errno));
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

    redirect("?directory=%s&token=%s", virt_dirname, token);

}

//
// Move File/Directory Internal Routine
// Called by move()
//
void fileio_move(void) {
    char final_destination[PHYS_DESTINATION_SIZE];
    struct stat fileinfo;

        // If moving file to a different directory we need to append the original file name
        if( stat(phys_destination, &fileinfo)==0 && S_ISDIR(fileinfo.st_mode) ) 
            snprintf(final_destination, sizeof(final_destination), "%s/%s", phys_destination, virt_filename);
        else
            strncpy(final_destination, phys_destination, sizeof(final_destination));
    
        if(rename(phys_filename, final_destination)!=0) 
            error("Unable to move file. <BR>[%d: %s]<BR>[SRC=%s] [DST=%s]", errno, strerror(errno), phys_filename, final_destination);
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

    redirect("?highlight=%s&directory=%s&token=%s", virt_destination, virt_dirname, token);
}


//
// Recursive Dir Size
//
// WARNING: will not count directories starting with .
off_t du(char *pdir) {
    DIR *dir;
    struct dirent *direntry;
    struct stat fileinfo;
    char child[PHYS_DIRNAME_SIZE];
    off_t tot=0;

    if(lstat(pdir, &fileinfo)==0)
        if(S_ISLNK(fileinfo.st_mode))
            return -1;

    dir=opendir(pdir);
    if(dir) {
        direntry=readdir(dir);
        while(direntry) {
            snprintf(child, PHYS_DIRNAME_SIZE, "%s/%s", pdir, direntry->d_name);
            if(lstat(child, &fileinfo)==0) {
                if(S_ISDIR(fileinfo.st_mode) && (direntry->d_name[0]!='.')) //TODO: count other ".files" except . & ..
                    tot+=du(child);
                else 
                    tot+=fileinfo.st_size;
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
	char child[VIRT_DIRNAME_SIZE];
	char phy_child[PHYS_DIRNAME_SIZE];
	char re_phys_dirname[PHYS_DIRNAME_SIZE];
	int n;
	int nentr, e;
	
	snprintf(re_phys_dirname, PHYS_DIRNAME_SIZE, "%s/%s", HOMEDIR, vdir);

	if(strlen(re_phys_dirname)<2 || strlen(re_phys_dirname)>(PHYS_DIRNAME_SIZE-2)) 
		error("Invalid directory name.");

	if(regexec(&dotdot, re_phys_dirname, 0, 0, 0)==0) error("Invalid directory name.");
	if(strlen(re_phys_dirname) < strlen(HOMEDIR)) error("Invalid directory name.");

    nentr=scandir(re_phys_dirname, &direntry, 0, alphasort);

    for(e=0; e<nentr; e++) {
        snprintf(phy_child, PHYS_DIRNAME_SIZE, "%s/%s/%s", HOMEDIR, vdir, direntry[e]->d_name);
        if((direntry[e]->d_name[0]!='.') && (lstat(phy_child, &fileinfo)==0) && S_ISDIR(fileinfo.st_mode))  {


            snprintf(child, VIRT_DIRNAME_SIZE, "%s/%s", vdir, direntry[e]->d_name);	

            fprintf(cgiOut, "<OPTION VALUE=\"%s\">", child);

            for (n=0; n<(level-1); n++)
                fprintf(cgiOut, "&nbsp;&nbsp;&nbsp;");

            fprintf(cgiOut, "&lfloor;&nbsp;%s</OPTION>\n", direntry[e]->d_name);

            // recurse
            re_dir_ui(child,level+1);
        }
        free(direntry[e]);
	}

}

//
// Scandir replacement function
//
int namesort(const void *d1, const void *d2) {
	return(strcmp( ((ASDIR*)d1)->name, ((ASDIR*)d2)->name));
}

int rnamesort(const void *d1, const void *d2) {
	return(strcmp( ((ASDIR*)d2)->name, ((ASDIR*)d1)->name));
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
    char filename[PATH_MAX];
    int entries=0;

	dirh=opendir(dir);
	if(dirh==NULL)
		return -1;
		
	names=(ASDIR*)malloc(sizeof(ASDIR));
	if(names==NULL)
		return -1;

    entry=readdir(dirh);
    while(entry!=NULL) {
        snprintf(filename, sizeof(filename), "%s/%s", dir, entry->d_name);
        if(stat(filename, &fileinfo)!=0)
            return -1;

        memset(&names[entries], 0, sizeof(ASDIR));
        strcpy(names[entries].name, entry->d_name);
        names[entries].type=fileinfo.st_mode;
        if(S_ISDIR(fileinfo.st_mode) && recursive_du)
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
        entry=readdir(dirh);
    }
    closedir(dirh);

    if(entries)
        qsort(&names[0], entries, sizeof(ASDIR), compar);

    *namelist=names;
	return entries;
}