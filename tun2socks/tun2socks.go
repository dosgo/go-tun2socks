package tun2socks

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/dosgo/go-tun2socks/core"
	"github.com/dosgo/go-tun2socks/socks"
	"github.com/dosgo/go-tun2socks/tun"
	"gvisor.dev/gvisor/pkg/bufferv2"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

var wrapnet uint32
var mask uint32
var relayip net.IP
var port uint16
var sock5Addr string
var tunDNS string

func StartTunDevice(tunDevice string, tunAddr string, tunMask string, tunGW string, mtu int, _sock5Addr string, _tunDNS string) {
	dev, err := tun.RegTunDev(tunDevice, tunAddr, tunMask, tunGW, tunDNS)
	if err != nil {
		fmt.Println("start tun err:", err)
		return
	}
	sock5Addr = _sock5Addr
	tunDNS = _tunDNS
	ForwardTransportFromIo(dev, mtu, rawTcpForwarder, rawUdpForwarder)
}

func rawTcpForwarder(conn core.CommTCPConn) error {
	defer conn.Close()
	socksConn, err1 := net.Dial("tcp", sock5Addr)
	if err1 != nil {
		log.Println(err1)
		return nil
	}
	defer socksConn.Close()
	if socks.SocksCmd(socksConn, 1, conn.LocalAddr().String()) == nil {
		go io.Copy(conn, socksConn)
		io.Copy(socksConn, conn)
	}
	return nil
}

func rawUdpForwarder(conn core.CommUDPConn, ep core.CommEndpoint) error {
	defer conn.Close()
	//dns port
	if strings.HasSuffix(conn.LocalAddr().String(), ":53") {
		dnsReq(conn, "udp", tunDNS)
	}
	return nil
}

func ForwardTransportFromIo(dev io.ReadWriteCloser, mtu int, tcpCallback core.ForwarderCall, udpCallback core.UdpForwarderCall) error {
	_, channelLinkID, err := core.NewDefaultStack(mtu, tcpCallback, udpCallback)
	if err != nil {
		log.Printf("err:%v", err)
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// write tun
	go func(_ctx context.Context) {
		for {
			info := channelLinkID.ReadContext(_ctx)
			if info == nil {
				log.Printf("channelLinkID exit \r\n")
				break
			}
			info.ToView().WriteTo(dev)
			info.DecRef()
		}
	}(ctx)

	// read tun data
	var buf = make([]byte, mtu+80)
	for {
		n, e := dev.Read(buf[:])
		if e != nil {
			log.Printf("err:%v", err)
			break
		}

		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: bufferv2.MakeWithData(buf[:n]),
		})

		switch header.IPVersion(buf) {
		case header.IPv4Version:
			channelLinkID.InjectInbound(header.IPv4ProtocolNumber, pkt)
		case header.IPv6Version:
			channelLinkID.InjectInbound(header.IPv6ProtocolNumber, pkt)
		}
		pkt.DecRef()
	}
	return nil
}

/*to dns*/
func dnsReq(conn core.CommUDPConn, action string, dnsAddr string) error {
	if action == "tcp" {
		dnsConn, err := net.Dial(action, dnsAddr)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		defer dnsConn.Close()
		go io.Copy(conn, dnsConn)
		io.Copy(dnsConn, conn)
		fmt.Printf("dnsReq Tcp\r\n")
		return nil
	} else {
		buf := make([]byte, 4096)
		var n = 0
		var err error
		n, err = conn.Read(buf)
		if err != nil {
			fmt.Printf("c.Read() = %v", err)
			return err
		}
		dnsConn, err := net.Dial("udp", dnsAddr)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		defer dnsConn.Close()
		_, err = dnsConn.Write(buf[:n])
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		n, err = dnsConn.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			fmt.Println(err.Error())
			return err
		}
	}
	return nil
}
