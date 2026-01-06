# tun2socks

A high-performance tun2socks implementation written in Go, powered by gVisor netstack.

# Usage

### Standard TUN Device
You can start a standard TUN device and proxy all traffic to a SOCKS5 server:
```go
var tunDevice = "tun0";
var tunAddr = "10.0.0.2";
var tunMask = "255.255.255.0";
var tunGw = "10.0.0.1";
var mtu = 1500;
var socksAddr = "127.0.0.1:1080";
var tunDns = "8.8.8.8:53";

tun2socks.StartTunDevice(tunDevice, tunAddr, tunMask, tunGw, mtu, socksAddr, tunDns);
go```
### Custom I/O (e.g., Android VpnService)
For environments like Android where you already have a File Descriptor or a custom stream, use the ForwardTransportFromIo interface:

// dev is an object that implements io.ReadWriteCloser
tun2socks.ForwardTransportFromIo(dev, mtu, rawTcpForwarder, rawUdpForwarder);

# Thank
* github.com/google/gvisor
* github.com/miekg/dns
