package AutoMPI

import (
	"fmt"
	"strconv"
	"time"
)

// WorkerTemplate template worker
type WorkerTemplate struct {
	GUID        string
	Group       string
	CreatedAt   time.Time
	MessageList []MapMessage
	Send        func(MapMessage)
	LastDidWork time.Time
}

// CreateWorkerTemplate as a template
func CreateWorkerTemplate(workerGUID string) IWorker {
	worker := new(WorkerTemplate)
	worker.GUID = workerGUID
	worker.Group = ""
	worker.CreatedAt = time.Now()
	worker.LastDidWork = time.Now()
	worker.MessageList = make([]MapMessage, 0)
	return worker
}

// GetGUID of the worker
func (base WorkerTemplate) GetGUID() string {
	return base.GUID
}

// GetGroup of the worker
func (base WorkerTemplate) GetGroup() string {
	return base.Group
}

// GetAge age of fhe link
func (base WorkerTemplate) GetAge() string {
	return strconv.FormatFloat(time.Now().Sub(base.CreatedAt).Seconds(), 'f', 0, 64)
}

func (base *WorkerTemplate) getDeltaTime() int64 {
	DeltaTime := time.Now().Sub(base.LastDidWork).Nanoseconds()
	base.LastDidWork = time.Now()
	return DeltaTime
}

// AttachSendMethod of the
func (base *WorkerTemplate) AttachSendMethod(parentsSendFunction func(MapMessage)) {
	base.Send = parentsSendFunction
}

// QueueMessage Queue messge for the worker
func (base *WorkerTemplate) QueueMessage(Message MapMessage) {
	base.MessageList = append(base.MessageList, Message)
}

// ProcessMessages process all messages for this worker
func (base *WorkerTemplate) ProcessMessages() {
	for len(base.MessageList) > 0 {
		Message := base.MessageList[0]
		base.MessageList = base.MessageList[1:]
		Message.DestinationGUID = ""
		// Process Message
	}
}

// DoWork do the work of the worker
func (base *WorkerTemplate) DoWork() {

	time.Sleep(50 * time.Millisecond)
}

// Close the Worker
func (base *WorkerTemplate) Close() {
	fmt.Println(base.GUID, " - Worker Closed")
}
