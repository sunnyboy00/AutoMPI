package AutoMPI

import (
	"net"
	"time"
)

const (
	// BroardCastAddress details
	BroardCastAddress = "255.255.255.255"
	// BroardCastPort details
	BroardCastPort  = ":9999"
	maxDatagramSize = 8192
)

// UDPBoardcaster is part of the discovry protocol
type UDPBoardcaster struct {
	SendingUDPConn   net.Conn
	ReceivingUDPConn *net.UDPConn
	ServerRun        bool
}

// Bind sets up the UDP listener
func (base *UDPBoardcaster) Bind(h func(MapMessage)) {

	SenderAddress, err := net.ResolveUDPAddr("udp", BroardCastAddress+BroardCastPort)
	if err != nil {
		println("ResolveUDPAddr failed: ", err)
	}
	base.SendingUDPConn, err = net.DialUDP("udp", nil, SenderAddress)
	if err != nil {
		println("DialUDP failed: ", err)
	}

	base.ServerRun = true
	ServerAddress, err := net.ResolveUDPAddr("udp", BroardCastPort)
	if err != nil {
		println("ResolveUDPAddr failed: ", err)
	}
	base.ReceivingUDPConn, err = net.ListenUDP("udp", ServerAddress)
	if err != nil {
		println("ListenUDP failed: ", err)
	}
	base.ReceivingUDPConn.SetReadBuffer(maxDatagramSize)
	go base.listenLoop(h)
}
func (base *UDPBoardcaster) listenLoop(h func(MapMessage)) {
	for base.ServerRun {
		var b = make([]byte, maxDatagramSize)
		readBytes, _, err := base.ReceivingUDPConn.ReadFromUDP(b)
		if err != nil {
			println("ReadFromUDP failed: ", err)
		}

		Message := MapMessage{}
		Message.FromBytes(b[0:readBytes])
		h(Message)
	}
}

// PingStart send n number of pings "hello world every 2 seconds
func (base *UDPBoardcaster) PingStart(n int) {
	Message := MapMessage{}
	pingMessage := map[string]string{
		"Hello": "World",
	}
	Message.Message = pingMessage
	for i := 0; i < n; i++ {
		base.SendingUDPConn.Write(Message.ToBytes())
		time.Sleep(2 * time.Second)
	}
}

// Boardcast board cast the message
func (base *UDPBoardcaster) Boardcast(Message MapMessage) {
	base.SendingUDPConn.Write(Message.ToBytes())

}

// Close close the UDP listener
func (base *UDPBoardcaster) Close() {
	base.ServerRun = false
	base.ReceivingUDPConn.Close()
	base.SendingUDPConn.Close()
}
