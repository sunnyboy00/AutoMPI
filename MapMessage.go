package AutoMPI

import (
	"encoding/json"
	"fmt"
)

// MapMessage the struct for Messages with Map and []byte
type MapMessage struct {
	DestinationGUID  string
	DestinationGroup string
	SourceGUID       string
	Message          map[string]string
	Data             []byte
}

// CreateMapMessageEmpty create an empty MapMessage
func CreateMapMessageEmpty() MapMessage {
	Message := MapMessage{}
	Message.Message = map[string]string{}
	return Message
}

// CreateMapMessage create a MapMessage initilised with the rawMessage
func CreateMapMessage(rawMessage map[string]string) MapMessage {
	Message := MapMessage{}
	Message.SetMessage(rawMessage)
	return Message
}

// IsAutoMPISystemMessage quick check to if the message is a AutoMPI system message
func (base MapMessage) IsAutoMPISystemMessage() bool {
	if _, ok := base.Message[SystemKeysAutoMPISystemMessage]; ok {

		return true
	}
	return false
}

// SetDestination set tdestination of the message
func (base *MapMessage) SetDestination(GUID string) {
	base.DestinationGUID = GUID
}

// SetMessage sets the message map
func (base *MapMessage) SetMessage(message map[string]string) {
	base.Message = message
}

// SetValue set a particular value
func (base *MapMessage) SetValue(key string, value string) {
	base.Message[key] = value
}

// GetValue get a particular value
func (base MapMessage) GetValue(key string) string {
	if val, ok := base.Message[key]; ok {
		return val
	}
	return ""
}

// SetData Set data bytes
func (base *MapMessage) SetData(data []byte) {
	base.Data = data
}

// GetData get the data bytes
func (base MapMessage) GetData() []byte {
	return base.Data
}

// ToBytes get JSON blob from message
func (base MapMessage) ToBytes() []byte {
	b, err := json.Marshal(base)
	if err != nil {
		fmt.Println("error: : : :", err)
	}
	return b
}

// BlobToScreen Print Blob to screen
func (base MapMessage) BlobToScreen() {
	b, _ := json.Marshal(base)

	println(b)

}

// FromBytes get message from jsonBlob
func (base *MapMessage) FromBytes(jsonBlob []byte) {
	err := json.Unmarshal(jsonBlob, base)
	if err != nil {
		fmt.Println("error:", err)
	}
}

// MapMessageFromBytes get message from jsonBlob
func MapMessageFromBytes(jsonBlob []byte) MapMessage {
	Message := MapMessage{}
	err := json.Unmarshal(jsonBlob, Message)
	if err != nil {
		fmt.Println("error:", err)
	}
	return Message
}
