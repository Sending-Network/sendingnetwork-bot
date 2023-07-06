package sdnclient

// RespError is the standard JSON error response from SDN node servers
type RespError struct {
	ErrCode string `json:"errcode"`
	Err     string `json:"error"`
}

// Error returns the errcode and error message.
func (e RespError) Error() string {
	return e.ErrCode + ": " + e.Err
}

// RespJoinRoom is the JSON response for JoinRoom
type RespJoinRoom struct {
	RoomID string `json:"room_id"`
}

// RespLeaveRoom is the JSON response for LeaveRoom
type RespLeaveRoom struct{}

// RespInviteUser is the JSON response for InviteUser
type RespInviteUser struct{}

// RespKickUser is the JSON response for KickUser
type RespKickUser struct{}

// RespJoinedRooms is the JSON response for JoinedRooms
type RespJoinedRooms struct {
	JoinedRooms []string `json:"joined_rooms"`
}

// RespJoinedMembers is the JSON response for JoinedMembers
type RespJoinedMembers struct {
	Joined map[string]struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	} `json:"joined"`
}

// RespSendEvent is the JSON response for SendEvent
type RespSendEvent struct {
	EventID string `json:"event_id"`
}

// RespLogout is the JSON response for Logout
type RespLogout struct{}

// RespCreateRoom is the JSON response for CreateRoom
type RespCreateRoom struct {
	RoomID string `json:"room_id"`
}

// RespUserDisplayName is the JSON response for GetDisplayName
type RespUserDisplayName struct {
	DisplayName string `json:"displayname"`
}

type RespCreateFilter struct {
	FilterID string `json:"filter_id"`
}

// RespSync is the JSON response for sync
type RespSync struct {
	NextBatch   string `json:"next_batch"`
	AccountData struct {
		Events []Event `json:"events"`
	} `json:"account_data"`
	Presence struct {
		Events []Event `json:"events"`
	} `json:"presence"`
	Rooms struct {
		Leave map[string]struct {
			State struct {
				Events []Event `json:"events"`
			} `json:"state"`
			Timeline struct {
				Events    []Event `json:"events"`
				Limited   bool    `json:"limited"`
				PrevBatch string  `json:"prev_batch"`
			} `json:"timeline"`
		} `json:"leave"`
		Join map[string]struct {
			State struct {
				Events []Event `json:"events"`
			} `json:"state"`
			Timeline struct {
				Events    []Event `json:"events"`
				Limited   bool    `json:"limited"`
				PrevBatch string  `json:"prev_batch"`
			} `json:"timeline"`
			Ephemeral struct {
				Events []Event `json:"events"`
			} `json:"ephemeral"`
		} `json:"join"`
		Invite map[string]struct {
			State struct {
				Events []Event
			} `json:"invite_state"`
		} `json:"invite"`
	} `json:"rooms"`
}
