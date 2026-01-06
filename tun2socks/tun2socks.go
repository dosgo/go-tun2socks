package tun2socks

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/dosgo/go-tun2socks/core"
	"github.com/dosgo/go-tun2socks/socks"
	"github.com/dosgo/go-tun2socks/tun"
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
	// DNS 请求保持不变
	if strings.HasSuffix(conn.LocalAddr().String(), ":53") {
		return dnsReq(conn, "udp", tunDNS)
	}

	// 建立到 Socks5 的控制连接 (TCP)
	socksConn, err := net.Dial("tcp", sock5Addr)
	if err != nil {
		return nil
	}
	// 注意：只要这个 TCP 连接不断开，UDP 映射就一直存在
	defer socksConn.Close()

	// 1. 发起 UDP 关联
	relayAddr, err := socks.SocksUdpCmd(socksConn, "0.0.0.0:0")
	if err != nil {
		return nil
	}

	// 2. 连接到代理分配的 UDP 中继端口
	udpRelay, err := net.Dial("udp", relayAddr.String())
	if err != nil {
		return nil
	}
	defer udpRelay.Close()

	targetAddr := conn.LocalAddr().String()

	// 3. 数据转发：中继 -> 网卡
	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := udpRelay.Read(buf)
			if err != nil {
				return
			}
			data, _ := socks.DecodeUDPPacket(buf[:n])
			conn.Write(data)
		}
	}()

	// 4. 数据转发：网卡 -> 中继
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		encodedPkt := socks.EncodeUDPPacket(targetAddr, buf[:n])
		udpRelay.Write(encodedPkt)
	}
	return nil
}

func ForwardTransportFromIo(dev io.ReadWriteCloser, mtu int, tcpCallback core.ForwarderCall, udpCallback core.UdpForwarderCall) error {
	//macAddr, _ := net.ParseMAC("de:ad:be:ee:ee:ef")
	//var channelLinkID = channel.New(1024, uint32(mtu), tcpip.LinkAddress(macAddr))

	ep := core.NewDirectEndpoint(
		dev,
		uint32(mtu),
	)

	_, err := core.NewDefaultStack(ep, mtu, tcpCallback, udpCallback)
	if err != nil {
		log.Printf("err:%v", err)
		return err
	}
	go ep.DispatchLoop()
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
