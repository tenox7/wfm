# WFM TODO

* acme needs acccess to cacerts outside of chroot
  maybe copy cacerts to acme cfg dir set SSL_CERT_DIR
  or maybe https://github.com/gwatts/rootcerts 
* CSS html/body height 100%
* top bar buttons have different bgcolors
* checkboxes, multi file routines
* zip/unzip archives
* iso files recursive list
* du with xdev as a go routine
* two factor auth
* custom html login window
  needed for two factor auth
  fail2ban improvements
* better handle cert chdir issue
  get and preload cert manually on start?
  hide acm cache dir?
  try https://github.com/go-acme/lego
* git client https://github.com/go-git/go-git
* file locking https://github.com/gofrs/flock
* docker support (no chroot) - mount dir as / ?
* webdav server
* html as template
* archive files in main view / graphical/table form
* support for different filesystems, S3, SMB, archive files as io/fs
* separate icons for different file types like images
* editable and non editable documents by extension, also for git checkins
* thumbnail / icon view for pictures, cache thumbnails on server
* glob filter (*.*) in dir view
* FastCGI Interface
* Chroot and User in Systemd Unit
* Security Hardening in Systemd Unit
* Highlight newly uploaded file/created dir/bookmark
* different icons for different file types
* garbage collect old f2b entries
* f2b dos prevention
* remove f2b dump