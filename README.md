
![alt text](AutoMPI.png)


# AutoMPI


# Distributed Service Platform #
### real-time data processing service platform in Golang

The purpose of this platform is to process large amounts of real-time data flows by working with data in atomic collections.
Keeping the state of the atomic collection private, any changes in state is updated to relevant workers on a need to know basis.

Nodes act as a host for workers.
Workers act as agents for the data.
The state of data is private to the worker.

Nodes provide an easy to use interface to the cluster to the workers. Removing the need for a worker to know where a recipent of a message is and how to deliver the message to the destination. Distributing workers across Nodes to balance workloads over a spread of devices and minimising data traffic through passing messages to relavant workers only.

Features implemented
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

```Go
node := CreateNode(
	// GUID of this node
	"NodeGUID00001", 
	// Local address of this node
	"192.168.1.20", 
	// An external message handler to process application messages
	msgHandler)
```

more message handlers can be attached with the attach function.
```Go
node.AttachExternalMessageHandler(msgHandler)
```

Message handler 
```Go
func msgHandler(Message AutoMPI.MapMessage, node *AutoMPI.Node) {}
```

## Workers 

Workers follow the IWorker interface 

```Go
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
An Example can be found in the 'WorkerTemplate.go' code file

Once the Node is running workers can be attached with the attach method

```Go
node.AttachWorker(AutoMPI.CreateWorkerTemplate("TemplateWorker0001"))
```


## AutoMPI Messages 

at the core of AutoMPI are messages which act both as command messages but also data.
Only the DestinationGUID of the message needs to be initialized for a message to be sent. 

```Go
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


