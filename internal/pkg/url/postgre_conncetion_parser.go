package url

import (
	"github.com/lib/pq"
	"strings"
)

// ParsePostgreConn determines if url contains `postgre://` or `postgresql://`
// And converts it to driver form `user=xxx password=xxx host=xxx...`
func ParsePostgreConn(url string) (string, error) {
	if !strings.HasPrefix(url, "postgre://") || !strings.HasPrefix(url, "postgresql://") {
		ret, err := pq.ParseURL(url)
		if err != nil {
			return "", err
		}
		return ret, nil
	}
	return url, nil
}
