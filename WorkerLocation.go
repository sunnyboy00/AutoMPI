package AutoMPI

import (
	"time"
)

// WorkerLocation details
type WorkerLocation struct {
	GUID              string
	CreationTimeStamp time.Time
	ModifiedTimeStamp time.Time
	parentNodeGUID    string
}

// CreateWorkerLocation as it says
func CreateWorkerLocation(GUID string, parentNodeGUID string) *WorkerLocation {
	location := new(WorkerLocation)
	location.GUID = GUID
	location.CreationTimeStamp = time.Now()
	location.ModifiedTimeStamp = time.Now()
	location.parentNodeGUID = parentNodeGUID
	return location
}
