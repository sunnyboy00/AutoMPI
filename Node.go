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
// call AutoMPI.CreateNode() to create and start
type Node struct {
	heartBeatMessage MapMessage
	heartBeating     bool

	GUID                   string
	localAddressString     string
	myLinkReceivingAddress *net.TCPAddr

	boardCaster UDPBoardcaster

	externalMessageProcessors []func(MapMessage, *Node)

	localListener *net.TCPListener
	listening     bool

	incommingLinks map[string]*NodeLink
	outgoingLinks  map[string]*NodeLink

	localWorkersLocation           map[string]*WorkerLocation
	localWorkersAnnouncingLocation map[string]int
	allWorkersLocation             map[string]*WorkerLocation
	allWorkersLocationLock         sync.RWMutex

	Workers        map[string]IWorker
	workerLock     sync.RWMutex
	workersWorking bool
}

// CreateNode create a Node of the AutoMPI
func CreateNode(
	// GUID of the node
	GUID string,
	// Local ip address for the node IE: "192.168.1.50"
	LocalAddress string,
	// Primary external message handler
	ExternalMessageProcessor func(MapMessage, *Node)) *Node {

	base := new(Node)
	base.GUID = GUID
	base.localAddressString = LocalAddress

	base.incommingLinks = map[string]*NodeLink{}
	base.outgoingLinks = map[string]*NodeLink{}
	base.externalMessageProcessors = make([]func(MapMessage, *Node), 0)
	base.localWorkersLocation = map[string]*WorkerLocation{}
	base.localWorkersAnnouncingLocation = map[string]int{}
	base.allWorkersLocation = map[string]*WorkerLocation{}
	base.Workers = map[string]IWorker{}

	base.AttachExternalMessageHandler(ExternalMessageProcessor)

	var err error
	err = nil
	println("Listening on: ", LocalAddress+receivingPort, "\n")
	base.myLinkReceivingAddress, err = net.ResolveTCPAddr("tcp", LocalAddress+receivingPort)
	if err != nil {
		println("myLinkReceivingAddress ResolveTCPAddr failed: ", err)
	}
	base.localListener, err = net.ListenTCP("tcp", base.myLinkReceivingAddress)
	if err != nil {
		println("TCPListener failed: ", err)
	}

	base.boardCaster.Bind(base.messageHandler)
	base.createheartBeatMessage(LocalAddress + receivingPort)

	base.heartBeating = true
	go base.heartBeatLoop()

	base.listening = true
	go base.listenerLoopForIncommingLink()

	base.workersWorking = true
	go base.workerWorkLoop()
	go base.workerProcessMessagesLoop()
	fmt.Printf("Workers goroutine started\n")

	go base.stateServe()
	fmt.Printf("HTTP state server started\n")

	return base
}

// listenerLoopForIncommingLink listens for new incomming links untill base.listening is set false
func (base *Node) listenerLoopForIncommingLink() {
	for base.listening {
		NewConnection, _ := base.localListener.Accept()
		CreateNodelinkReceiveIncomingConnection(NewConnection, base.GUID, base.messageHandler, base.attachIncommingLink, base.detachIncommingLink)
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
					if base.GUID != Message.GetValue(SystemMessageDataPartGUIDNode) {
						base.ConnectToHost(Message.GetValue(SystemMessageDataPartGUIDNode), Message.GetValue(SystemMessageDataPartHOSTIPEP))
					}
					break
				case SystemKeyDetailsWorkerLocation:
					if !base.isALocalWorker(Message.GetValue(SystemMessageDataPartGUIDWorker)) {
						base.processWorkerLocation(Message)
					}
					break
				case SystemKeyDetailsConnectToMe:
					if base.GUID != Message.GetValue(SystemMessageDataPartGUIDNode) {
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
			for index := range base.externalMessageProcessors {
				base.externalMessageProcessors[index](Message, base)
			}
		}
		break
	}
}

// AttachExternalMessageHandler attaches an additional external message handler
// multiple handlers can be attached
func (base *Node) AttachExternalMessageHandler(MessageHandlerFunction func(MapMessage, *Node)) {
	base.externalMessageProcessors = append(base.externalMessageProcessors, MessageHandlerFunction)
}

// ConnectToHost connect to the host of this GUID and IP address, Most cases only called from inside the Node
func (base *Node) ConnectToHost(TheirGUID string, RemoteIPEndPoint string) {

	_, ok := base.outgoingLinks[TheirGUID]
	if ok {
		return
	}

	_, connected := CreateNodelinkOutgoingConnection(base.localAddressString, RemoteIPEndPoint, base.GUID, base.messageHandler, base.attachOutgoingLink, base.detachOutgoingLink)
	if !connected {
		println("Failed to link with: ", base.GUID, " X ", TheirGUID)
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
		base.boardCaster.Boardcast(Message)
		break
	case base.isALocalWorker(Message.DestinationGUID):
		base.Workers[Message.DestinationGUID].QueueMessage(Message)
		break
	default:
		parentNode, IsWorkerAndKnowsLocation := base.getHostingNodeOfWorkerBySearchingCollections(Message.DestinationGUID) // test as Worker location
		if IsWorkerAndKnowsLocation {
			link, ok := base.outgoingLinks[parentNode]
			if ok {
				link.Send(Message)
			}
		} else { // try Node as destination
			link, ok := base.outgoingLinks[Message.DestinationGUID]
			if ok {
				link.Send(Message)
			}
		}

	}
}

// Close the node
func (base *Node) Close() {

	base.workersWorking = false

	for _, value := range base.Workers {
		base.DetachWorker(value.GetGUID())
	}

	base.heartBeating = false

	base.listening = false
	base.localListener.Close()

	base.CloseAllLinks()

}

// CloseAllLinks close all outgoing and incomming links
// called by Close, but could also be a method of reinitilisation without closing all Workers
func (base *Node) CloseAllLinks() {
	for _, element := range base.outgoingLinks {
		element.Close()
	}
	for _, element := range base.incommingLinks {
		element.Close()
	}
}

// create the stanndard hartbeat message for this node
func (base *Node) createheartBeatMessage(MyIPEndPoint string) {
	heartBeatMessage := make(map[string]string)
	heartBeatMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsNodeLocation
	heartBeatMessage[SystemMessageDataPartGUIDNode] = base.GUID
	heartBeatMessage[SystemMessageDataPartHOSTIPEP] = MyIPEndPoint

	base.heartBeatMessage.SetMessage(heartBeatMessage)
}

func (base *Node) heartBeatLoop() {
	for base.heartBeating {
		base.boardCaster.Boardcast(base.heartBeatMessage)
		base.announceAllAgentsInLocalAgentsAnnouncing()
		time.Sleep(time.Second * heartBeatInterval)
	}
}

func (base *Node) attachIncommingLink(TheirGUID string, Link *NodeLink) {
	base.incommingLinks[TheirGUID] = Link
	fmt.Printf("Attached Incoming link: %s <- %s \n", base.GUID, Link.GUIDSource)
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.GUID

	for key := range base.localWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.incommingLinks[TheirGUID].Send(CreateMapMessage(TemplateMessage))
	}
}
func (base *Node) attachOutgoingLink(TheirGUID string, Link *NodeLink) {
	base.outgoingLinks[TheirGUID] = Link
	fmt.Printf("Attached Outgoing link: %s -> %s \n", base.GUID, Link.GUIDDestination)
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.GUID

	for key := range base.localWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.outgoingLinks[TheirGUID].Send(CreateMapMessage(TemplateMessage))
	}
}

// IsOutgoingLink is there a link to the Node
func (base *Node) IsOutgoingLink(TheirGUID string) bool {
	if _, ok := base.outgoingLinks[TheirGUID]; ok {
		return true
	}
	return false
}

func (base *Node) detachIncommingLink(TheirGUID string) {
	_, ok := base.incommingLinks[TheirGUID]
	if ok {
		delete(base.incommingLinks, TheirGUID)
	}
	println("Detaching Incoming Link To:", TheirGUID)
}
func (base *Node) detachOutgoingLink(TheirGUID string) {
	_, ok := base.outgoingLinks[TheirGUID]
	if ok {
		delete(base.outgoingLinks, TheirGUID)
	}
	println("Detaching Outgoing Link To:", TheirGUID)
}
