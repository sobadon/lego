// original: https://github.com/masa23/lego/blob/3b9bdfb360ae0307e8997a738a28a924d7e1ec1f/providers/dns/selfdns/dnsserver.go

package selfdns

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type record struct {
	hostname string
	fqdn     string
	values   []string
}

var rec []record

func findRecord(fqdn string) (record, bool) {
	fqdn = strings.ToLower(fqdn)
	for _, r := range rec {
		if r.fqdn == fqdn {
			return r, true
		}
	}
	return record{}, false
}

func request(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	if r.Opcode == dns.OpcodeQuery {
		for _, q := range r.Question {
			rec, ok := findRecord(q.Name)
			if !ok {
				continue
			}
			switch q.Qtype {
			case dns.TypeSOA:
				rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN SOA %s. admin.%s. %d %d %d %d %d", q.Name, rec.hostname, rec.hostname, time.Now().Unix(), 10, 10, 10, 10))
				if err != nil {
					log.Fatalf("Failed to create RR: %v", err)
				}
				m.Answer = append(m.Answer, rr)
			case dns.TypeNS:
				rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN NS %s.", q.Name, rec.hostname))
				if err != nil {
					log.Fatalf("Failed to create RR: %v", err)
				}
				m.Answer = append(m.Answer, rr)
			case dns.TypeTXT:
				for _, value := range rec.values {
					rr, err := dns.NewRR(fmt.Sprintf("%s 10 IN TXT %s", q.Name, value))
					if err != nil {
						log.Fatalf("Failed to create RR: %v", err)
					}
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
	w.WriteMsg(m)
}

func (d *DNSProvider) Run(fqdn, value string) error {
	// DNSサーバの起動
	// recが空の場合は初回起動
	if len(rec) == 0 {
		d.config.serverUDP = dns.Server{Addr: d.config.ListenAddress + ":53", Net: "udp"}
		d.config.serverTCP = dns.Server{Addr: d.config.ListenAddress + ":53", Net: "tcp"}
	}

	// レコードの設定
	alreadyExistsRecord := false
	for i, r := range rec {
		if r.fqdn == fqdn {
			alreadyExistsRecord = true
			rec[i].values = append(rec[i].values, value)
			break
		}
	}
	if !alreadyExistsRecord {
		rec = append(rec, record{
			hostname: d.config.ServerHostname,
			fqdn:     fqdn,
			values:   []string{value},
		})
	}

	dns.HandleFunc(".", request)
	go func() {
		d.config.serverUDP.ListenAndServe()
	}()

	go func() {
		d.config.serverTCP.ListenAndServe()
	}()

	return nil
}

func (d *DNSProvider) Stop(domain, keyAuth string) error {
	// fqdn, value := dns01.GetRecord(domain, keyAuth)
	// fmt.Printf("Stop() 1: domain=%s, keyAuth=%s, fqdn=%s, value=%s, rec=%+v\n", domain, keyAuth, fqdn, value, rec)

	// // // recからレコードを削除
	// newRec := rec[:0]
	// // for _, r := range rec {
	// // 	if r.fqdn != fqdn && r.value != value {
	// // 		// 完了していないものは残す
	// // 		newRec = append(newRec, r)
	// // 	}
	// // }
	// rec = newRec

	// fmt.Printf("Stop() 2: domain=%s, keyAuth=%s, rec=%+v newRec=%+v\n", domain, keyAuth, rec, newRec)

	// いろいろ変えた結果、rec が空になることはないので、あんまりよくない

	// recが空の場合はサーバを停止
	if len(rec) == 0 {
		err := d.config.serverUDP.Shutdown()
		if err != nil {
			return err
		}
		return d.config.serverTCP.Shutdown()
	}
	return nil
}
