# WFM - Web File Manager
WFM is a lightweight web based file manager. It allows to perform
regular file and folder operations such as upload, download, rename, delete files
and organize directory tree structure. Text, markup and markdown files can be
edited directly in the browser. WFM can also create and open bookmarks, link and
shortcut files, list inside archives and ISO files.

You can use WFM as a web interface for a NAS box, a "personal cloud", document
sharing site or a lightweight Content Management System (CMS). WFM can also serve
public, static html files from a selected directory which you can manage from the
private interface. See usage scenarios for more information.

WFM is a standalone service with it's own web server. It runs from systemd, sysvinit,
launchd, bsd rc or Docker. TLS/SSL is supported with automatic certificate generation
by Lets Encrypt / Certbot / ACME.

Written in Go language, much like Docker, Kubernetes, Hugo, etc. The binary is
fully self contained and has zero dependencies. No need for Python, PHP, SQL, JavaScript,
Node or any other bloat. WFM works on both modern and old web browsers going back to
Internet Explorer 2.x and Netscape 3.x. It outputs validated HTML 4.01 without JavaScript.

## Deployment scenarios

### Init

### Docker

### Running as regular user

### Chroot and Setuid

## Auto Cert Manager

## Authentication

### Json password file

### Binary hardcoded

### Generate password hash

```sh
$ echo -n "SaltMyPassword" | shasum -a 256 | cut -f 1 -d" "
```

## History
WFM begun its life around 1994 as a CGI Perl script for CERN httpd server, to allow
uploading and managing customer logs by field support engineers over the web and
as a front end to FTP server. Later rewritten in C language, when CGIC library and
Apache httpd were released. Up to 2015 WFM has been a closed source commercial
application used for lightweight document management and supported by a few customers.
It has been open sourced. In 2022 WFM has been rewritten in Go as a stand-alone
application with built-in web server for more modern deployment styles.
