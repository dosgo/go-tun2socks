package tun

import "golang.zx2c4.com/wireguard/tun"

type DevReadWriteCloser struct {
	tunDev *tun.NativeTun
}

func (conn DevReadWriteCloser) Read(buf []byte) (int, error) {
	data := [][]byte{buf}
	dataLen := []int{0}
	_, err := conn.tunDev.Read(data, dataLen, 0)
	return dataLen[0], err
}

func (conn DevReadWriteCloser) Write(buf []byte) (int, error) {
	data := [][]byte{buf}
	return conn.tunDev.Write(data, 0)
}

func (conn DevReadWriteCloser) Close() error {
	if conn.tunDev == nil {
		return nil
	}
	return conn.tunDev.Close()
}
