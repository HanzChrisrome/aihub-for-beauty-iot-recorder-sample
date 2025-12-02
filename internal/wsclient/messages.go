package wsclient

type BaseMessage struct {
	Command string `json:"command"`
}

type StartRecordingMessage struct {
	Command     string `json:"command"`
	SessionID   string `json:"session_id"`
	DeviceIndex int    `json:"device_index"`
}

type StopRecordingMessage struct {
	Command   string `json:"command"`
	SessionID string `json:"session_id"`
}

type ListDevicesMessage struct {
	Command string `json:"command"`
}

type StopAllMessage struct {
	Command string `json:"command"`
}

type ResponseMessage struct {
	Command string      `json:"command"`
	Status  string      `json:"status"` // "success" or "error"
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}