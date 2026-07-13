package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
)

// anonUser is the reserved identity served for unauthenticated requests
// when -anon_ro is enabled; it can never match a real password db entry.
const anonUser = "anonymous"

func auth(w http.ResponseWriter, r *http.Request) (string, bool) {
	if len(users) == 0 {
		return "n/a", *noPwdDbRW
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
		return "", false
	}

	if f2b.check(ip) {
		log.Printf("auth: %v is banned", ip)
		htErrStatus(w, r, http.StatusTooManyRequests, "Too many attempts",
			"your address is temporarily banned after repeated bad username/password attempts")
		return "", false
	}

	// the lock icon links to ?fn=login which must always challenge so the
	// browser asks for credentials; other requests may fall back to anonymous
	login := r.URL.Query().Get("fn") == "login"

	u, p, ok := r.BasicAuth()
	if !ok || u == "" {
		if *anonRO && !login {
			return anonUser, false
		}
		log.Printf("no auth header for %v (u=%v)", ip, u)
		return challenge(w, r)
	}

	for _, usr := range users {
		if subtle.ConstantTimeCompare([]byte(u), []byte(usr.User)) != 1 {
			continue
		}

		s := fmt.Sprintf("%x", sha256.Sum256([]byte(usr.Salt+p)))
		if subtle.ConstantTimeCompare([]byte(s), []byte(usr.Hash)) == 1 {
			go f2b.unban(ip)
			return usr.User, usr.RW
		}
	}

	if *anonRO && !login {
		// browsers resend cached credentials forever; banning or challenging
		// stale ones here would lock the user out of anonymous browsing
		log.Printf("auth: bad creds ip=%v u=%v, serving anonymous ro", ip, u)
		return anonUser, false
	}

	log.Printf("auth: found no matching usr/pwd ip=%v u=%v)", ip, u)
	f2b.ban(ip)
	// ideally we should return here but firefox keeps feeding wrong creds
	// setting authenticate header instead to force new user/pass window
	// empty username will not ban the client so it will work at second try

	return challenge(w, r)
}

// challenge makes the browser show its native login box via 401 +
// WWW-Authenticate; the styled body is only rendered if the user cancels it.
func challenge(w http.ResponseWriter, r *http.Request) (string, bool) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"wfm\"")
	htErrStatus(w, r, http.StatusUnauthorized, "Unauthorized", "a valid username and password are required")
	return "", false
}

// this doesn't really work: Basic Auth has no real logout, 401 just nudges the
// browser to drop cached credentials.
func (r *wfmRequest) logout() {
	htErrStatus(r.w, r.req, http.StatusUnauthorized, "Logged out", "close the browser to fully clear stored credentials")
}
