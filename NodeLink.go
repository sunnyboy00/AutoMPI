package AutoMPI

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	bufferSize            = 8192
	sleepTimeBetweenReads = 5
)

// NodeLink primary link between nodes
type NodeLink struct {
	GUIDSource, GUIDDestination, MyGUID, TheirGUID string

	LocalAddress  *net.TCPAddr
	RemoteAddress *net.TCPAddr

	Receiving               bool
	Connection              net.Conn
	ConnectionEstablishedAt time.Time

	SendLock sync.Mutex

	MessageHandlerFunction func(MapMessage)

	Attach func(string, *NodeLink)
	Detach func(string)
}

// CreateNodelinkOutgoingConnection the Link to another Node
func CreateNodelinkOutgoingConnection(LocalAddress string, RemoteAddress string, MyGUID string, MessageHandler func(MapMessage), Attach func(string, *NodeLink), Detach func(string)) (*NodeLink, bool) {
	base := new(NodeLink)
	base.GUIDSource = MyGUID
	base.MyGUID = MyGUID
	base.ConnectionEstablishedAt = time.Now()
	var err error

	port := rand.Intn(65534-49152) + 49152

	base.LocalAddress, err = net.ResolveTCPAddr("tcp", LocalAddress+":"+strconv.Itoa(port))
	if err != nil {
		println("NodeLink Local ResolveTCPAddr failed: ", LocalAddress+":"+strconv.Itoa(port), err)
	}
	base.RemoteAddress, err = net.ResolveTCPAddr("tcp", RemoteAddress)
	if err != nil {
		println("NodeLink Remote ResolveTCPAddr failed: ", err)
	}

	base.MessageHandlerFunction = MessageHandler
	base.Attach = Attach
	base.Detach = Detach

	base.Connection, err = net.DialTCP("tcp", base.LocalAddress, base.RemoteAddress)
	if err != nil {
		println("DialTCP failed: ", LocalAddress+":"+strconv.Itoa(port), RemoteAddress, err)
	}

	_, err = base.Connection.Write([]byte(base.GUIDSource))
	if err != nil {
		println("Write GUIDSource failed: ", err)
	}
	var result bool
	base.GUIDDestination, result = base.readString()
	base.TheirGUID = base.GUIDDestination
	if result {
		base.Attach(base.TheirGUID, base)
	} else {
		// panic
		return nil, false
	}

	base.Receiving = true
	go base.readLoop()
	return base, true
}

// CreateNodelinkReceiveIncomingConnection recive a connection from another node
func CreateNodelinkReceiveIncomingConnection(Connection net.Conn, MyGUID string, MessageHandler func(MapMessage), Attach func(string, *NodeLink), Detach func(string)) *NodeLink {
	base := new(NodeLink)
	base.ConnectionEstablishedAt = time.Now()

	base.MyGUID = MyGUID
	base.GUIDDestination = MyGUID

	var err error

	base.MessageHandlerFunction = MessageHandler
	base.Attach = Attach
	base.Detach = Detach

	base.Connection = Connection

	var result bool

	base.GUIDSource, result = base.readString()
	base.TheirGUID = base.GUIDSource

	_, err = base.Connection.Write([]byte(base.GUIDDestination))
	if err != nil {
		println("Write GUIDSource failed: ", err)
	}

	if result {
		base.Attach(base.TheirGUID, base)
	}

	base.Receiving = true
	go base.readLoop()

	return base
}

// GetRemoteAddressAsString get the address part of the remote address
func (base *NodeLink) GetRemoteAddressAsString() string {
	parts := strings.Split(base.Connection.RemoteAddr().String(), ":")
	return parts[0]
}

// GetRemoteAddressAndPortAsString full address and path
func (base *NodeLink) GetRemoteAddressAndPortAsString() string {
	return base.Connection.RemoteAddr().String()
}

// GetAge age of fhe link
func (base *NodeLink) GetAge() string {
	return strconv.FormatFloat(time.Now().Sub(base.ConnectionEstablishedAt).Seconds(), 'f', 0, 64)
}

func (base *NodeLink) readLoop() {
	for base.Receiving {
		MessageData, result := base.read()
		if result {
			Message := MapMessage{}
			Message.FromBytes(MessageData)
			base.MessageHandlerFunction(Message)
		}
	}
	time.Sleep(time.Millisecond * sleepTimeBetweenReads) // TODO work a better method
}

func (base *NodeLink) read() ([]byte, bool) {
	ReadBuffer := make([]byte, 0)
	if base.IsConnected() {
		//	println(" --- func (base *NodeLink) read() ([]byte, bool) --- ")
		lengthBuffer := make([]byte, 4)
		sizeBytesRead, err := base.Connection.Read(lengthBuffer)
		if base.breakOnFatelError(err) {
			base.Close()
			return make([]byte, 0), false
		}
		if sizeBytesRead != 4 {
			println("Error -> NodeLink sizeBytesRead: ", sizeBytesRead)
		}

		lengthToRead := int(binary.LittleEndian.Uint32(lengthBuffer))
		//	println(lengthToRead, "Bytes to read")

		if lengthToRead > 0 {
			readSoFar := 0
			var buffer bytes.Buffer

			for readSoFar < lengthToRead {
				ReadBuffer = makeBufferOfCorrectSize(lengthToRead, readSoFar)
				readBytes, err := base.Connection.Read(ReadBuffer)
				readSoFar += readBytes
				if base.breakOnFatelError(err) {
					base.Close()
					return make([]byte, 0), false
				}
				buffer.Write(ReadBuffer[0:readBytes])
			}
			if readSoFar != lengthToRead {
				println(readSoFar, "!=", lengthToRead)
			}

			if buffer.Len() > 0 {
				MessageData := buffer.Bytes()
				//		println(len(MessageData), "Bytes Read")
				return MessageData, true
			}
		}
	} else {
		base.Close()
	}
	return make([]byte, 0), false
}

func makeBufferOfCorrectSize(totalToRead int, readSoFar int) []byte {
	var ToReadThisTime int
	YetToRead := totalToRead - readSoFar
	if YetToRead >= bufferSize {
		ToReadThisTime = bufferSize
	} else {
		ToReadThisTime = YetToRead % bufferSize
	}
	/*	if ToReadThisTime != bufferSize {
		println(ToReadThisTime)
	}*/
	return make([]byte, ToReadThisTime)
}

// IsConnected is the link connected?
func (base *NodeLink) IsConnected() bool {
	one := []byte{}
	base.Connection.SetReadDeadline(time.Now())
	if _, err := base.Connection.Read(one); err == io.EOF {
		println("detected closed LAN connection", base.TheirGUID)
		base.Close()
		return false
	}
	base.Connection.SetReadDeadline(time.Time{})
	return true
}

// breakOnFatelError breaks when the connection has a error
func (base *NodeLink) breakOnFatelError(err error) bool {
	if err != nil {
		switch err {
		case io.ErrUnexpectedEOF:
			println(" io.ErrUnexpectedEOF", base.MyGUID)
			base.Close()
			return true
		case io.EOF:
			println("io.EOF", base.MyGUID)
			base.Close()
			return true
		case io.ErrNoProgress:
			println("ErrNoProgress", base.MyGUID)
			return false
		default:
			switch SubErr := err.(type) {
			case net.Error:
				if SubErr.Timeout() {
					fmt.Println("This was a net.Error with a Timeout")
					return false
				}
				break
			default:
				println("0ther Error", base.MyGUID)
				base.Close()
				break
			}
			return true
		}
	}
	return false
}

// Send data to the oppisate end of the Link
func (base *NodeLink) Send(Message MapMessage) {
	b := Message.ToBytes()
	bSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(bSize, uint32(len(b)))
	base.SendLock.Lock()
	//	println("Sending message to:", Message.DestinationGUID)
	base.Connection.Write(bSize)
	base.Connection.Write(b)
	base.SendLock.Unlock()
}

func (base *NodeLink) readString() (string, bool) {
	var buffer bytes.Buffer
	var b = make([]byte, bufferSize)

	readBytes, err := base.Connection.Read(b)
	if err != nil {

	} else {

		buffer.Write(b[:readBytes])

		if buffer.Len() > 0 {
			return buffer.String(), true
		}

	}
	return "", false
}

// Close Close the link
func (base *NodeLink) Close() {
	base.Receiving = false
	base.Connection.Close()
	base.Connection = nil
	base.Detach(base.TheirGUID)
}