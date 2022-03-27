package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type userDB struct {
	User, Salt, Hash string
	RW               bool
}

var (
	// you can also hardcode users here instead of loading password file
	users = []userDB{}
)

func loadUsers() {
	pwd, err := ioutil.ReadFile(*passwdDb)
	if err != nil {
		log.Fatal("unable to read password file: ", err)
	}
	err = json.Unmarshal(pwd, &users)
	if err != nil {
		log.Fatal("unable to parse password file: ", err)
	}
	log.Printf("Loaded %q (%d users)", *passwdDb, len(users))
}

func saveUsers() {
	u, err := json.Marshal([]userDB{})
	if err != nil {
		log.Fatal(err)
	}
	// TODO: pretty format file
	err = ioutil.WriteFile(*passwdDb, u, 600)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Saved %q (%v users)", *passwdDb, len(users))
}

func manageUsers() {
	switch flag.Arg(1) {
	case "list":
		listUsers()
	case "newfile":
		saveUsers()
	case "add":
		addUser(flag.Arg(2), rwStrBool(flag.Arg(3)))
	case "delete":
		delUser(flag.Arg(2))
	case "passwd":
		pwdUser(flag.Arg(2))
	case "access":
		setUser(flag.Arg(2), rwStrBool(flag.Arg(3)))
	default:
		fmt.Println("usage: user <list|add|delete|passwd|access|newfile> [username] [rw|ro]")
	}
}

func listUsers() {
	loadUsers()
	for _, u := range users {
		fmt.Printf("User: %q, RW: %v\n", u.User, u.RW)
	}
}

func addUser(usr string, rw bool) {
	if usr == "" {
		log.Fatal("user add requires username and ro/rw\n")
	}
	loadUsers()
	fmt.Print("Password: ")
	var pwd string
	fmt.Scanln(&pwd)
	salt := rndStr(8)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(salt+pwd)))
	users = append(users, userDB{User: usr, Salt: salt, Hash: hash, RW: rw})
	saveUsers()
}

func delUser(usr string) {
	loadUsers()
	var udb []userDB
	for _, u := range users {
		if u.User == usr {
			continue
		}
		udb = append(udb, u)
	}
	if len(users) == len(udb) {
		log.Fatal("User not found / nothing changed")
	}
	users = udb
	saveUsers()
}

func pwdUser(usr string) {
	if usr == "" {
		log.Fatal("user passwd requires username\n")
	}
	loadUsers()
	fmt.Print("Password: ")
	var pwd string
	fmt.Scanln(&pwd)
	salt := rndStr(8)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(salt+pwd)))
	chg := false
	for i, u := range users {
		if u.User != usr {
			continue
		}
		users[i].Salt = salt
		users[i].Hash = hash
		chg = true
	}
	if !chg {
		log.Fatal("User not found / nothing changed")
	}
	saveUsers()
}

func setUser(usr string, rw bool) {
	if usr == "" {
		log.Fatal("user add requires username and ro/rw\n")
	}
	loadUsers()
	chg := false
	for i, u := range users {
		if u.User != usr {
			continue
		}
		users[i].RW = rw
		chg = true
	}
	if !chg {
		log.Fatal("User not found / nothing changed")
	}
	saveUsers()
}

func rwStrBool(acc string) bool {
	var rw bool
	switch acc {
	case "rw":
		rw = true
	case "ro":
		rw = false
	default:
		log.Fatal("access must be either 'ro' or 'rw'")
	}
	return rw
}

func rndStr(len int) string {
	b := make([]byte, len)
	rand.Seed(time.Now().Unix())
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)[:len]
}
