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
	u, err := json.Marshal(users)
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
	case "add":
		addUser(flag.Arg(2), flag.Arg(3))
	case "delete":
		delUser(flag.Arg(2))
	default:
		fmt.Println("usage: user <list|add|delete|passwd|setrw|setro> [username] [rw|ro]")
	}
}

func listUsers() {
	for _, u := range users {
		fmt.Printf("User: %q, RW: %v\n", u.User, u.RW)
	}
}

func addUser(usr, rw string) {
	if usr == "" || rw == "" {
		log.Fatal("user add requires username and ro/rw\n")
	}
	var bRW bool
	if rw == "rw" {
		bRW = true
	}

	fmt.Print("Password: ")
	var pwd string
	fmt.Scanln(&pwd)
	salt := rndStr(8)
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(salt+pwd)))
	users = append(users, userDB{User: usr, Salt: salt, Hash: hash, RW: bRW})
	saveUsers()
}

func delUser(usr string) {
	var udb []userDB
	for _, u := range users {
		if u.User == usr {
			continue
		}
		udb = append(udb, u)
	}
	users = udb
	saveUsers()
}

func rndStr(len int) string {
	b := make([]byte, len)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)[:len]
}
