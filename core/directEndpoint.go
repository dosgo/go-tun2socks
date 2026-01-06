package core

import (
	"io"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type DirectEndpoint struct {
	dev        io.ReadWriteCloser
	mtu        uint32
	dispatcher stack.NetworkDispatcher
}

func NewDirectEndpoint(dev io.ReadWriteCloser, mtu uint32) *DirectEndpoint {
	return &DirectEndpoint{
		dev: dev,
		mtu: mtu,
	}
}

// WritePackets 实现了协议栈发包接口
// 性能提升点：协议栈直接回调，无需 chan 阻塞和上下文切换
func (e *DirectEndpoint) WritePackets(pkts stack.PacketBufferList) (int, tcpip.Error) {
	n := 0
	for _, pkt := range pkts.AsSlice() {
		// ToView().WriteTo 直接将数据流式写入 dev，减少了大块内存申请
		if _, err := pkt.ToView().WriteTo(e.dev); err != nil {
			return n, &tcpip.ErrAborted{}
		}
		n++
	}
	return n, nil
}

// DispatchLoop 负责从 dev 读取数据推入协议栈
func (e *DirectEndpoint) DispatchLoop() error {
	// 性能提升点：在循环外预分配 buffer，避免 GC 压力
	buf := make([]byte, e.mtu+80)
	for {
		n, err := e.dev.Read(buf)
		if err != nil {
			return err
		}

		// 使用引用的方式构造数据包，减少拷贝
		pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
			Payload: buffer.MakeWithData(append([]byte(nil), buf[:n]...)),
		})

		var proto tcpip.NetworkProtocolNumber
		switch header.IPVersion(buf) {
		case header.IPv4Version:
			proto = header.IPv4ProtocolNumber
		case header.IPv6Version:
			proto = header.IPv6ProtocolNumber
		default:
			pkt.DecRef()
			continue
		}

		if e.dispatcher != nil {
			e.dispatcher.DeliverNetworkPacket(proto, pkt)
		}
		pkt.DecRef()
	}
}

// 接口必须的其他方法 (保持默认即可)
func (e *DirectEndpoint) MTU() uint32                                  { return e.mtu }
func (e *DirectEndpoint) Capabilities() stack.LinkEndpointCapabilities { return stack.CapabilityNone }
func (e *DirectEndpoint) MaxHeaderLength() uint16                      { return 0 }
func (e *DirectEndpoint) LinkAddress() tcpip.LinkAddress               { return "" }
func (e *DirectEndpoint) Attach(d stack.NetworkDispatcher)             { e.dispatcher = d }
func (e *DirectEndpoint) IsAttached() bool                             { return e.dispatcher != nil }
func (e *DirectEndpoint) Wait()                                        {}
func (e *DirectEndpoint) ARPHardwareType() header.ARPHardwareType      { return header.ARPHardwareNone }

// AddHeader implements stack.LinkEndpoint.AddHeader.
func (*DirectEndpoint) AddHeader(*stack.PacketBuffer) {}
func (e *DirectEndpoint) Close() {
	e.dev.Close()
}

// ParseHeader implements stack.LinkEndpoint.ParseHeader.
func (*DirectEndpoint) ParseHeader(*stack.PacketBuffer) bool { return true }

// SetOnCloseAction implements stack.LinkEndpoint.
func (*DirectEndpoint) SetOnCloseAction(func()) {}

// SetLinkAddress implements stack.LinkEndpoint.SetLinkAddress.
func (e *DirectEndpoint) SetLinkAddress(addr tcpip.LinkAddress) {

}
func (e *DirectEndpoint) SetMTU(mtu uint32) {
	e.mtu = mtu
}
