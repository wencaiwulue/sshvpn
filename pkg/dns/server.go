package dns

import (
	"context"
	"time"

	miekgdns "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"k8s.io/apimachinery/pkg/util/cache"
)

var (
	maxConcurrent int64 = 1024
	logInterval         = 2 * time.Second
)

type server struct {
	dnsCache      *cache.LRUExpireCache
	needForward   func(domain string) bool
	localServer   string
	localClient   *miekgdns.Client
	forwardServer string
	forwardClient *miekgdns.Client
	addRoute      func(msg *miekgdns.Msg)
	fwdSem        *semaphore.Weighted // Limit the number of concurrent external DNS requests in-flight
	logInterval   rate.Sometimes      // Rate-limit logging about hitting the fwdSem limit
}

func NewDNSServer(network, address string, needForward func(domain string) bool, addRoute func(msg *miekgdns.Msg), forwardServer, localServer string) error {
	return miekgdns.ListenAndServe(address, network, &server{
		dnsCache:      cache.NewLRUExpireCache(1000),
		needForward:   needForward,
		localServer:   localServer,
		localClient:   &miekgdns.Client{Net: "udp", SingleInflight: true, Timeout: time.Second * 30},
		forwardServer: forwardServer,
		forwardClient: &miekgdns.Client{Net: "udp", SingleInflight: true, Timeout: time.Second * 30},
		addRoute:      addRoute,
		fwdSem:        semaphore.NewWeighted(maxConcurrent),
		logInterval:   rate.Sometimes{Interval: logInterval},
	})
}

func (s *server) ServeDNS(w miekgdns.ResponseWriter, msg *miekgdns.Msg) {
	defer w.Close()
	if len(msg.Question) == 0 {
		return
	}
	var answer *miekgdns.Msg
	var err error
	if s.needForward(msg.Question[0].Name) {
		answer, _, err = s.forwardClient.ExchangeContext(context.Background(), msg, s.forwardServer)
		s.addRoute(answer)
	} else {
		answer, _, err = s.localClient.ExchangeContext(context.Background(), msg, s.localServer)
	}
	if err != nil {
		log.Warnf("lookup dns failed: %v", err)
	}
	if answer != nil {
		_ = w.WriteMsg(answer)
	} else {
		msg.Response = true
		_ = w.WriteMsg(msg)
	}
}
