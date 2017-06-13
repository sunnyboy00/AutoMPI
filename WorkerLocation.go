package AutoMPI

import (
	"time"
)

// WorkerLocation details
type WorkerLocation struct {
	GUID              string
	Group             string
	CreationTimeStamp time.Time
	ModifiedTimeStamp time.Time
	parentNodeGUID    string
}

// CreateWorkerLocation as it says
func CreateWorkerLocation(GUID string, Group string, parentNodeGUID string) *WorkerLocation {
	location := new(WorkerLocation)
	location.GUID = GUID
	location.Group = Group
	location.CreationTimeStamp = time.Now()
	location.ModifiedTimeStamp = time.Now()
	location.parentNodeGUID = parentNodeGUID
	return location
}
