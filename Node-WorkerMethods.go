package AutoMPI

import "time"
import "fmt"

const (
	defautNumberOfWorkerLocationAnnouncements = 3
	secondsToSleepIfNoWorkers                 = 1
)

func (base *Node) workerWorkLoop() {
	for {
		if len(base.Workers) > 0 {
			for _, value := range base.Workers {
				value.DoWork()
			}
		} else {
			time.Sleep(secondsToSleepIfNoWorkers * time.Second)
		}
	}
}
func (base *Node) workerProcessMessagesLoop() {
	for {
		if len(base.Workers) > 0 {
			for _, value := range base.Workers {
				value.ProcessMessages()
			}
		} else {
			time.Sleep(secondsToSleepIfNoWorkers * time.Second)
		}
	}
}

// AttachWorker to the node
// AttachSendMethod to enable Sending of messages
// Add worker to worker processing loop
// Announce the location on the worker
func (base *Node) AttachWorker(worker IWorker) {
	if worker != nil {
		worker.AttachSendMethod(base.Send)
		base.Workers[worker.GetGUID()] = worker
		base.addLocalWorkerLocation(worker.GetGUID())
	}
}

// DetachWorker Close() the worker and remove it from the Node
func (base *Node) DetachWorker(workerGUID string) {
	_, ok := base.Workers[workerGUID]
	if ok {
		base.Workers[workerGUID].Close()
		delete(base.Workers, workerGUID)
		base.removeLocalWorkerLocation(workerGUID)
		fmt.Printf("Worker %s stoped\n", workerGUID)
	} else {
		fmt.Printf("Stop worker command received but such worker found: %s \n", workerGUID)
	}
}

func (base *Node) addLocalWorkerLocation(WorkerGUID string) {
	base.LocalWorkersLocation[WorkerGUID] = CreateWorkerLocation(WorkerGUID, base.MyNodeGUID)
	base.LocalWorkersAnnouncingLocation[WorkerGUID] = defautNumberOfWorkerLocationAnnouncements
}
func (base *Node) removeLocalWorkerLocation(WorkerGUID string) {

	_, ok := base.LocalWorkersLocation[WorkerGUID]
	if ok {
		delete(base.LocalWorkersLocation, WorkerGUID)
	}
}
func (base *Node) isALocalWorker(WorkerGUID string) bool {
	_, ok := base.LocalWorkersLocation[WorkerGUID]
	return ok
}

func (base *Node) processWorkerLocation(Message MapMessage) {
	base.AllWorkersLocation[Message.GetValue(SystemMessageDataPartGUIDWorker)] = CreateWorkerLocation(Message.GetValue(SystemMessageDataPartGUIDWorker), Message.GetValue(SystemMessageDataPartGUIDNode))
}

func (base *Node) announceAllAgentsInLocalAgentsAnnouncing() {

	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.MyNodeGUID

	for key, value := range base.LocalWorkersAnnouncingLocation {
		if value > 0 {
			TemplateMessage[SystemMessageDataPartGUIDWorker] = key
			base.BoardCaster.Boardcast(CreateMapMessage(TemplateMessage))
			base.LocalWorkersAnnouncingLocation[key]--
		} else {
			delete(base.LocalWorkersAnnouncingLocation, key)
		}

	}
}

func (base *Node) announceAllAgentsInLocalAgentsOnce() {
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.MyNodeGUID

	for key := range base.LocalWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.BoardCaster.Boardcast(CreateMapMessage(TemplateMessage))
	}
}

func (base *Node) getHostingNodeOfWorkerBySearchingCollections(WorkerGUID string) (NodeGUID string, IsKnown bool) {

	workerLocation, ok := base.AllWorkersLocation[WorkerGUID]
	if ok {
		return workerLocation.parentNodeGUID, true
	}
	return "", false

}
