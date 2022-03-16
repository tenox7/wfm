package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

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
	if !*dof2b {
		return false
	}
	db.Lock()
	defer db.Unlock()

	// TODO: purge old entries

	l, ok := db.entr[ip]
	if !ok {
		//log.Printf("auth: ip=%v not in DB", ip)
		return false
	}
	//log.Printf("auth: found ip=%v for=%v no#tries=%v",
	//ip, time.Until(l.banUntil), l.noTries)
	return time.Now().Before(l.banUntil)
}

func (db *f2bDB) ban(ip string) {
	if !*dof2b {
		return
	}
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
	if !*dof2b {
		return
	}
	db.Lock()
	defer db.Unlock()
	delete(db.entr, ip)
}

func (db *f2bDB) dump(w http.ResponseWriter) {
	db.Lock()
	defer db.Unlock()

	for i, l := range db.entr {
		fmt.Fprintf(w, "ip=%v for=%v tries=%v\n", i, time.Until(l.banUntil), l.noTries)
	}
}

func dumpf2b(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprintf(w, "F2B DB\n\n")
	f2b.dump(w)
}
