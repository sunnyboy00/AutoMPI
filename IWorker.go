package AutoMPI

// IWorker worker interface
type IWorker interface {
	// Get the guid of this worker
	GetGUID() string
	// add a message to this workers queue
	QueueMessage(MapMessage)
	// process all queued messsages
	ProcessMessages()
	// attache the AutoMPI.Node.Send(func(MapMessage)) method to this worker
	AttachSendMethod(func(MapMessage))
	// AttachNodeMethod attach a nodes.method into the worker to -
	AttachNodeMethod(string, func(interface{}) interface{})
	// AttachParent
	AttachParentNode(*Node)
	// do work
	DoWork()
	// close the worker
	Close()
	// get the age of the worker
	GetAge() string
}
