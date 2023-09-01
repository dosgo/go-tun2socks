//go:build !windows && !wasm
// +build !windows,!wasm

package tun

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/songgao/water"
	"golang.zx2c4.com/wireguard/tun"
)

func GetCidrIpRange(cidr string) (string, string) {
	ip := strings.Split(cidr, "/")[0]
	ipSegs := strings.Split(ip, ".")
	maskLen, _ := strconv.Atoi(strings.Split(cidr, "/")[1])
	seg3MinIp, seg3MaxIp := getIpSeg3Range(ipSegs, maskLen)
	seg4MinIp, seg4MaxIp := getIpSeg4Range(ipSegs, maskLen)
	ipPrefix := ipSegs[0] + "." + ipSegs[1] + "."

	return ipPrefix + strconv.Itoa(seg3MinIp) + "." + strconv.Itoa(seg4MinIp),
		ipPrefix + strconv.Itoa(seg3MaxIp) + "." + strconv.Itoa(seg4MaxIp)
}

// 得到第四段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg4Range(ipSegs []string, maskLen int) (int, int) {
	ipSeg, _ := strconv.Atoi(ipSegs[3])
	segMinIp, segMaxIp := getIpSegRange(uint8(ipSeg), uint8(32-maskLen))
	return segMinIp + 1, segMaxIp
}

// 得到第三段IP的区间（第一片段.第二片段.第三片段.第四片段）
func getIpSeg3Range(ipSegs []string, maskLen int) (int, int) {
	if maskLen > 24 {
		segIp, _ := strconv.Atoi(ipSegs[2])
		return segIp, segIp
	}
	ipSeg, _ := strconv.Atoi(ipSegs[2])
	return getIpSegRange(uint8(ipSeg), uint8(24-maskLen))
}

// 根据用户输入的基础IP地址和CIDR掩码计算一个IP片段的区间
func getIpSegRange(userSegIp, offset uint8) (int, int) {
	var ipSegMax uint8 = 255
	netSegIp := ipSegMax << offset
	segMinIp := netSegIp & userSegIp
	segMaxIp := userSegIp&(255<<offset) | ^(255 << offset)
	return int(segMinIp), int(segMaxIp)
}

func GetWaterConf(tunAddr string, tunMask string) water.Config {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun2"
	return config
}

/*windows linux mac use tun dev*/
func RegTunDev(tunDevice string, tunAddr string, tunMask string, tunGW string, tunDNS string) (*water.Interface, error) {
	if len(tunDevice) == 0 {
		tunDevice = "tun0"
	}
	if len(tunAddr) == 0 {
		tunAddr = "10.0.0.2"
	}
	if len(tunMask) == 0 {
		tunMask = "255.255.255.0"
	}
	if len(tunGW) == 0 {
		tunGW = "10.0.0.1"
	}
	if len(tunDNS) == 0 {
		tunDNS = "114.114.114.114"
	}

	config := GetWaterConf(tunAddr, tunMask)
	ifce, err := water.New(config)
	if err != nil {
		fmt.Println("start tun err:", err)
		return nil, err
	}

	if runtime.GOOS == "linux" {
		//sudo ip addr add 10.1.0.10/24 dev O_O
		masks := net.ParseIP(tunMask).To4()
		maskAddr := net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3])}
		CmdHide("ip", "addr", "add", maskAddr.String(), "dev", ifce.Name()).Run()
		CmdHide("ip", "link", "set", "dev", ifce.Name(), "up").Run()
	} else if runtime.GOOS == "darwin" {
		//ifconfig utun2 10.1.0.10 10.1.0.20 up
		masks := net.ParseIP(tunMask).To4()
		maskAddr := net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3])}
		ipMin, ipMax := GetCidrIpRange(maskAddr.String())
		CmdHide("ifconfig", "utun2", ipMin, ipMax, "up").Run()
	}
	return ifce, nil
}

/*windows use wintun*/
func RegTunDevTest(tunDevice string, tunAddr string, tunMask string, tunGW string, tunDNS string) (*DevReadWriteCloser, error) {
	if len(tunDevice) == 0 {
		tunDevice = "socksTun0"
	}
	if len(tunAddr) == 0 {
		tunAddr = "10.0.0.2"
	}
	if len(tunMask) == 0 {
		tunMask = "255.255.255.0"
	}
	if len(tunGW) == 0 {
		tunGW = "10.0.0.1"
	}
	if len(tunDNS) == 0 {
		tunDNS = "114.114.114.114"
	}
	tunDev, err := tun.CreateTUN(tunDevice, 1500)
	if err != nil {
		return nil, err
	}
	tunDevName, _ := tunDev.Name()
	if runtime.GOOS == "linux" {
		//sudo ip addr add 10.1.0.10/24 dev O_O
		masks := net.ParseIP(tunMask).To4()
		maskAddr := net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3])}
		CmdHide("ip", "addr", "add", maskAddr.String(), "dev", tunDevName).Run()
		CmdHide("ip", "link", "set", "dev", tunDevName, "up").Run()
	} else if runtime.GOOS == "darwin" {
		//ifconfig utun2 10.1.0.10 10.1.0.20 up
		masks := net.ParseIP(tunMask).To4()
		maskAddr := net.IPNet{IP: net.ParseIP(tunAddr), Mask: net.IPv4Mask(masks[0], masks[1], masks[2], masks[3])}
		ipMin, ipMax := GetCidrIpRange(maskAddr.String())
		CmdHide("ifconfig", tunDevName, ipMin, ipMax, "up").Run()
	}
	return &DevReadWriteCloser{tunDev.(*tun.NativeTun)}, nil
}

func CmdHide(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
