package AutoMPI

import (
	"fmt"
	"strconv"
	"time"
)

// WorkerExampleKeyValueStore template worker
type WorkerExampleKeyValueStore struct {
	GUID        string
	Group       string
	CreatedAt   time.Time
	MessageList []MapMessage
	Send        func(MapMessage)
	LastDidWork time.Time
	DataStored  map[string][]byte
}

// CreateWorkerExampleKeyValueStore as a template
func CreateWorkerExampleKeyValueStore(workerGUID string) IWorker {
	worker := new(WorkerExampleKeyValueStore)
	worker.GUID = workerGUID
	worker.Group = ""
	worker.CreatedAt = time.Now()
	worker.LastDidWork = time.Now()
	worker.MessageList = make([]MapMessage, 0)
	worker.DataStored = make(map[string][]byte)
	return worker
}

// ProcessMessages process all messages for this worker
func (base *WorkerExampleKeyValueStore) ProcessMessages() {

	for len(base.MessageList) > 0 {
		Message := base.MessageList[0]
		base.MessageList = base.MessageList[1:]

		// Process Message
		switch Message.GetValue("command") {
		case "get":
			data, ok := base.DataStored[Message.GetValue("key")]
			if ok {
				ReturnMessage := CreateMapMessageEmpty()
				ReturnMessage.DestinationGUID = Message.SourceGUID
				ReturnMessage.SourceGUID = base.GUID
				ReturnMessage.SetData(data)
				ReturnMessage.SetValue("command", "getReturn")
				ReturnMessage.SetValue("key", Message.GetValue("key"))
				base.Send(ReturnMessage)
			}
			break
		case "set":
			base.DataStored[Message.GetValue("key")] = Message.GetData()
			fmt.Printf("%s - Data stored Key:%s\n", base.GUID, Message.GetValue("key"))
			break
		case "delete":
			_, ok := base.DataStored[Message.GetValue("key")]
			if ok {
				delete(base.DataStored, Message.GetValue("key"))
			}
			break
		default:
			break
		}
	}
}

// DoWork do the work of the worker
func (base *WorkerExampleKeyValueStore) DoWork() {

	//	fmt.Println(base.GUID, " - WorkDone")
	time.Sleep(50 * time.Millisecond)
}

// GetGUID of the worker
func (base WorkerExampleKeyValueStore) GetGUID() string {
	return base.GUID
}

// GetGroup of the worker
func (base WorkerExampleKeyValueStore) GetGroup() string {
	return base.Group
}

// GetAge age of fhe link
func (base WorkerExampleKeyValueStore) GetAge() string {
	return strconv.FormatFloat(time.Now().Sub(base.CreatedAt).Seconds(), 'f', 0, 64)
}

func (base *WorkerExampleKeyValueStore) getDeltaTime() int64 {
	DeltaTime := time.Now().Sub(base.LastDidWork).Nanoseconds()
	base.LastDidWork = time.Now()
	return DeltaTime
}

// AttachSendMethod of the
func (base *WorkerExampleKeyValueStore) AttachSendMethod(parentsSendFunction func(MapMessage)) {
	base.Send = parentsSendFunction
}

// QueueMessage Queue messge for the worker
func (base *WorkerExampleKeyValueStore) QueueMessage(Message MapMessage) {
	base.MessageList = append(base.MessageList, Message)
}

// Close the Worker
func (base *WorkerExampleKeyValueStore) Close() {
	fmt.Println(base.GUID, " - Worker Closed")
}
