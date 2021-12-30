// Web File Manager
//
// TODO:
// * dirlist with sorting
// * file routines
// * authentication
// * setuid/setgid
// * https/certbot
// * git client

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"syscall"
)

var (
	addr = flag.String("addr", ":8080", "Listen address and port")
	base = flag.String("base_dir", "", "Base directory path")
)

func listFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	d, err := ioutil.ReadDir("/")
	if err != nil {
		fmt.Fprintf(w, "Error: %v\n", err)
		log.Printf("Error: %v", err)
		return
	}

	for _, f := range d {
		fmt.Fprintf(w, "* %v\n", f.Name())
	}
}

func main() {
	flag.Parse()
	var err error
	if *base != "" {
		err = syscall.Chroot(*base)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("Starting WFM on %v for directory %v", *addr, *base)

	http.HandleFunc("/", listFiles)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal(err)
	}
}
