package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/axeljohnsson/indeed"
)

var addr = flag.String("addr", ":8080", "HTTP network address")

func main() {
	if err := mainErr(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mainErr() error {
	flag.Parse()

	rdap := indeed.NewRDAPClient(indeed.RDAPBaseURL)
	whois := indeed.NewWHOISClient()
	feedHandler := indeed.LogHandler(indeed.NewFeedHandler(rdap, whois), slog.Default())
	http.Handle("/feed", feedHandler)

	return http.ListenAndServe(*addr, nil)
}
