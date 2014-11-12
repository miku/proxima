package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	mux.HandleFunc("/gnd/rdf/", func(w http.ResponseWriter, r *http.Request) {
		gnd := r.URL.Path[len("/gnd/rdf/"):]
		url := fmt.Sprintf("http://d-nb.info/gnd/%s/about/rdf", gnd)
		http.Redirect(w, r, fmt.Sprintf("/u?url=%s", url), 303)
	})
	mux.HandleFunc("/gnd/sa/", func(w http.ResponseWriter, r *http.Request) {
		gnd := r.URL.Path[len("/gnd/sa/"):]
		url := fmt.Sprintf("http://%s/u?url=http://d-nb.info/gnd/%s/about/rdf", r.Host, gnd)
		body, resp, err := proxima.Fetch(url)
		if err != nil || resp.StatusCode >= 400 {
			w.WriteHeader(resp.StatusCode)
			w.Write([]byte(resp.Status))
			return
		}

		set := make(map[string]struct{})
		decoder := xml.NewDecoder(strings.NewReader(string(body)))

		for {
			// Read tokens from the XML document in a stream.
			t, _ := decoder.Token()
			if t == nil {
				break
			}
			switch se := t.(type) {
			case xml.StartElement:
				if se.Name.Space == "http://www.w3.org/2002/07/owl#" && se.Name.Local == "sameAs" {
					for _, attr := range se.Attr {
						if attr.Name.Space == "http://www.w3.org/1999/02/22-rdf-syntax-ns#" && attr.Name.Local == "resource" {
							set[attr.Value] = struct{}{}
						}
					}
				}
			default:
			}
		}
		links := make([]string, 0)
		for key := range set {
			links = append(links, key)
		}
		result := make(map[string]interface{})
		result["uri"] = fmt.Sprintf("http://d-nb.info/gnd/%s", gnd)
		result["sa"] = links
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
		w.Write(b)
	})

	log.Printf("Listening on %s, %s\n", *addr, *cacheDir)
	http.ListenAndServe(*addr, mux)
}
