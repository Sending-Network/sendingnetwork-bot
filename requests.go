package sdnclient

// ReqCreateRoom is the JSON request for create room
type ReqCreateRoom struct {
	Name            string                 `json:"name,omitempty"`
	RoomAliasName   string                 `json:"room_alias_name,omitempty"`
	Topic           string                 `json:"topic,omitempty"`
	Invite          []string               `json:"invite,omitempty"`
	CreationContent map[string]interface{} `json:"creation_content,omitempty"`
	InitialState    []Event                `json:"initial_state,omitempty"`
	Preset          string                 `json:"preset,omitempty"`
}

// ReqInvite3PID is the JSON request invite 3pid
type ReqInvite3PID struct {
	IDServer string `json:"id_server"`
	Medium   string `json:"medium"`
	Address  string `json:"address"`
}

// ReqInviteUser is the JSON request for invite user
type ReqInviteUser struct {
	UserID string `json:"user_id"`
}

// ReqKickUser is the JSON request for kick user
type ReqKickUser struct {
	Reason string `json:"reason,omitempty"`
	UserID string `json:"user_id"`
}
