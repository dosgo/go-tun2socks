package socks

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

/*to socks5*/
func SocksCmd(socksConn net.Conn, cmd uint8, host string) error {
	//socks5 auth
	socksConn.Write([]byte{0x05, 0x01, 0x00})
	authBack := make([]byte, 2)
	_, err := io.ReadFull(socksConn, authBack)
	if err != nil {
		log.Println(err)
		return err
	}
	//connect head
	hosts := strings.Split(host, ":")
	rAddr := net.ParseIP(hosts[0])
	_port, _ := strconv.Atoi(hosts[1])
	msg := []byte{0x05, cmd, 0x00, 0x01}
	buffer := bytes.NewBuffer(msg)
	//ip
	binary.Write(buffer, binary.BigEndian, rAddr.To4())
	//port
	binary.Write(buffer, binary.BigEndian, uint16(_port))
	socksConn.Write(buffer.Bytes())
	conectBack := make([]byte, 10)
	_, err = io.ReadFull(socksConn, conectBack)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// SocksUdpCmd 发起 UDP ASSOCIATE 请求并返回代理服务器分配的 UDP 监听地址
func SocksUdpCmd(socksConn net.Conn, host string) (net.Addr, error) {
	// 1. 认证 (保持不变)
	socksConn.Write([]byte{0x05, 0x01, 0x00})
	authBack := make([]byte, 2)
	if _, err := io.ReadFull(socksConn, authBack); err != nil {
		return nil, err
	}

	// 2. 发送 UDP ASSOCIATE 请求 (cmd = 0x03)
	// host 通常传 "0.0.0.0:0"
	hosts := strings.Split(host, ":")
	rAddr := net.ParseIP(hosts[0])
	_port, _ := strconv.Atoi(hosts[1])

	msg := []byte{0x05, 0x03, 0x00, 0x01} // 0x03 是 UDP ASSOCIATE
	buffer := bytes.NewBuffer(msg)
	binary.Write(buffer, binary.BigEndian, rAddr.To4())
	binary.Write(buffer, binary.BigEndian, uint16(_port))
	socksConn.Write(buffer.Bytes())

	// 3. 解析返回结果
	// 前 4 字节是固定头，后面是 BND.ADDR 和 BND.PORT
	replyHead := make([]byte, 4)
	if _, err := io.ReadFull(socksConn, replyHead); err != nil {
		return nil, err
	}

	if replyHead[1] != 0x00 {
		return nil, log.Output(0, "Socks5 UDP Associate failed")
	}

	// 解析地址 (BND.ADDR)
	var bndIP net.IP
	addrType := replyHead[3]
	switch addrType {
	case 0x01: // IPv4
		ipv4 := make([]byte, 4)
		io.ReadFull(socksConn, ipv4)
		bndIP = net.IP(ipv4)
	case 0x03: // Domain
		domainLen := make([]byte, 1)
		io.ReadFull(socksConn, domainLen)
		domain := make([]byte, int(domainLen[0]))
		io.ReadFull(socksConn, domain)
		// 简单处理：UDP 通常需要明确的 IP
	}

	// 解析端口 (BND.PORT)
	portBuf := make([]byte, 2)
	io.ReadFull(socksConn, portBuf)
	bndPort := binary.BigEndian.Uint16(portBuf)

	return &net.UDPAddr{IP: bndIP, Port: int(bndPort)}, nil
}

// EncodeUDPPacket 封装 Socks5 UDP 头部
func EncodeUDPPacket(targetHost string, data []byte) []byte {
	hosts := strings.Split(targetHost, ":")
	rAddr := net.ParseIP(hosts[0]).To4()
	_port, _ := strconv.Atoi(hosts[1])

	msg := []byte{0x00, 0x00, 0x00, 0x01} // RSV, FRAG, ATYP(IPv4)
	buffer := bytes.NewBuffer(msg)
	buffer.Write(rAddr)
	binary.Write(buffer, binary.BigEndian, uint16(_port))
	buffer.Write(data)
	return buffer.Bytes()
}

// DecodeUDPPacket 解封装 Socks5 UDP 头部，返回数据部分
func DecodeUDPPacket(packet []byte) ([]byte, error) {
	if len(packet) < 10 {
		return nil, log.Output(0, "packet too short")
	}
	// 简单实现：跳过前 3 字节，根据第 4 字节 ATYP 判断地址长度
	// 这里假设是 IPv4 (0x01)，头部长度为 10 字节
	return packet[10:], nil
}
