# WFM TODO

## Interfaces
* WebDAV server
* FastCGI Interface

## Security
* User manager add/delete/chpw user via arg
* userless/guest read-only mode, user rw
  requires custom login window
* two factor auth
  requires custom login window
* docker support
  no chroot - mount dir as / ?
  env vars for port, etc?
* systemd support
  chroot
  user
* garbage collect old f2b entries
* f2b ddos prevention, sleep on too many bans?

## ACME / Auto Cert Manager
* acme dir with key/cert is exposed inside chroot dir
  get and preload cert manually on start before chroot?
  hide acm cache dir?
* try https://github.com/go-acme/lego

## Layout / UI
* add flag to specify own favicon.ico
* top bar too long on mobile/small screen
* custom html login window - needed for two factor auth?
* editable and non editable documents by extension, also for git checkins
* thumbnail / icon view for pictures, cache thumbnails on server
* glob filter (*.*) in dir view
* separate icons for different file types like images, docs, etc
* errors in dialog boxes instead of plain text
* html as template

## File IO
* do not log FormValue["text"] as it contains text data from edit
* exclude/hide folders based on list
* udf iso format https://github.com/mogaika/udf
* zip/unzip archives
* iso files recursive list
* zipped iso like .iso.gz, .iso.xz, .iso.lz
* auto unpack via mime type...
* add more formats like tgz/txz, etc
* du with xdev as a go routine
* git client https://github.com/go-git/go-git
* file locking https://github.com/gofrs/flock
* support for different filesystems, S3, SMB, archive files as io/fs
* archive files in main view / graphical/table form
