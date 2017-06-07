package AutoMPI

// IWorker worker interface
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
