package main

import (
	"github.com/dosgo/go-tun2socks/core"
	"github.com/dosgo/go-tun2socks/dns"
)

func main() {
	var tunDns = "127.0.0.1:53"
	var socksAddr = "127.0.0.1:1080"

	//start local dns server (doh)
	dns.StartDns(tunDns)

	core.StartTunDevice("tun0", "10.0.0.2", "255.255.255.0", "10.0.0.1", 1500, socksAddr, tunDns)
}
