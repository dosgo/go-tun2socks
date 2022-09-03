package main

import (
	"github.com/dosgo/go-tun2socks/dns"
	"github.com/dosgo/go-tun2socks/tun2socks"
)

func main() {
	var tunDns = "127.0.0.1:53"
	var socksAddr = "127.0.0.1:1080"

	//start local dns server (doh)
	dns.StartDns(tunDns)

	tun2socks.StartTunDevice("tun0", "10.0.0.2", "255.255.255.0", "10.0.0.1", 1500, socksAddr, tunDns)
}
