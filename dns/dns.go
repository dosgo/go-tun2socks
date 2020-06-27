package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"github.com/sensepost/godoh/dnsclient"
	"time"
)

var dohC *dnsclient.CloudflareDNS

func StartDns(dnsPort string) error {


	dohC= dnsclient.NewCloudFlareDNS()

	udpServer := &dns.Server{
		Net:          "udp",
		Addr:         ":"+dnsPort,
		Handler:      dns.HandlerFunc(serveDNS),
		UDPSize:      4096,
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
	}
	tcpServer:= &dns.Server{
		Net:          "tcp",
		Addr:         ":"+dnsPort,
		Handler:      dns.HandlerFunc(serveDNS),
		UDPSize:      4096,
		ReadTimeout:  time.Duration(10) * time.Second,
		WriteTimeout: time.Duration(10) * time.Second,
	}
	go udpServer.ListenAndServe();
	tcpServer.ListenAndServe();
	return nil;
}


func isIPv4Query(q dns.Question) bool {
	if q.Qclass == dns.ClassINET && q.Qtype == dns.TypeA {
		return true
	}
	return false
}


func  resolve(r *dns.Msg) (*dns.Msg, error) {
	fmt.Printf("dns ipv6\r\n")
	m :=  &dns.Msg{}
	m.SetReply(r)
	m.Authoritative = false
	domain := r.Question[0].Name
	m.Answer = append(r.Answer, &dns.AAAA{
		Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 60},
		AAAA:   net.ParseIP("fd3e:4f5a:5b81::1"),
	})
	return m, nil
}


func  doIPv4Query(r *dns.Msg) (*dns.Msg, error) {
	m := &dns.Msg{}
	m.SetReply(r)
	m.Authoritative = false
	domain := r.Question[0].Name
	var ip string;
	DnsRes:=dohC.Lookup(domain,dns.TypeA)
	ip=DnsRes.Data;
	//use dot   dot
	m.Answer = append(r.Answer, &dns.A{
		Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
		A:   net.ParseIP(ip),
	})
	// final
	return m, nil
}



func  serveDNS(w dns.ResponseWriter, r *dns.Msg) {
	isIPv4 := isIPv4Query(r.Question[0])
	var msg *dns.Msg
	var err error
	if isIPv4 {
		msg, err = doIPv4Query(r)
	} else {
		msg, err = resolve(r)
	}
	if err != nil {
		dns.HandleFailed(w, r)
	} else {
		w.WriteMsg(msg)
	}
}

