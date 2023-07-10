package sdnclient

import (
	"crypto/ecdsa"
	"encoding/json"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"sync"
	"time"

	ethereumcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

// Config represents a configuration for client
type Config struct {
	Endpoint      string `yaml:"endpoint"`
	WalletAddress string `yaml:"wallet_address"`
	PrivateKey    string `yaml:"private_key"`
	UserID        string `yaml:"user_id"`
	AccessToken   string `yaml:"access_token"`
}

// Client represents a SDN client
type Client struct {
	UserID        string
	AccessToken   string
	Endpoint      string
	WalletAddress string
	PrivateKey    *ecdsa.PrivateKey
	httpClient    *http.Client
	PathPrefix    string
	Syncer        Syncer
	Store         Storer
	syncingMutex  sync.Mutex // protects syncingID
	syncingID     uint32     // Identifies the current Sync. Only one Sync can be active at any given time.
}

// NewClient create a new SDN client with the given configuration
func NewClient(config *Config) (*Client, error) {
	privateKey, err := ethereumcrypto.HexToECDSA(config.PrivateKey)
	if err != nil {
		return nil, err
	}
	if len(config.AccessToken) == 0 || len(config.UserID) == 0 {
		accessToken, userID, err := Login(config.Endpoint, config.WalletAddress, privateKey)
		if err != nil {
			return nil, err
		}
		log.Infof("login userID: %s, accessToken: %s", userID, accessToken)
		config.AccessToken = accessToken
		config.UserID = userID
	}
	store := NewInMemoryStore()
	return &Client{
		UserID:        config.UserID,
		AccessToken:   config.AccessToken,
		Endpoint:      config.Endpoint,
		WalletAddress: config.WalletAddress,
		PrivateKey:    privateKey,
		httpClient:    http.DefaultClient,
		PathPrefix:    "/_api/client/r0",
		Syncer:        NewDefaultSyncer(config.UserID, store),
		Store:         store,
	}, nil
}

// BuildURL builds a URL to send request to
func (cli *Client) BuildURL(urlPath ...string) string {
	return cli.Endpoint + path.Join(cli.PathPrefix, path.Join(urlPath...))
}

// BuildURLWithQuery builds a URL with query parameters
func (cli *Client) BuildURLWithQuery(urlPath []string, urlQuery map[string]string) string {
	u, _ := url.Parse(cli.BuildURL(urlPath...))
	q := u.Query()
	for k, v := range urlQuery {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// GetJoinedRooms GetJoinedRooms
func (cli *Client) GetJoinedRooms() (resp *RespJoinedRooms, err error) {
	urlPath := cli.BuildURL("joined_rooms")
	err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

// CreateRoom creates a new SDN room
func (cli *Client) CreateRoom(req *ReqCreateRoom) (resp *RespCreateRoom, err error) {
	urlPath := cli.BuildURL("createRoom")
	err = cli.MakeRequest("POST", urlPath, req, &resp)
	return
}

// JoinRoom joins the client to a room ID or alias
func (cli *Client) JoinRoom(roomIDorAlias string) (resp *RespJoinRoom, err error) {
	u := cli.BuildURL("join", roomIDorAlias)
	err = cli.MakeRequest("POST", u, struct{}{}, &resp)
	return
}

// LeaveRoom leaves the given room
func (cli *Client) LeaveRoom(roomID string) (resp *RespLeaveRoom, err error) {
	u := cli.BuildURL("rooms", roomID, "leave")
	err = cli.MakeRequest("POST", u, struct{}{}, &resp)
	return
}

// InviteUser invites a user to a room
func (cli *Client) InviteUser(roomID string, req *ReqInviteUser) (resp *RespInviteUser, err error) {
	u := cli.BuildURL("rooms", roomID, "invite")
	err = cli.MakeRequest("POST", u, req, &resp)
	return
}

// KickUser kicks a user from a room
func (cli *Client) KickUser(roomID string, req *ReqKickUser) (resp *RespKickUser, err error) {
	u := cli.BuildURL("rooms", roomID, "kick")
	err = cli.MakeRequest("POST", u, req, &resp)
	return
}

// JoinedMembers returns a map of joined room members
func (cli *Client) JoinedMembers(roomID string) (resp *RespJoinedMembers, err error) {
	u := cli.BuildURL("rooms", roomID, "joined_members")
	err = cli.MakeRequest("GET", u, nil, &resp)
	return
}

// Logout the current user.
func (cli *Client) Logout() (resp *RespLogout, err error) {
	urlPath := cli.BuildURL("logout")
	err = cli.MakeRequest("POST", urlPath, nil, &resp)
	return
}

// GetDisplayName returns the client's display name.
func (cli *Client) GetDisplayName() (resp *RespUserDisplayName, err error) {
	urlPath := cli.BuildURL("profile", cli.UserID, "displayname")
	err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

// SetDisplayName sets the client's profile display name.
func (cli *Client) SetDisplayName(displayName string) (err error) {
	urlPath := cli.BuildURL("profile", cli.UserID, "displayname")
	s := struct {
		DisplayName string `json:"displayname"`
	}{displayName}
	err = cli.MakeRequest("PUT", urlPath, &s, nil)
	return
}

// GetAvatarURL gets the client's avatar URL.
func (cli *Client) GetAvatarURL() (avatarUrl string, err error) {
	urlPath := cli.BuildURL("profile", cli.UserID, "avatar_url")
	s := struct {
		AvatarURL string `json:"avatar_url"`
	}{}

	err = cli.MakeRequest("GET", urlPath, nil, &s)
	if err != nil {
		return "", err
	}
	return s.AvatarURL, nil
}

// SetAvatarURL sets the client's avatar URL.
func (cli *Client) SetAvatarURL(url string) (err error) {
	urlPath := cli.BuildURL("profile", cli.UserID, "avatar_url")
	s := struct {
		AvatarURL string `json:"avatar_url"`
	}{url}
	err = cli.MakeRequest("PUT", urlPath, &s, nil)
	if err != nil {
		return err
	}
	return nil
}

func txnID() string {
	return "go" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

// GetStateEvent get a state event from a room.
func (cli *Client) GetStateEvent(roomID, eventType, stateKey string) (resp map[string]interface{}, err error) {
	urlPath := cli.BuildURL("rooms", roomID, "state", eventType, stateKey)
	err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

// SendStateEvent sends a state event into a room.
// contentJSON should be a pointer to something that can be encoded as JSON using json.Marshal.
func (cli *Client) SendStateEvent(roomID, eventType, stateKey string, contentJSON interface{}) (resp *RespSendEvent, err error) {
	urlPath := cli.BuildURL("rooms", roomID, "state", eventType, stateKey)
	err = cli.MakeRequest("PUT", urlPath, contentJSON, &resp)
	return
}

// SendMessageEvent sends a message event into a room.
// contentJSON should be a pointer to something that can be encoded as JSON using json.Marshal.
func (cli *Client) SendMessageEvent(roomID string, eventType string, contentJSON interface{}) (resp *RespSendEvent, err error) {
	txnID := txnID()
	urlPath := cli.BuildURL("rooms", roomID, "send", eventType, txnID)
	err = cli.MakeRequest("PUT", urlPath, contentJSON, &resp)
	return
}

// SendText sends an m.room.message event into the given room with a msgtype of m.text
func (cli *Client) SendText(roomID, text string) (*RespSendEvent, error) {
	return cli.SendMessageEvent(roomID, "m.room.message",
		TextMessage{MsgType: "m.text", Body: text})
}

// SendFormattedText sends an m.room.message event into the given room with a msgtype of m.text, supports a subset of HTML for formatting.
func (cli *Client) SendFormattedText(roomID, text, formattedText string) (*RespSendEvent, error) {
	return cli.SendMessageEvent(roomID, "m.room.message",
		TextMessage{MsgType: "m.text", Body: text, FormattedBody: formattedText, Format: "org.sdn.custom.html"})
}

// SendImage sends an m.room.message event into the given room with a msgtype of m.image
func (cli *Client) SendImage(roomID, body, url string) (*RespSendEvent, error) {
	return cli.SendMessageEvent(roomID, "m.room.message",
		ImageMessage{
			MsgType: "m.image",
			Body:    body,
			URL:     url,
		})
}

// SendVideo sends an m.room.message event into the given room with a msgtype of m.video
func (cli *Client) SendVideo(roomID, body, url string) (*RespSendEvent, error) {
	return cli.SendMessageEvent(roomID, "m.room.message",
		VideoMessage{
			MsgType: "m.video",
			Body:    body,
			URL:     url,
		})
}

// SendNotice sends an m.room.message event into the given room with a msgtype of m.notice
func (cli *Client) SendNotice(roomID, text string) (*RespSendEvent, error) {
	return cli.SendMessageEvent(roomID, "m.room.message",
		TextMessage{MsgType: "m.notice", Body: text})
}

// Sync starts syncing with the provided server. If Sync() is called twice then the first sync will be stopped and the
// error will be nil.
//
// This function will block until a fatal /sync error occurs, so it should almost always be started as a new goroutine.
// Fatal sync errors can be caused by:
//   - The failure to create a filter.
//   - Client.Syncer.OnFailedSync returning an error in response to a failed sync.
//   - Client.Syncer.ProcessResponse returning an error.
//
// If you wish to continue retrying in spite of these fatal errors, call Sync() again.
func (cli *Client) Sync() error {
	// Mark the client as syncing.
	// We will keep syncing until the syncing state changes. Either because
	// Sync is called or StopSync is called.
	syncingID := cli.incrementSyncingID()
	nextBatch := cli.Store.LoadNextBatch(cli.UserID)
	filterID := cli.Store.LoadFilterID(cli.UserID)
	if filterID == "" {
		filterJSON := cli.Syncer.GetFilterJSON(cli.UserID)
		resFilter, err := cli.CreateFilter(filterJSON)
		if err != nil {
			return err
		}
		filterID = resFilter.FilterID
		cli.Store.SaveFilterID(cli.UserID, filterID)
	}

	for {
		log.Infof("syncing with %s", nextBatch)
		resSync, err := cli.SyncRequest(30000, nextBatch, filterID, false, "")
		if err != nil {
			duration, err2 := cli.Syncer.OnFailedSync(resSync, err)
			if err2 != nil {
				return err2
			}
			time.Sleep(duration)
			continue
		}

		// Check that the syncing state hasn't changed
		// Either because we've stopped syncing or another sync has been started.
		// We discard the response from our sync.
		if cli.getSyncingID() != syncingID {
			return nil
		}

		// Save the token now *before* processing it. This means it's possible
		// to not process some events, but it means that we won't get constantly stuck processing
		// a malformed/buggy event which keeps making us panic.
		cli.Store.SaveNextBatch(cli.UserID, resSync.NextBatch)
		if err = cli.Syncer.ProcessResponse(resSync, nextBatch); err != nil {
			return err
		}

		nextBatch = resSync.NextBatch
	}
}

func (cli *Client) incrementSyncingID() uint32 {
	cli.syncingMutex.Lock()
	defer cli.syncingMutex.Unlock()
	cli.syncingID++
	return cli.syncingID
}

func (cli *Client) getSyncingID() uint32 {
	cli.syncingMutex.Lock()
	defer cli.syncingMutex.Unlock()
	return cli.syncingID
}

// StopSync stops the ongoing sync started by Sync.
func (cli *Client) StopSync() {
	// Advance the syncing state so that any running Syncs will terminate.
	cli.incrementSyncingID()
}

// SyncRequest makes an sync request
func (cli *Client) SyncRequest(timeout int, since, filterID string, fullState bool, setPresence string) (resp *RespSync, err error) {
	query := map[string]string{
		"timeout": strconv.Itoa(timeout),
	}
	if since != "" {
		query["since"] = since
	}
	if filterID != "" {
		query["filter"] = filterID
	}
	if setPresence != "" {
		query["set_presence"] = setPresence
	}
	if fullState {
		query["full_state"] = "true"
	}
	urlPath := cli.BuildURLWithQuery([]string{"sync"}, query)
	err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

// CreateFilter .
func (cli *Client) CreateFilter(filter json.RawMessage) (resp *RespCreateFilter, err error) {
	urlPath := cli.BuildURL("user", cli.UserID, "filter")
	err = cli.MakeRequest("POST", urlPath, &filter, &resp)
	return
}
