# WFM - Web File Manager
WFM is a lightweight web based file manager. It allows to perform
regular file and folder operations such as upload, download, rename, delete files
and organize directory tree structure. Text, markup and markdown files can be
edited directly in the browser. WFM can also create and open bookmarks, link and
shortcut files, list inside archives and ISO files.

![wfm screenshot](screenshot.png "WFM Screenshot")

You can use WFM as a web interface for a NAS box, a "personal cloud", document
sharing site or a lightweight Content Management System (CMS). WFM can also serve
public, static html files from a selected directory which you can manage from the
private interface. See usage scenarios for more information.

WFM is a standalone service with it's own web server. No need for Apache, Nginx or
anything else. It directly runs from systemd, sysvinit, launchd, bsd rc or Docker.
TLS/SSL is supported with automatic certificate generation by Lets Encrypt / Certbot.

Written in Go language, much like Docker, Kubernetes, Hugo, etc. The binary is statically
linked, fully self contained and has zero external dependencies. Icons are Unicode
emojis. CA Certs are embedded at built time. No need for Python, PHP, SQL, JavaScript,
Node or any other bloat. WFM works on both modern and legacy web browsers going back to
Internet Explorer 2.x and Netscape 3.x. It outputs validated HTML 4.01 without JavaScript.

## Deployment scenarios

WFM relies on chroot for limiting which directory to use. Chroot can be set by WFM own
`-chroot=/dir` flag or by Systemd `RootDirectory=`. Also depending on what port you want
WFM to listen to (eg 80/443 vs 8080) you need to run it as root or regular user. If ran
by root WFM supports flag `-setuid=<user>` to setuid after port bind is complete.

### Systemd

An example service file is provided [here](service/systemd/wfm80.service). By default it
starts the process as root to allow to bind to port 80. You can specify destination
directory in `-chroot=/datadir` and user to run as in `-setuid=myuser`. WFM will
automatically chroot and setuid after port bind is complete.

You can specify Systemd `User=` other than root if you also use `RootDirectory=` for
chroot and use non privileged port (above 1024, eg 8080), or your binary has adequate
capabilities set. Example [here](service/systemd/wfm8080.service).


### Docker

TBD

## SSL / TLS / Auto Cert Manager

You can use WFM as a SSL / TLS / https secure web server with Lets Encrypt Auto Cert Manager.
ACM will automatically obtain SSL certificate for your site as well as the keypair.

Example deployment with SSL:

```text
ExecStart=/usr/local/sbin/wfm \
	-passwd=/usr/local/etc/wfmpasswd.json \
	-chroot=/home/user \
	-setuid=user \
	-addr=:443 \
	-acm_addr=:80 \
	-acm_dir=/.certs \
	-acm_host=www.snakeoil.com
```

The flag `-addr=:443` makes WFM listen on port 443 for https requests.
Flag `-acm_addr=:80` is used for Auto Cert Manager to obtain the cert
and then redirect to port 443/https. `-acm_dir=/.certs` is where the
certificate and key are stored. This directory is inside chroot jail
and currently accessible to users (TODO: fix this) so insecure. The
`-acm_host=` is a repeated flag that adds specific host to a whitelist.
ACM will only obtain certificates for whitelisted hosts. If your WFM
site has multiple names in DNS you need to add them to the whitelist.

If the https site is exposed externally outside of your firewall its
sometimes desired to have a local http (non-SSL) listener as well. To
enable this use `-addr_extra=:8080` flag.

## Authentication

If no password file is specified and no hardcoded passwords are present
WFM will not ask for password. By default it will be in read-only mode
unless you specify `-nopass_rw` flag. The password file can be specified
via `-passwd=/path/users.json` flag. Passwords are read on startup and
therefore can be placed outside of chroot directory.

Passwords can also be hardcoded in the binary, se below.

### Json password file

An example file is [provided](users.json). The format is a simple list of
users with "User", "Salt", "Hash" strings and "RW" boolean field. User
is self explanatory. Salt is a short random string used to make passwords
harder to crack. It can be anything but it must be different for every user.
The same salt must also be passed when generating the password. Hash is
a hashed salt + password string. RW boolean specifies if user has read only
or read write access.

### Binary hardcoded

Password file can also be hardcoded inside the binary at compile time.
To add hardcoded users add entries in to `users` var in `auth.go`.

### Generating password hash

```sh
$ echo -n "SaltMyPassword" | shasum -a 256 | cut -f 1 -d" "
```

### Example adding user

For example you want to add user `customer` with password `gh34j3n1`.

Add a new entry in the json file. Pick a unique salt, eg `zzx`:

```json
[
  { "User": "customer", "Salt": "zzx", "Hash": "", "RW": true }
]
```

Run:

```sh
$ echo -n "zzxgh34j3n1" | shasum -a 256 | cut -f 1 -d" "
```

Get the encoded string and paste it into Hash: "" value.

## Prefix

By default WFM serves requests from "/" prefix of the built in web server.
You can move it to a different prefix for example "/data" or "/wfm" with the
flag `-prefix=/pfx`.

## Doc dir

In addition to it's own Web UI, WFM can also act as a simple web server for
static html files, etc. To enable this you can use `-doc_srv=/var/www/html:/docs`
flag. You can serve it on `/` prefix if you move WFM prefix to another location
via `-prefix`. The physical directory is inside chroot jail.

With this you can create a trivial content management server. For example:

```shell
$ wfm \
  -doc_srv=/:/ \
  -prefix=/admin \
  -passwd=/path/users.json /
  -chroot=/somedir
```

In this example WFM will serve html files from `/somedir` on / http prefix
with `/admin` as a password protected admin interface where you can edit
and manage the site.

## Flags

```text
Usage of ./wfm:
  -about_runtime
        Display runtime info in About Dialog (default true)
  -acm_addr string
        autocert manager listen address, eg: :80
  -acm_dir string
        autocert cache, eg: /var/cache (affected by chroot)
  -acm_host value
        autocert manager allowed hostnames
  -addr string
        Listen address, eg: :443 (default "127.0.0.1:8080")
  -addr_extra string
        Extra non-TLS listener address, eg: :8081
  -allow_root
        allow to run as uid=0/root without setuid
  -cache_ctl string
        HTTP Header Cache Control (default "no-cache")
  -chroot string
        Directory to chroot to
  -doc_srv string
        Serve regular http files, fsdir:prefix, eg /var/www:/home
  -f2b
        ban ip addresses on user/pass failures (default true)
  -f2b_dump string
        enable f2b dump at this prefix, eg. /f2bdump (default no)
  -logfile string
        Log file name (default stdout)
  -nopass_rw
        allow read-write access if there is no password file
  -passwd string
        wfm password file, eg: /usr/local/etc/wfmpw.json
  -prefix string
        Default prefix for WFM access (default "/")
  -proto string
        tcp, tcp4, tcp6, etc (default "tcp")
  -setuid string
        Username to setuid to
  -show_dot
        show dot files and folders
```

## History
WFM begun its life around 1994 as a CGI Perl script for CERN httpd server, to allow
uploading and managing customer logs by field support engineers over the web and
as a front end to FTP server. Later rewritten in C language, when CGIC library and
Apache httpd were released. Up to 2015 WFM has been a closed source commercial
application used for lightweight document management and supported by a few customers.
It has been open sourced. In 2022 WFM has been rewritten in Go as a stand-alone
application with built-in web server for more modern deployment scenarios.
