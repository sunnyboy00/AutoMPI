package AutoMPI

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	// Import might be needed for saving images
	_ "image/jpeg"
)

// WorkerExampleIPCamera example worker
type WorkerExampleIPCamera struct {
	GUID        string
	CreatedAt   time.Time
	MessageList list.List
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
	return worker
}

// DoWork do the work of the worker
func (base *WorkerExampleIPCamera) DoWork() {

	for base.MessageList.Len() > 0 {
		Message := base.MessageList.Front()
		base.MessageList.Remove(Message)

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
	/*
		reader := bytes.NewReader(ImageData)

			var Texture image.Image
			Texture, _, err = image.Decode(reader)
			if err != nil {
				log.Fatal(err)
			}
	*/

	//Texture.

	imgFile, _ := os.Create("image.jpg")
	imgFile.Write(ImageData)
	imgFile.Close()

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
	base.MessageList.PushBack(Message)
}

// Close the Worker
func (base *WorkerExampleIPCamera) Close() {
	fmt.Println(base.GUID, " - Worker Closed")
}
