# WFM - Web File Manager
WFM is a web file management application. It allows to perform regular
file and folder operations such as upload, download, rename, delete files
and organize directory tree structure using a standard web browser. It also
allows editing small text files directly in the browser using textarea.

Unlike Dropbox, box.net and others, WFM is compatible with most modern and
legacy web browsers dating back to Internet Explorer 1.5, Netscape 1.0 and
Mosaic 3.0. It outputs W3C certified HTML 4.01. JavaScript is optional and
only used for non-essential luxuries on modern web browsers.

This program is written using portable C code and compiles natively
for many flavors of Unix. It runs as a CGI application on most httpd servers
and does not require PHP, Perl, Python or any other interpreted language. 
It's very small and lightning fast. I runs on resource constrained embedded
or vintage / historical systems.


## History
The application begun its life in 1994 as a perl CGI script for CERN httpd
server to allow uploading and managing customer logs by field support
engineers over the web. Later rewritten in C language, when CGIC library and
Apache httpd were released. Up to 2015 WFM has been a closed source commercial
application, supported by a few large enterprise, telco and ISP corporations.
WFM is now released as open source.


## Installation
WFM binary is self-contained including all icons/images. You only need
to copy the compiled wfm binary (with any name) to your cgi execution
directory, usually cgi-bin. Include a config file of the same name as the
binary file plus .cfg extension. Example:

    /home/user/web/cgi-bin
      wfm
      wfm.cfg

Edit the .cfg file according to your needs.

Point your browser to http://x.x.x.x/cgi-bin/wfm and you are done.

## Configuration
This application was designed with multiple instances in mind. An instance
is made by copying or linking WFM binary with a different name. In a more
advanced configuration different instances can suexec to different users.

In basic form each instance reads it's configuration file of 
instance name + .cfg extension from the current working directory.
For instance if you decide to use "ftpadmin" as name of the executable
(or link) it will read file named "ftpadmin.cfg" for the configuration.
Below is a simple, self-explanatory configuration file example:

    # tagline or application name
    tagline=Snake Oil File Exchange

    # home directory, typically local directory on the server or SMB/NFS mount
    directory=/home/user/wfmdata

    # recursively calculate directory sizes - only enable if you have
    # fast disk (eg. SSD), large cache or a small directory tree structure
    # note that file and folder names starting with dot (.) are not counted
    # recursive-du=true

    # favicon / application icon, must be one of the embedded/compiled icon files
    # by default wfmicon.gif
    #favicon=home.gif

    # When clicking on txt file - edit instead of download by default
    #txt-default-edit=true

    # Edit any file as it was txt
    #edit-any-file=false

    # optional browser url prefix - aka "external link" - if defined, file
    # names will be glued to it giving option to be opened directly with the
    # external link button without going through cgi routines
    browser-url=http://x.x.x.x/files/

    # access lists - ace type is either access-ip or access-user
    # mixable, eg access-ip=ro:* with number of access-user=rw
    # level is ro|rw, one host or username per line  * denotes all hosts
    # user is username:password combination
    access-ip=ro:*
    access-ip=rw:127.0.0.1
    access-user=ro:guest:secret
    access-user=rw:admin:password

If you use mixed ro/rw access for instance ip=ro:* and user=rw:admin
then in order to authenticate click on the lock sign on right side of
the top status bar.

## Copyrights and Credits
Copyright (c) 1994-2016 by Antoni Sawicki  
Copyright (c) 1996-2011 by Thomas Boutell and Boutell.Com, Inc.  
Copyright (c) 2002 by Aladdin Enterprises  
Copyright (c) 1999-2009 by Paul Johnston  
Copyright (c) 2010 by Yusuke Kamiyamane  
WFM implemented by [Antoni Sawicki](http://www.tenox.net/)  
[CGIC Library](https://www.boutell.com/cgic/) by Thomas Boutell  
Server Side RFC 1321 implementation by L. Peter Deutsch  
Client Side RFC 1321 implementation by Paul Johnston  
Icons by [Yusuke Kamiyamane](http://p.yusukekamiyamane.com/)  
Web browser testing by [BrowserStack](http://www.browserstack.com/)