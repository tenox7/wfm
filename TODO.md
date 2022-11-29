# WFM TODO

## Interfaces

* Multiple --prefix instances, this should be possible with map of afero.FS
  indexed by prefix name so it can be looked up inside wfmMain 
* WebDAV server
* FastCGI Interface
* Caddy module
* Web folder with no index, on a separate port?
* Use custom FS implementation to resolve and deny symlinks outside of srv directory
https://github.com/crazcalm/go/commit/8b0b644cd02c59fe2461908304c44d64e8be431e
  maybe afero?
* use url.Parse to get correct url/path

## Security

* seperate prefix for admin/rw users eg /admin with readonly on /
* userless/guest read-only mode, user rw
  requires custom login window
* two factor auth
  requires custom login window
  https://github.com/vkuznet/2fa-server
* garbage collect old f2b entries
* f2b ddos prevention, sleep on too many bans?
* use certmagic for acme? https://github.com/caddyserver/certmagic
* use lego for acme? https://github.com/go-acme/lego
* qps rate limiter
  https://github.com/didip/tollbooth
  https://github.com/uber-go/ratelimit
  https://github.com/sethvargo/go-limiter

## Layout / UI

* top bar too long on mobile/small screen
* custom html login window - needed for two factor auth?
* editable and non editable documents by extension, also for git checkins
* thumbnail / icon view for pictures (cache thumbnails on server?)
* glob filter (*.*) in dir view
* errors in dialog boxes instead of plain text
* html as template

## File IO

* redirects to use new uri paths
* file search function
* path prefix per user, defined in json
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
