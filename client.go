package sdnclient

import (
	"crypto/ecdsa"
	"net/http"
	"path"

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
	return &Client{
		UserID:        config.UserID,
		AccessToken:   config.AccessToken,
		Endpoint:      config.Endpoint,
		WalletAddress: config.WalletAddress,
		PrivateKey:    privateKey,
		httpClient:    http.DefaultClient,
		PathPrefix:    "/_api/client/r0",
	}, nil
}

// BuildURL builds a URL to send request to
func (cli *Client) BuildURL(urlPath ...string) string {
	return cli.Endpoint + path.Join(cli.PathPrefix, path.Join(urlPath...))
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

// SendStateEvent sends a state event into a room
func (cli *Client) SendStateEvent(roomID, eventType, stateKey string, contentJSON interface{}) (resp *RespSendEvent, err error) {
	urlPath := cli.BuildURL("rooms", roomID, "state", eventType, stateKey)
	err = cli.MakeRequest("PUT", urlPath, contentJSON, &resp)
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
