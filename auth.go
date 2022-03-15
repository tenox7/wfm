package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	// you can also hardcode users here instead of loading password file
	users = []struct{ User, Salt, Hash string }{}
	f2b   = newf2b()
)

func loadPwdDb(pwdb string) {
	pwd, err := ioutil.ReadFile(pwdb)
	if err != nil {
		log.Fatal("unable to read password file: ", err)
	}
	err = json.Unmarshal(pwd, &users)
	if err != nil {
		log.Fatal("unable to parse password file: ", err)
	}
	log.Printf("Loaded %q (%d users)", pwdb, len(users))
}

func auth(w http.ResponseWriter, r *http.Request) string {
	if len(users) == 0 {
		return "n/a"
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Print(err)
		return ""
	}

	if f2b.check(ip) {
		log.Printf("auth: %v is banned", ip)
		http.Error(w, "Too many bad username/password attempts", http.StatusTooManyRequests)
		return ""
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
			f2b.unban(ip)
			return u
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
	return ""
}

// this doesn't really work
func logout(w http.ResponseWriter) {
	http.Error(w, "Logged out", http.StatusUnauthorized)
}

type f2bDBentr struct {
	banUntil time.Time
	noTries  int
}

type f2bDB struct {
	entr map[string]f2bDBentr
	sync.Mutex
}

func newf2b() *f2bDB {
	l := new(f2bDB)
	l.entr = make(map[string]f2bDBentr)
	return l
}

func (db *f2bDB) check(ip string) bool {
	db.Lock()
	defer db.Unlock()
	// purge old entries, for bigger systems this should be in a separate goroutine
	/*for k, v := range li.db {
		if v.banUntil.After(time.Now()) {
			delete(li.db, k)
		}
	}*/
	// check client
	l, ok := db.entr[ip]
	if !ok {
		log.Printf("auth: %v not in DB", ip)
		return false
	}
	log.Printf("auth: found IP=%v For=%v No#Tries=%v",
		ip, time.Until(l.banUntil), l.noTries)
	return time.Now().Before(l.banUntil)
}

func (db *f2bDB) ban(ip string) {
	db.Lock()
	defer db.Unlock()
	l, ok := db.entr[ip]
	if !ok {
		l = f2bDBentr{noTries: 0}
	}
	l.banUntil = time.Now().Add(time.Minute * time.Duration(l.noTries))
	l.noTries++
	db.entr[ip] = l

	log.Printf("auth: Banning ip=%v for=%v no#tries=%v", ip, time.Until(l.banUntil), l.noTries)
}

func (db *f2bDB) unban(ip string) {
	db.Lock()
	defer db.Unlock()
	delete(db.entr, ip)

	log.Printf("auth: Unbanning ip=%v", ip)
}

func (db *f2bDB) dump(w http.ResponseWriter) {
	db.Lock()
	defer db.Unlock()

	for i, l := range db.entr {
		fmt.Fprintf(w, "ip=%v for=%v until=%v tries=%v\n", i, time.Until(l.banUntil), l.banUntil, l.noTries)
	}
}

func dumpf2b(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprintf(w, "Limiter DB\n\n")
	f2b.dump(w)
}
