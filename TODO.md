# WFM TODO

## General
* checkboxes, multi file routines
* webdav server
* FastCGI Interface

## Security
* two factor auth
* docker support
  no chroot - mount dir as / ?
  env vars for port, etc?
* garbage collect old f2b entries
* f2b ddos prevention, sleep on too many bans?
* remove f2b dump
* Chroot and User in Systemd Unit
* Security Hardening in Systemd Unit
* RW/RO per user and no user read-only mode

## ACME / Auto Cert Manager
* acme dir with key/cert is exposed inside chroot dir
  obtain acme cert before chroot?? self call https?
  get and preload cert manually on start?
  hide acm cache dir?
* try https://github.com/go-acme/lego


## Layout / UI
* top bar too long on mobile/small screen
* Highlight newly uploaded file/created dir/bookmark
* different icons for different file types
* custom html login window - needed for two factor auth?
* editable and non editable documents by extension, also for git checkins
* thumbnail / icon view for pictures, cache thumbnails on server
* glob filter (*.*) in dir view
* separate icons for different file types like images, docs, etc
* errors in dialog boxes instead of plain text
* html as template

## File IO
* exclude/hide folders based on list
* zip/unzip archives
* iso files recursive list
* zipped iso like .iso.gz, .iso.xz, .iso.lz
* auto unpack via mime type...
* udf iso format https://github.com/mogaika/udf
* add more formats like tgz/txz, etc
* du with xdev as a go routine
* git client https://github.com/go-git/go-git
* file locking https://github.com/gofrs/flock
* support for different filesystems, S3, SMB, archive files as io/fs
* archive files in main view / graphical/table form
