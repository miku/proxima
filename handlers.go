package proxima

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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
