package proxima

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gregjones/httpcache"
)

const Version = "0.1.3"

// Options are application options passed to all handlers
type Options struct {
	Cache   httpcache.Cache
	Verbose bool
}

// timeTrack logs execution times
func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s\n", name, elapsed)
}

// Fetch fetches a single URL and returns the body as []byte
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
