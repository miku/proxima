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

func main() {
	addr := flag.String("addr", ":3000", "host:port to listen on")
	cacheDir := flag.String("dir", filepath.Join(os.TempDir(), "proxima.v1"), "cache directory")
	version := flag.Bool("v", false, "prints current program version")

	flag.Parse()

	if *version {
		fmt.Println(proxima.Version)
		os.Exit(0)
	}

	options := proxima.Options{Cache: diskcache.New(*cacheDir)}
	mux := http.NewServeMux()

	mux.Handle("/u", proxima.URLHandler(options))
	mux.Handle("/gnd/sa/", proxima.SameAsHandler(options))

	mux.HandleFunc("/gnd/rdf/", func(w http.ResponseWriter, r *http.Request) {
		gnd := r.URL.Path[len("/gnd/rdf/"):]
		url := fmt.Sprintf("http://d-nb.info/gnd/%s/about/rdf", gnd)
		http.Redirect(w, r, fmt.Sprintf("/u?url=%s", url), 303)
	})

	log.Printf("Listening on %s, %s\n", *addr, *cacheDir)
	http.ListenAndServe(*addr, mux)
}
