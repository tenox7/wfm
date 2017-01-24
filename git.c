//
// File Version Control with GIT
//

// TODO: recursive delete

#ifdef WFMGIT
#include <git2.h>
#include "wfm.h"

enum { ADD, DEL };

git_repository *repo;

void wfm_git_stage(int op, char *filename) {
    git_index *index;
    
    git_repository_index(&index, repo);   
    
    if(op==ADD)
        git_index_add_bypath(index, filename);                   
    else if(op==DEL)
        git_index_remove_bypath(index, filename);                   
    else
        return;

    git_index_write(index);

    git_index_free(index);

    return;
}


void wfm_git_commit(char *msg) {
    git_oid tree_oid, head_oid, commit_oid;
    git_tree *tree;
    git_commit *head_commit;
    git_index *index;
    git_signature *sig;
    int ret;
    char username[1024]={0};
    char email[1024]={0};

    snprintf(username, sizeof(username), "%s %s", basename(cgiScriptName), (strlen(loggedinuser)) ? loggedinuser:"(none)");
    snprintf(email, sizeof(email), "%s@%s.wfm", (strlen(loggedinuser)) ? loggedinuser:"(none)", cgiRemoteAddr);

    git_signature_now(&sig, username, email);        
    

    git_repository_index(&index, repo);                     
    git_index_write_tree(&tree_oid, index);                  
    git_tree_lookup(&tree, repo, &tree_oid);                 
    ret=git_reference_name_to_id(&head_oid, repo, "HEAD");     

    if(ret) {
        git_commit_create_v(&commit_oid, repo, "HEAD", sig, sig, NULL, msg, tree, 0); 
    }
    else {
        git_commit_lookup(&head_commit, repo, &head_oid);           
        git_commit_create_v(&commit_oid, repo, "HEAD", sig, sig, NULL, msg, tree, 1, head_commit); 
    }
    
    git_index_write(index);

    git_tree_free(tree);
    git_index_free(index);
    git_signature_free(sig);
    return;
}
#endif

int wfm_commit(int op, char *fname) {
#ifdef WFMGIT
    int ret;
    char repodir[sizeof(HOMEDIR)+10]={0};
    char msg[1024];
    char stage_filename_buf[1024]={0};
    char *stage_filename=stage_filename_buf;
    char stage_newname_buf[1024]={0};
    char *stage_newname=stage_newname_buf;
    char *opstr[]={ "Change", "Delete", "Move" };

    if(op>2)
        return 0;

	git_libgit2_init();

    snprintf(repodir, sizeof(repodir), "%s/.git", HOMEDIR);
    ret=git_repository_open(&repo, repodir);                
    if(ret)
        return ret;

    if(fname && strlen(fname)) 
        strncpy(stage_filename_buf, fname, sizeof(stage_filename_buf));
    else
        strncpy(stage_filename_buf, phys_filename, sizeof(stage_filename_buf));
    stage_filename+=strlen(HOMEDIR);
    while(*stage_filename=='/')
        stage_filename++;

    if(op==MOVE) {
        strncpy(stage_newname_buf, final_destination, sizeof(stage_newname_buf));
        stage_newname+=strlen(HOMEDIR);
        while(*stage_newname=='/')
            stage_newname++;
    }

    snprintf(msg, sizeof(msg), 
        "WFM %s Commit: "
        "Filename=[%s%s%s] "
        "Instance=[%s] "
        "User=[%s] "
        "RemoteIP=[%s]\n", 
        opstr[op], 
        stage_filename,
        (op==MOVE) ? " => " : "",
        (op==MOVE) ? stage_newname : "",
        basename(cgiScriptName),
        (strlen(loggedinuser)) ? loggedinuser:"(none)",
        cgiRemoteAddr
    );

    if(op==CHANGE) {
        wfm_git_stage(ADD, stage_filename);
    }
    else if(op==DELETE) {
        wfm_git_stage(DEL, stage_filename);
    }
    else if(op==MOVE) {
        wfm_git_stage(DEL, stage_filename);
        wfm_git_stage(ADD, stage_newname);
    }

    wfm_git_commit(msg);
    git_repository_free(repo);
#endif
    return 0;
}



int repo_check() {
    int ret=1;
#ifdef WFMGIT
    char repodir[sizeof(HOMEDIR)+10]={0};

	git_libgit2_init();

    snprintf(repodir, sizeof(repodir), "%s/.git", HOMEDIR);
    ret=git_repository_open(&repo, repodir);                
    if(ret==0)
        git_repository_free(repo);

#endif
    return ret;
}
