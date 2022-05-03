package util

import (
	"fmt"
	"net"

	miekgdns "github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
)

func NewDNSServer(addr string) error {
	svr := &miekgdns.Server{
		Addr:      addr,
		Net:       "tcp-tls",
		TLSConfig: ssl.TlsConfigServer,
		Handler:   &server{},
	}

	return svr.ListenAndServe()
}

type server struct {
}

func (s *server) ServeDNS(w miekgdns.ResponseWriter, r *miekgdns.Msg) {
	ip, err := net.LookupIP(r.Question[0].Name)

	if err != nil {
		err = w.WriteMsg(r)
		if err != nil {
			log.Warnln(err)
		}
	} else {
		m := new(miekgdns.Msg)
		m.SetReply(r)
		m.Compress = r.Compress

		switch r.Question[0].Qtype {
		case miekgdns.TypeA:
			rr, err := miekgdns.NewRR(fmt.Sprintf("%s A %s", r.Question[0].Name, ip[0]))
			if err == nil {
				m.Answer = append(m.Answer, rr)
			}
		}

		err = w.WriteMsg(r)
		if err != nil {
			log.Warnln(err)
		}
	}
}
