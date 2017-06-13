package AutoMPI

// IWorker worker interface
type IWorker interface {
	// Get the guid of this worker
	GetGUID() string
	// Get the group of this worker
	GetGroup() string
	// add a message to this workers queue
	QueueMessage(MapMessage)
	// process all queued messsages
	ProcessMessages()
	// attache the AutoMPI.Node.Send(func(MapMessage)) method to this worker
	AttachSendMethod(func(MapMessage))
	// do work
	DoWork()
	// close the worker
	Close()
	// get the age of the worker
	GetAge() string
}
