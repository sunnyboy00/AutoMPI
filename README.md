# AutoMPI

Future home of real-time distributed service platform in Golang, Will be updated once the core of the platform has been ported from C#


Key features implemented
* Autodiscovery of local Nodes
* Establish links between Nodes
* Cleaning of broken links
* passing of JSON Messages 
* Node / Worker ( Host / Agent ) archicture 

Untested
* Worker - Worker performance 

Yet to be implemented
* Storage interface 

# How To use #


Create a Node of the AutoMPI, and attach an external message handler

<code>
node := CreateNode("NodeGUID00001", "192.168.1.20", msgHandler)
</code>
Parameters supplied 
GUID of this node
Local address of this node
An external message handler to process application messages

More messages can be attached with the attach function.
<code>
node.AttachExternalMessageHandler(msgHandler)
</code>

Message handler 
<code>
func msgHandler(Message AutoMPI.MapMessage, node *AutoMPI.Node) {}
</code>

Once the Node is running workers can be attached with the attach method
<code>
node.AttachWorker(AutoMPI.CreateWorkerTemplate("TemplateWorker0001"))
</code>


# Extended how to #

After the Node is setup and any static workers are created the primary methods used are the AutoMPI.Node.Send(MapMessage) and the attached message handler

* AutoMPI.Node.Send(MapMessage) to send messaes (commands) to other nodes
* func msgHandler(Message AutoMPI.MapMessage, node *AutoMPI.Node) {} to receive messages (commands) from other nodes


# AutoMPI Messages #

at the core of AutoMPI are messages which act both as command messages but also data.
Only the DestinationGUID of the message needs to be initialized for a message to be sent. 
<code>
type MapMessage struct {
	DestinationGUID string
	SourceGUID      string
	Message         map[string]string
	Data            []byte
}
</code>

Messages are checked(and passed) in this order when received by a node
* AutoMPI system message
* Worker Message
* Mode Message (Passed to external message handler(s))


