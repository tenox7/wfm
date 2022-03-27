package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
)

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
		http.Error(w, "Too many bad username/password attempts", http.StatusTooManyRequests)
		return "", false
	}

	u, p, ok := r.BasicAuth()
	if !ok || u == "" {
		log.Printf("no auth header for %v (u=%v)", ip, u)
		goto unauth
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

	log.Printf("auth: found no matching usr/pwd ip=%v u=%v)", ip, u)
	f2b.ban(ip)
	// ideally we should return here but firefox keeps feeding wrong creds
	// setting authenticate header instead to force new user/pass window
	// empty username will not ban the client so it will work at second try

unauth:
	w.Header().Set("WWW-Authenticate", "Basic realm=\"wfm\"")
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
	return "", false
}

// this doesn't really work
func logout(w http.ResponseWriter) {
	http.Error(w, "Logged out", http.StatusUnauthorized)
}
