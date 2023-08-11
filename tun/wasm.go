// +build wasm

package tun

import (
	"github.com/songgao/water"
	"os/exec"
)

func GetWaterConf(tunAddr string, tunMask string) water.Config {
	config := water.Config{
		DeviceType: water.TUN,
	}
	return config
}

/*windows linux mac use tun dev*/
func RegTunDev(tunDevice string, tunAddr string, tunMask string, tunGW string, tunDNS string) (*water.Interface, error) {

	config := GetWaterConf(tunAddr, tunMask)
	ifce, err := water.New(config)
	return ifce, err
}

func CmdHide(name string, arg ...string) *exec.Cmd {
	return exec.Command(name, arg...)
}
