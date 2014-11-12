package proxima

import "github.com/gregjones/httpcache"

const Version = "0.1.0"

// Options are application options passed to all handlers
type Options struct {
	Cache httpcache.Cache
}
