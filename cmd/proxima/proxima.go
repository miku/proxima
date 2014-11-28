package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gregjones/httpcache/diskcache"
	"github.com/miku/proxima"
)

// prepend helps to run this service under a directory by appending a base
// TODO: there must be support for this in mux.Handle, or not?
func prepend(s, base string) string {
	if base == "/" {
		return s
	}
	return fmt.Sprintf("%s%s", base, s)
}

func main() {
	addr := flag.String("addr", ":3000", "host:port to listen on")
	cacheDir := flag.String("dir", filepath.Join(os.TempDir(), "proxima.v1"), "cache directory")
	version := flag.Bool("v", false, "prints current program version")
	base := flag.String("base", "/", "URL base")
	verbose := flag.Bool("verbose", false, "log requests")

	flag.Parse()

	if *version {
		fmt.Println(proxima.Version)
		os.Exit(0)
	}

	options := proxima.Options{Cache: diskcache.New(*cacheDir), Verbose: *verbose}
	mux := http.NewServeMux()

	mux.Handle(prepend("/u", *base), proxima.URLHandler(options))
	mux.Handle(prepend("/g/s/", *base), proxima.SameAsHandler(options))
	mux.Handle(prepend("/g/i/", *base), proxima.GndImageHandler(options))

	mux.HandleFunc("/g/r/", func(w http.ResponseWriter, r *http.Request) {
		gnd := r.URL.Path[len("/g/r/"):]
		url := fmt.Sprintf("http://d-nb.info/gnd/%s/about/rdf", gnd)
		http.Redirect(w, r, fmt.Sprintf("/u?url=%s", url), 303)
	})

	log.Printf("Listening on %s, %s\n", *addr, *cacheDir)
	http.ListenAndServe(*addr, mux)
}
