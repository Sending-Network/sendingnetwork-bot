package sdnclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPError An HTTP Error response, which may wrap an underlying native Go Error.
type HTTPError struct {
	Contents     []byte
	WrappedError error
	Message      string
	Code         int
}

func (e HTTPError) Error() string {
	var wrappedErrMsg string
	if e.WrappedError != nil {
		wrappedErrMsg = e.WrappedError.Error()
	}
	return fmt.Sprintf("contents=%v msg=%s code=%d wrapped=%s", e.Contents, e.Message, e.Code, wrappedErrMsg)
}

// MakeRequest makes a JSON HTTP request to the given URL
func (cli *Client) MakeRequest(method string, httpURL string, reqBody interface{}, resBody interface{}) error {
	var req *http.Request
	var err error
	if reqBody != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return err
		}
		req, err = http.NewRequest(method, httpURL, buf)
	} else {
		req, err = http.NewRequest(method, httpURL, nil)
	}

	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	if cli.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+cli.AccessToken)
	}

	res, err := cli.httpClient.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK { // not 2xx
		contents, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var wrap error
		var respErr RespError
		if _ = json.Unmarshal(contents, &respErr); respErr.ErrCode != "" {
			wrap = respErr
		}

		// If we failed to decode as RespError, don't just drop the HTTP body, include it in the
		// HTTP error instead (e.g proxy errors which return HTML).
		msg := "Failed to " + method + " JSON to " + req.URL.Path
		if wrap == nil {
			msg = msg + ": " + string(contents)
		}

		return HTTPError{
			Contents:     contents,
			Code:         res.StatusCode,
			Message:      msg,
			WrappedError: wrap,
		}
	}

	if resBody != nil && res.Body != nil {
		return json.NewDecoder(res.Body).Decode(&resBody)
	}

	return nil
}
