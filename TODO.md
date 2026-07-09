# WFM TODO

## Interfaces

* WebDAV server
* Caddy module
* Use custom FS implementation to resolve and deny symlinks outside of srv directory
https://github.com/crazcalm/go/commit/8b0b644cd02c59fe2461908304c44d64e8be431e
  maybe afero?

## Security

* two factor auth
  requires custom login window
  https://github.com/vkuznet/2fa-server
* garbage collect old f2b entries
* f2b ddos prevention, sleep on too many bans?
* qps rate limiter
  https://github.com/didip/tollbooth
  https://github.com/uber-go/ratelimit
  https://github.com/sethvargo/go-limiter

## Layout / UI

* custom html login window
* thumbnail / icon view for pictures (cache thumbnails on server?)

## File IO

* file search function
* editable and non editable documents by extension, also for git checkins
* zip/unzip archives
* udf iso format https://github.com/mogaika/udf
* iso files recursive list
* zipped iso like .iso.gz, .iso.xz, .iso.lz
* du with xdev as a go routine
* git client https://github.com/go-git/go-git
* file locking https://github.com/gofrs/flock
* support for different filesystems, S3, SMB, archive files as io/fs
  https://github.com/spf13/afero ?
* archive files in main view / graphical/table form
