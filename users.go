package main

import (
	"encoding/json"
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

func loadUsers(pwdb string) {
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
