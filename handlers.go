package proxima

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

// fetch fetches a single URL and returns the body as []byte
func Fetch(url string) ([]byte, *http.Response, error) {
	defer timeTrack(time.Now(), fmt.Sprintf("fetching %s", url))
	client := new(http.Client)
	resp, err := client.Get(url)
	if err != nil {
		return nil, resp, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}
	return b, resp, nil
}

// UrlHander handles basic caching of URL content
func URLHandler(options Options) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		s := r.FormValue("url")
		if s == "" {
			w.WriteHeader(400)
			w.Write([]byte("url parameter required\n"))
			return
		}
		switch r.Method {
		case "GET":
			val, ok := options.Cache.Get(s)
			if !ok {
				b, resp, err := Fetch(s)
				if err != nil || resp.StatusCode >= 400 {
					w.WriteHeader(resp.StatusCode)
					w.Write([]byte(resp.Status))
					return
				}
				options.Cache.Set(s, b)
			}
			val, ok = options.Cache.Get(s)
			if !ok {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
				return
			}
			w.Write(val)
		case "DELETE":
			options.Cache.Get(s)
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte(http.StatusText(http.StatusNoContent)))
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
			return
		}
	}
	return http.HandlerFunc(fn)
}

// SameAsHandler returns the reference to VIAF and DBP for a given GND
func SameAsHandler(options Options) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		gnd := r.URL.Path[len("/gnd/sa/"):]
		url := fmt.Sprintf("http://%s/u?url=http://d-nb.info/gnd/%s/about/rdf", r.Host, gnd)
		body, resp, err := Fetch(url)
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
	}
	return http.HandlerFunc(fn)

}
