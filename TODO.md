# WFM TODO

## General
* checkboxes, multi file routines
* webdav server
* FastCGI Interface

## Security
* two factor auth
* docker support (no chroot) - mount dir as / ?
* garbage collect old f2b entries
* f2b ddos prevention, sleep on too many bans?
* remove f2b dump
* Chroot and User in Systemd Unit
* Security Hardening in Systemd Unit

## ACME / Auto Cert Manager
* acme needs acccess to cacerts outside of chroot
  maybe copy cacerts to acme cfg dir set SSL_CERT_DIR
  or maybe https://github.com/gwatts/rootcerts 
* acme dir with key/cert is exposed inside chroot dir
  obtain acme cert before chroot?? self call https?
  get and preload cert manually on start?
  hide acm cache dir?
  try https://github.com/go-acme/lego

## Layout / UI
* CSS html/body height 100% especially for dialogs, textarea
* top bar too long on mobile/small screen
* html as template
* Highlight newly uploaded file/created dir/bookmark
* different icons for different file types
* custom html login window - needed for two factor auth?
* editable and non editable documents by extension, also for git checkins
* thumbnail / icon view for pictures, cache thumbnails on server
* glob filter (*.*) in dir view
* separate icons for different file types like images
* errors in dialog boxes

## File IO
* exclude/deny folders based on list
* zip/unzip archives
* iso files recursive list
* du with xdev as a go routine
* git client https://github.com/go-git/go-git
* file locking https://github.com/gofrs/flock
* support for different filesystems, S3, SMB, archive files as io/fs
* archive files in main view / graphical/table form
* combine doc_dir:doc_pfx in to one string flag
