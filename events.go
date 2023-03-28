package sdnclient

// Event represents a single SDN event
type Event struct {
	StateKey    *string                `json:"state_key,omitempty"`    // The state key for the event. Only present on State Events.
	Sender      string                 `json:"sender"`                 // The user ID of the sender of the event
	Type        string                 `json:"type"`                   // The event type
	Timestamp   int64                  `json:"origin_server_ts"`       // The unix timestamp when this message was sent by the origin server
	ID          string                 `json:"event_id"`               // The unique ID of this event
	RoomID      string                 `json:"room_id"`                // The room the event was sent to. May be nil (e.g. for presence)
	Redacts     string                 `json:"redacts,omitempty"`      // The event ID that was redacted if a m.room.redaction event
	Unsigned    map[string]interface{} `json:"unsigned"`               // The unsigned portions of the event, such as age and prev_content
	Content     map[string]interface{} `json:"content"`                // The JSON content of the event.
	PrevContent map[string]interface{} `json:"prev_content,omitempty"` // The JSON prev_content of the event.
}

// Body returns the value of the "body" key in the event content if it is
// present and is a string.
func (event *Event) Body() (body string, ok bool) {
	value, exists := event.Content["body"]
	if !exists {
		return
	}
	body, ok = value.(string)
	return
}

// MessageType returns the value of the "msgtype" key in the event content if
// it is present and is a string.
func (event *Event) MessageType() (msgtype string, ok bool) {
	value, exists := event.Content["msgtype"]
	if !exists {
		return
	}
	msgtype, ok = value.(string)
	return
}
