package AutoMPI

import (
	"fmt"
	"net"
	"sync"
	"time"
)

const (
	localLoggingFlag  = true
	heartBeatInterval = 1
	receivingPort     = ":8888"
	hTTPStatePort     = ":8800"
)

// Node base struct for the AutoMPI node
type Node struct {
	HeartBeatMessage       MapMessage
	heartBeating           bool
	CurrentMastersGUID     string
	MyNodeGUID             string
	LocalAddressString     string
	MyLinkReceivingAddress *net.TCPAddr

	BoardCaster UDPBoardcaster

	ExternalMessageProcessors []func(MapMessage, *Node)

	LocalListener *net.TCPListener
	listening     bool

	IncommingLinks map[string]*NodeLink
	OutgoingLinks  map[string]*NodeLink

	LocalWorkersLocation           map[string]*WorkerLocation
	LocalWorkersAnnouncingLocation map[string]int
	AllWorkersLocation             map[string]*WorkerLocation

	Workers        map[string]IWorker
	WorkerLock     sync.RWMutex
	WorkersWorking bool
}

// CreateNode create a Node of the AutoMPI
func CreateNode(
	// GUID of the node
	MyNodeGUID string,
	// Local ip address for the node IE: "192.168.1.50"
	LocalAddress string,
	// Primary external message handler
	ExternalMessageProcessor func(MapMessage, *Node)) *Node {

	base := new(Node)
	base.MyNodeGUID = MyNodeGUID
	base.LocalAddressString = LocalAddress

	base.IncommingLinks = map[string]*NodeLink{}
	base.OutgoingLinks = map[string]*NodeLink{}
	base.ExternalMessageProcessors = make([]func(MapMessage, *Node), 0)
	base.LocalWorkersLocation = map[string]*WorkerLocation{}
	base.LocalWorkersAnnouncingLocation = map[string]int{}
	base.AllWorkersLocation = map[string]*WorkerLocation{}
	base.Workers = map[string]IWorker{}

	base.AttachExternalMessageHandler(ExternalMessageProcessor)

	var err error
	err = nil
	println("Listening on: ", LocalAddress+receivingPort, "\n")
	base.MyLinkReceivingAddress, err = net.ResolveTCPAddr("tcp", LocalAddress+receivingPort)
	if err != nil {
		println("MyLinkReceivingAddress ResolveTCPAddr failed: ", err)
	}
	base.LocalListener, err = net.ListenTCP("tcp", base.MyLinkReceivingAddress)
	if err != nil {
		println("TCPListener failed: ", err)
	}

	base.BoardCaster.Bind(base.messageHandler)
	base.createHeartbeatMessage(LocalAddress + receivingPort)

	base.heartBeating = true
	go base.heartBeatLoop()

	base.listening = true
	go base.ListenerLoopForIncommingLink()

	base.WorkersWorking = true
	go base.workerWorkLoop()
	go base.workerProcessMessagesLoop()
	fmt.Printf("Workers goroutine started\n")

	go base.stateServe()
	fmt.Printf("HTTP state server started\n")

	return base
}

// ListenerLoopForIncommingLink listens for new incomming links untill base.listening is set false
func (base *Node) ListenerLoopForIncommingLink() {
	for base.listening {
		NewConnection, _ := base.LocalListener.Accept()
		CreateNodelinkReceiveIncomingConnection(NewConnection, base.MyNodeGUID, base.messageHandler, base.attachIncommingLink, base.detachIncommingLink)
	}
}

// messageHandler node message handler
// Sorts messages in the order of AutoMPI system message, Worker message, Node message
func (base *Node) messageHandler(Message MapMessage) {
	switch {
	case Message.IsAutoMPISystemMessage():
		{
			if base.listening {
				switch Message.GetValue(SystemKeysAutoMPISystemMessage) {
				case SystemKeyDetailsNodeLocation:
					if base.MyNodeGUID != Message.GetValue(SystemMessageDataPartGUIDNode) {
						base.ConnectToHost(Message.GetValue(SystemMessageDataPartGUIDNode), Message.GetValue(SystemMessageDataPartHOSTIPEP))
					}
					break
				case SystemKeyDetailsWorkerLocation:
					if !base.isALocalWorker(Message.GetValue(SystemMessageDataPartGUIDWorker)) {
						base.processWorkerLocation(Message)
					}
					break
				case SystemKeyDetailsConnectToMe:
					if base.MyNodeGUID != Message.GetValue(SystemMessageDataPartGUIDNode) {
						base.ConnectToHost(Message.GetValue(SystemMessageDataPartGUIDNode), Message.GetValue(SystemMessageDataPartHOSTIPEP))
					}
					break
				default:
					fmt.Println("unknown SystemKeysAutoMPISystemMessage")
					for key, value := range Message.Message {
						fmt.Println("Key:", key, "Value:", value)

					}
					//panic("unknown SystemKeysAutoMPISystemMessage")
				}
			}
		}
	case base.isALocalWorker(Message.DestinationGUID):
		{
			base.Workers[Message.DestinationGUID].QueueMessage(Message)
		}
		break
	case Message.DestinationGUID == DestinationAllWorkers:
		for _, value := range base.Workers {
			value.QueueMessage(Message)
		}
		break
	case Message.DestinationGUID == DestinationAllNodes:
		fallthrough
	default:
		{
			for index := range base.ExternalMessageProcessors {
				base.ExternalMessageProcessors[index](Message, base)
			}
		}
		break
	}
}

// AttachExternalMessageHandler attaches an additional external message handler
// multiple handlers can be attached
func (base *Node) AttachExternalMessageHandler(MessageHandlerFunction func(MapMessage, *Node)) {
	base.ExternalMessageProcessors = append(base.ExternalMessageProcessors, MessageHandlerFunction)
}

// ConnectToHost connect to the host of this GUID and IP address, Most cases only called from inside the Node
func (base *Node) ConnectToHost(TheirGUID string, RemoteIPEndPoint string) {

	_, ok := base.OutgoingLinks[TheirGUID]
	if ok {
		return
	}

	_, connected := CreateNodelinkOutgoingConnection(base.LocalAddressString, RemoteIPEndPoint, base.MyNodeGUID, base.messageHandler, base.attachOutgoingLink, base.detachOutgoingLink)
	if !connected {
		println("Failed to link with: ", base.MyNodeGUID, " X ", TheirGUID)
	}
}

// Send a message
// Current order of destination check: All, Local, Worker, Node
// Note destination "all by default will only be delivered to the Node's external handler
func (base *Node) Send(Message MapMessage) {
	switch {
	case Message.DestinationGUID == DestinationAllNodes:
		fallthrough
	case Message.DestinationGUID == DestinationAllWorkers:
		base.BoardCaster.Boardcast(Message)
		break
	case base.isALocalWorker(Message.DestinationGUID):
		base.Workers[Message.DestinationGUID].QueueMessage(Message)
		break
	default:
		parentNode, IsWorkerAndKnowsLocation := base.getHostingNodeOfWorkerBySearchingCollections(Message.DestinationGUID) // test as Worker location
		if IsWorkerAndKnowsLocation {
			link, ok := base.OutgoingLinks[parentNode]
			if ok {
				link.Send(Message)
			}
		} else { // try Node as destination
			link, ok := base.OutgoingLinks[Message.DestinationGUID]
			if ok {
				link.Send(Message)
			}
		}

	}
}

// Close the node
func (base *Node) Close() {

	base.WorkersWorking = false

	for _, value := range base.Workers {
		base.DetachWorker(value.GetGUID())
	}

	base.heartBeating = false

	base.listening = false
	base.LocalListener.Close()

	base.CloseAllLinks()

}

// CloseAllLinks close all outgoing and incomming links
// called by Close, but could also be a method of reinitilisation without closing all Workers
func (base *Node) CloseAllLinks() {
	for _, element := range base.OutgoingLinks {
		element.Close()
	}
	for _, element := range base.IncommingLinks {
		element.Close()
	}
}

// create the stanndard hartbeat message for this node
func (base *Node) createHeartbeatMessage(MyIPEndPoint string) {
	HeartBeatMessage := make(map[string]string)
	HeartBeatMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsNodeLocation
	HeartBeatMessage[SystemMessageDataPartGUIDNode] = base.MyNodeGUID
	HeartBeatMessage[SystemMessageDataPartHOSTIPEP] = MyIPEndPoint

	base.HeartBeatMessage.SetMessage(HeartBeatMessage)
}

func (base *Node) heartBeatLoop() {
	for base.heartBeating {
		base.BoardCaster.Boardcast(base.HeartBeatMessage)
		base.announceAllAgentsInLocalAgentsAnnouncing()
		time.Sleep(time.Second * heartBeatInterval)
	}
}

func (base *Node) attachIncommingLink(TheirGUID string, Link *NodeLink) {
	base.IncommingLinks[TheirGUID] = Link
	println("Attached Incoming link: ", base.MyNodeGUID, " <- ", Link.GUIDSource)
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.MyNodeGUID

	for key := range base.LocalWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.IncommingLinks[TheirGUID].Send(CreateMapMessage(TemplateMessage))
	}
}
func (base *Node) attachOutgoingLink(TheirGUID string, Link *NodeLink) {
	base.OutgoingLinks[TheirGUID] = Link
	println("Attached Outgoing link: ", base.MyNodeGUID, " -> ", Link.GUIDDestination)
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.MyNodeGUID

	for key := range base.LocalWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.OutgoingLinks[TheirGUID].Send(CreateMapMessage(TemplateMessage))
	}
}

// IsOutgoingLink is there a link to the Node
func (base *Node) IsOutgoingLink(TheirGUID string) bool {
	if _, ok := base.OutgoingLinks[TheirGUID]; ok {
		return true
	}
	return false
}

func (base *Node) detachIncommingLink(TheirGUID string) {
	_, ok := base.IncommingLinks[TheirGUID]
	if ok {
		delete(base.IncommingLinks, TheirGUID)
	}
	println("Detaching Incoming Link To:", TheirGUID)
}
func (base *Node) detachOutgoingLink(TheirGUID string) {
	_, ok := base.OutgoingLinks[TheirGUID]
	if ok {
		delete(base.OutgoingLinks, TheirGUID)
	}
	println("Detaching Outgoing Link To:", TheirGUID)
}
