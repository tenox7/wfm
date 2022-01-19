# WFM - Web File Manager
WFM is a lightweight web based file manager. It allows to perform
regular file and folder operations such as upload, download, rename, delete files
and organize directory tree structure. Text, markup and markdown files can be
edited directly in the browser. WFM can also create and open bookmarks, link and
shortcut files, list inside archives and ISO files.

You can use WFM as a web interface foa NAS box, a "personal cloud", document
sharing site or a lighweight Content Management System (CMS). WFM can also serve
public, static html files from a selected directory which you can manage from the
private interface. See usage scenarios for more information.

WFM is written in Go language much like Docker, Kubernetes, Hugo, etc. The binary is
fully self contained and has zero dependencies. No need for Python, PHP, SQL, JavaScript,
Node or any other bloat. It's blazingly fast and the output is a pure, validated and
certified HTML 4.01 without JavaScript. It's tested on vide variety of modern and vintage
web browsers.

## Deployment scenarios

### Running as regular user

### Chroot and Setuid

## Certbot

## Authentication

## History
WFM begun its life around 1994 as a CGI Perl script for CERN httpd server to allow
uploading and managing customer logs by field support engineers over the web and a
front end to FTP servers. Later rewritten in C language, when CGIC library and
Apache httpd were released. Up to 2015 WFM has been a closed source commercial
application used for lighweight document management and supported by a few customers.
Later it has been open sourced. In 2022 WFM has been rewritten in Go as a stand-alone
application with built-in web server for more modern deployment styles.
