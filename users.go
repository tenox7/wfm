package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

var (
	// you can also hardcode users here instead of loading password file
	users = []struct {
		User, Salt, Hash string
		RW               bool
	}{}
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

func listUsers() {
	for _, u := range users {
		fmt.Printf("User: %q, RW: %v\n", u.User, u.RW)
	}
}

func manageUsers() {
	switch flag.Arg(1) {
	case "list":
		listUsers()
	default:
		fmt.Println("usage: user <list|add|delete|passwd|rw|ro> [username]")
	}
}
