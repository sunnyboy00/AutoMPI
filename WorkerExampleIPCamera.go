package AutoMPI

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	// Import might be needed for saving images
	_ "image/jpeg"
)

// WorkerExampleIPCamera example worker
type WorkerExampleIPCamera struct {
	GUID        string
	CreatedAt   time.Time
	MessageList []MapMessage
	Send        func(MapMessage)
	LastDidWork time.Time
	sourceURL   string
}

// CreateWorkerExampleIPCamera as a template
func CreateWorkerExampleIPCamera(workerGUID string, sourceURL string) IWorker {
	worker := new(WorkerExampleIPCamera)
	worker.GUID = workerGUID
	worker.sourceURL = sourceURL
	worker.CreatedAt = time.Now()
	worker.LastDidWork = time.Now()
	worker.MessageList = make([]MapMessage, 0)
	return worker
}

// DoWork do the work of the worker
func (base *WorkerExampleIPCamera) DoWork() {

	for len(base.MessageList) > 0 {
		Message := base.MessageList[0]
		base.MessageList = base.MessageList[1:]

		Message.DestinationGUID = ""

		// Process Message
	}

	res, err := http.Get(base.sourceURL)
	if err != nil {
		log.Fatal(err)
	}
	ImageData, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	// write the image to a file
	/*
		imgFile, _ := os.Create("image.jpg")
		imgFile.Write(ImageData)
		imgFile.Close()
	*/

	// send the image to a store

	ReturnMessage := CreateMapMessageEmpty()
	ReturnMessage.DestinationGUID = "Store0001"
	ReturnMessage.SourceGUID = base.GUID
	ReturnMessage.SetData(ImageData)
	ReturnMessage.SetValue("command", "set")
	ReturnMessage.SetValue("key", "image-"+time.Now().String())
	base.Send(ReturnMessage)

	time.Sleep(2 * time.Second)
}

// GetGUID of the worker
func (base WorkerExampleIPCamera) GetGUID() string {
	return base.GUID
}

// GetAge age of fhe link
func (base WorkerExampleIPCamera) GetAge() string {
	return strconv.FormatFloat(time.Now().Sub(base.CreatedAt).Seconds(), 'f', 0, 64)
}

// AttachSendMethod of the
func (base *WorkerExampleIPCamera) AttachSendMethod(parentsSendFunction func(MapMessage)) {
	base.Send = parentsSendFunction
}

// QueueMessage Queue messge for the worker
func (base *WorkerExampleIPCamera) QueueMessage(Message MapMessage) {
	base.MessageList = append(base.MessageList, Message)
}

// Close the Worker
func (base *WorkerExampleIPCamera) Close() {
	fmt.Println(base.GUID, " - Worker Closed")
}
