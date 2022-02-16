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
