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
			base.workerLock.RLock()
			for _, value := range base.Workers {
				value.DoWork()
			}
			base.workerLock.RUnlock()
		} else {
			time.Sleep(secondsToSleepIfNoWorkers * time.Second)
		}
	}
}
func (base *Node) workerProcessMessagesLoop() {
	for {
		if len(base.Workers) > 0 {
			base.workerLock.RLock()
			for _, value := range base.Workers {
				value.ProcessMessages()
			}
			base.workerLock.RUnlock()
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
		worker.AttachParentNode(base)
		base.workerLock.Lock()
		base.Workers[worker.GetGUID()] = worker
		base.workerLock.Unlock()
		base.addLocalWorkerLocation(worker.GetGUID())
	}
}

// DetachWorker Close() the worker and remove it from the Node
func (base *Node) DetachWorker(workerGUID string) {
	_, ok := base.Workers[workerGUID]
	if ok {
		base.workerLock.Lock()
		base.Workers[workerGUID].Close()
		delete(base.Workers, workerGUID)
		base.workerLock.Unlock()
		base.removeLocalWorkerLocation(workerGUID)
		fmt.Printf("Worker %s stoped\n", workerGUID)
	} else {
		fmt.Printf("Stop worker command received but such worker found: %s \n", workerGUID)
	}
}

func (base *Node) addLocalWorkerLocation(WorkerGUID string) {
	base.localWorkersLocation[WorkerGUID] = CreateWorkerLocation(WorkerGUID, base.GUID)
	base.localWorkersAnnouncingLocation[WorkerGUID] = defautNumberOfWorkerLocationAnnouncements
}
func (base *Node) removeLocalWorkerLocation(WorkerGUID string) {

	_, ok := base.localWorkersLocation[WorkerGUID]
	if ok {
		delete(base.localWorkersLocation, WorkerGUID)
	}
}
func (base *Node) isALocalWorker(WorkerGUID string) bool {
	base.workerLock.RLock()
	_, ok := base.localWorkersLocation[WorkerGUID]
	base.workerLock.RUnlock()
	return ok
}

func (base *Node) processWorkerLocation(Message MapMessage) {
	base.allWorkersLocationLock.Lock()
	base.allWorkersLocation[Message.GetValue(SystemMessageDataPartGUIDWorker)] = CreateWorkerLocation(Message.GetValue(SystemMessageDataPartGUIDWorker), Message.GetValue(SystemMessageDataPartGUIDNode))
	base.allWorkersLocationLock.Unlock()
}

func (base *Node) announceAllAgentsInLocalAgentsAnnouncing() {

	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.GUID

	for key, value := range base.localWorkersAnnouncingLocation {
		if value > 0 {
			TemplateMessage[SystemMessageDataPartGUIDWorker] = key
			base.boardCaster.Boardcast(CreateMapMessage(TemplateMessage))
			base.localWorkersAnnouncingLocation[key]--
		} else {
			delete(base.localWorkersAnnouncingLocation, key)
		}

	}
}

func (base *Node) announceAllAgentsInLocalAgentsOnce() {
	TemplateMessage := make(map[string]string)
	TemplateMessage[SystemKeysAutoMPISystemMessage] = SystemKeyDetailsWorkerLocation
	TemplateMessage[SystemMessageDataPartGUIDNode] = base.GUID

	for key := range base.localWorkersLocation {
		TemplateMessage[SystemMessageDataPartGUIDWorker] = key
		base.boardCaster.Boardcast(CreateMapMessage(TemplateMessage))
	}
}

func (base *Node) getHostingNodeOfWorkerBySearchingCollections(WorkerGUID string) (NodeGUID string, IsKnown bool) {

	base.allWorkersLocationLock.RLock()
	workerLocation, ok := base.allWorkersLocation[WorkerGUID]
	base.allWorkersLocationLock.RUnlock()
	if ok {
		return workerLocation.parentNodeGUID, true
	}
	return "", false

}

// IsThisAKnownWorkerBySearchingCollections as the method says
func (base *Node) IsThisAKnownWorkerBySearchingCollections(WorkerGUID string) bool {
	base.allWorkersLocationLock.RLock()
	_, IsKnown := base.allWorkersLocation[WorkerGUID]
	base.allWorkersLocationLock.RUnlock()
	return IsKnown

}
