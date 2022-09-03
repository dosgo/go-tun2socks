// +build !windows

package tun

import (
	"fmt"
	"github.com/dosgo/xsocks/comm"
	"github.com/songgao/water"
	"net"
	"runtime"
)


/*windows linux mac use tun dev*/
func RegTunDev(tunDevice string,tunAddr string,tunMask string,tunGW string,tunDNS string)(*water.Interface,error){
	if len(tunDevice)==0 {
		tunDevice="tun0";
	}
	if len(tunAddr)==0 {
		tunAddr="10.0.0.2";
	}
	if len(tunMask)==0 {
		tunMask="255.255.255.0";
	}
	if len(tunGW)==0 {
		tunGW="10.0.0.1";
	}
	if len(tunDNS)==0 {
		tunDNS="114.114.114.114";
	}

	config := comm.GetWaterConf(tunAddr,tunMask);
	ifce, err := water.New(config)
	if err != nil {
		fmt.Println("start tun err:", err)
		return nil,err;
	}
	//set ifco conf
	if runtime.GOOS=="windows" {
		comm.CmdHide("netsh", "interface","ip","set","address","name="+ifce.Name(),"source=static","addr="+tunAddr,"mask="+tunMask,"gateway=none").Run();
	}else if runtime.GOOS=="linux"{
		//sudo ip addr add 10.1.0.10/24 dev O_O
		masks:=net.ParseIP(tunMask).To4();
		maskAddr:=net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3] )}
		comm.CmdHide("ip", "addr","add",maskAddr.String(),"dev",ifce.Name()).Run();
		comm.CmdHide("ip", "link","set","dev",ifce.Name(),"up").Run();
	}else if runtime.GOOS=="darwin"{
		//ifconfig utun2 10.1.0.10 10.1.0.20 up
		masks:=net.ParseIP(tunMask).To4();
		maskAddr:=net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3] )}
		ipMin,ipMax:=comm.GetCidrIpRange(maskAddr.String());
		comm.CmdHide("ifconfig", "utun2",ipMin,ipMax,"up").Run();
	}
	return ifce,nil;
}


