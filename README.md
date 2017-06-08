
![alt text](AutoMPI.png)


# AutoMPI


# Distributed Service Platform #
### distributed real-time computation service platform in Golang
in the process of being ported from C#

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

Example workers
* WorkerTemplate
* WorkerExampleIPCamera
* WorkerExampleKeyValueStore (ytbi)
* WorkerExampleScheduler (ytbi)

## How To use 

Create a Node of the AutoMPI, and attach an external message handler

```
node := CreateNode("NodeGUID00001", "192.168.1.20", msgHandler)
```

Parameters supplied 
GUID of this node
Local address of this node
An external message handler to process application messages

More messages can be attached with the attach function.
```
node.AttachExternalMessageHandler(msgHandler)
```

Message handler 
```
func msgHandler(Message AutoMPI.MapMessage, node *AutoMPI.Node) {}
```

## Workers 

Workers follow the IWorker interface 

```
type IWorker interface {
	// Get the guid of this worker
 	GetGUID() string
 	// add a message to this workers queue
 	QueueMessage(MapMessage)
 	// attache the AutoMPI.Node.Send(func(MapMessage)) method to this worker
 	AttachSendMethod(func(MapMessage))
 	// do work, passing the nanoseconds since the last call
 	DoWork()
 	// close the worker
 	Close()
 	// get the age of the worker
 	GetAge() string
}
```

Once the Node is running workers can be attached with the attach method

```
node.AttachWorker(AutoMPI.CreateWorkerTemplate("TemplateWorker0001"))
```


## AutoMPI Messages 

at the core of AutoMPI are messages which act both as command messages but also data.
Only the DestinationGUID of the message needs to be initialized for a message to be sent. 

```
type MapMessage struct {
	DestinationGUID string
	SourceGUID      string
	Message         map[string]string
	Data            []byte
}
```


After the Node is setup and any static workers are created the primary methods used on a node are the AutoMPI.Node.Send(MapMessage) and any attached message handler(s)

* AutoMPI.Node.Send(MapMessage) to send messaes (commands) to other nodes
* func msgHandler(Message AutoMPI.MapMessage, node *AutoMPI.Node) {} to receive messages (commands) from other nodes

Messages are checked(and passed) in this order when received by a node
* AutoMPI system message
* Worker Message
* Mode Message (Passed to external message handler(s))


