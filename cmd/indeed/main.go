package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/axeljohnsson/indeed"
)

const rdapBaseURL = "https://rdap.org/"

var addr = flag.String("addr", ":8080", "HTTP network address")

func main() {
	if err := mainErr(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Parse()
	handler := indeed.NewHandler(indeed.NewRDAPClient(rdapBaseURL))
	return http.ListenAndServe(*addr, handler)
}
