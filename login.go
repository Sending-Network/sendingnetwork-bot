package sdnclient

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethereumcrypto "github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
)

type DIDListResponse struct {
	Data []string `json:"data"`
}

type CreateDIDResponse struct {
	DID     string `json:"did"`
	Message string `json:"message"`
	Updated string `json:"updated"`
}

type PreLoginResponse struct {
	DID     string `json:"did"`
	Message string `json:"message"`
	Updated string `json:"updated"`
}

type LoginIdentifier struct {
	DID   string `json:"did"`
	Token string `json:"token"`
}

type DIDLoginRequest struct {
	Type       string          `json:"type"`
	Updated    string          `json:"updated"`
	Identifier LoginIdentifier `json:"identifier"`
	DeviceId   string          `json:"device_id"`
}

type DIDLoginResponse struct {
	AccessToken string `json:"access_token"`
	UserId      string `json:"user_id"`
	DeviceId    string `json:"device_id"`
}

func GetDIDList(ctx context.Context, hostname, address string) ([]string, error) {
	resByte, err := sendRequest("GET",
		fmt.Sprintf("%v/_api/client/unstable/address/%v", hostname, address), "", nil)
	if err != nil {
		log.Errorf("room-state-test-case GetDIDList fail. err:%v", err.Error())
		return nil, err
	}

	res := DIDListResponse{}
	if err := json.Unmarshal(resByte, &res); err != nil {
		log.Errorf("room-state-test-case GetDIDList fail. err:%v", err.Error())
		return nil, err
	}

	return res.Data, nil
}

func CreateDID(ctx context.Context, hostname, address string) (*CreateDIDResponse, error) {
	req := struct {
		Address string `json:"address"`
	}{
		Address: address,
	}
	body, _ := json.Marshal(req)
	resByte, err := sendRequest("POST", fmt.Sprintf("%v/_api/client/unstable/did/create", hostname), "", body)
	if err != nil {
		log.Errorf("room-state-test-case CreateDID fail. err:%v", err.Error())
		return nil, err
	}

	res := &CreateDIDResponse{}
	if err := json.Unmarshal(resByte, res); err != nil {
		log.Errorf("room-state-test-case CreateDID fail. err:%v", err.Error())
		return nil, err
	}

	return res, nil
}

func SaveDID(ctx context.Context, hostname, did, signature, operation, address, updated string) error {
	req := struct {
		Signature string `json:"signature"`
		Operation string `json:"operation"`
		Address   string `json:"address"`
		Updated   string `json:"updated"`
	}{
		Signature: signature,
		Operation: operation,
		Address:   address,
		Updated:   updated,
	}
	body, _ := json.Marshal(req)
	_, err := sendRequest("POST", fmt.Sprintf("%v/_api/client/unstable/did/%v", hostname, did), "", body)
	if err != nil {
		log.Errorf("room-state-test-case SaveDID fail. err:%v", err.Error())
		return err
	}

	return nil
}

func PreLogin(ctx context.Context, hostname, did string) (*PreLoginResponse, error) {

	body, _ := json.Marshal(struct{}{})
	resByte, err := sendRequest("POST", fmt.Sprintf("%v/_api/client/unstable/did/%v/pre_login", hostname, did), "", body)
	if err != nil {
		log.Errorf("room-state-test-case PreLogin fail. err:%v", err.Error())
		return nil, err
	}

	res := &PreLoginResponse{}
	if err := json.Unmarshal(resByte, res); err != nil {
		log.Errorf("room-state-test-case PreLogin fail. err:%v", err.Error())
		return nil, err
	}

	return res, nil
}

func DIDLogin(ctx context.Context, hostname, loginType, updated, did, token, deviceId string) (*DIDLoginResponse, error) {

	req := DIDLoginRequest{
		Type:    loginType,
		Updated: updated,
		Identifier: LoginIdentifier{
			DID:   did,
			Token: token,
		},
		DeviceId: deviceId,
	}
	body, _ := json.Marshal(req)
	resByte, err := sendRequest("POST", fmt.Sprintf("%v/_api/client/unstable/did/login", hostname), "", body)
	if err != nil {
		log.Errorf("room-state-test-case DIDLogin fail. err:%v", err.Error())
		return nil, err
	}

	res := &DIDLoginResponse{}
	if err := json.Unmarshal(resByte, res); err != nil {
		log.Errorf("room-state-test-case DIDLogin fail. err:%v", err.Error())
		return nil, err
	}

	return res, nil
}

// Login client login
func Login(endpoint, address string, privateKey *ecdsa.PrivateKey) (accessToken, userID string, err error) {
	ctx := context.Background()
	didList, err := GetDIDList(ctx, endpoint, address)
	if err != nil {
		return "", "", err
	}

	did := ""
	if len(didList) == 0 {
		createDIDResponse, err := CreateDID(ctx, endpoint, address)
		if err != nil {
			return "", "", err
		}

		signature, _ := ethereumcrypto.Sign(accounts.TextHash([]byte(createDIDResponse.Message)), privateKey)
		if err := SaveDID(
			ctx, endpoint, createDIDResponse.DID, hexutil.Encode(signature),
			"create", address, createDIDResponse.Updated); err != nil {
			return "", "", err
		}

		did = createDIDResponse.DID
	} else {
		did = didList[0]
	}

	preLoginResponse, err := PreLogin(ctx, endpoint, did)
	if err != nil {
		return "", "", err
	}

	signature, _ := ethereumcrypto.Sign(accounts.TextHash([]byte(preLoginResponse.Message)), privateKey)
	didLoginResponse, err := DIDLogin(ctx, endpoint, "m.login.did.identity", preLoginResponse.Updated,
		did, hexutil.Encode(signature), "")
	if err != nil {
		return "", "", err
	}

	return didLoginResponse.AccessToken, didLoginResponse.UserId, nil
}

func sendRequest(method, url, accessToken string, content []byte) ([]byte, error) {
	var body io.Reader
	if content != nil {
		body = bytes.NewBuffer(content)
	}
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}
