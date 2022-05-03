package mode

import (
	"fmt"
	"sync"

	"github.com/jackwakefield/gopac"
	log "github.com/sirupsen/logrus"
)

var parser = new(gopac.Parser)
var once = &sync.Once{}

// https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Proxy_servers_and_tunneling/Proxy_Auto-Configuration_PAC_file
func Pac(host string) (string, error) {
	once.Do(func() {
		// use parser.Parse(path) to parse a local file
		// or parser.ParseUrl(url) to parse a remote file
		if err := parser.ParseUrl("http://immun.es/pac"); err != nil {
			log.Fatalf("Failed to parse PAC (%s)", err)
		}
	})

	entry, err := parser.FindProxy("", host)

	if err != nil {
		return "", fmt.Errorf("failed to find proxy entry (%s)", err)
	}

	return entry, nil
}

func init() {
}
