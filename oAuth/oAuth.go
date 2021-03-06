package bookoauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ankitanwar/GoAPIUtils/errors"

	"github.com/mercadolibre/golang-restclient/rest"
)

const (
	headerXPublic   = "X-Public"
	headerXClientID = "X-Client-Id"
	headerXCallerID = "X-Caller-Id"

	paramAccessToken = "access_token"
)

var (
	oauthRestClient = rest.RequestBuilder{
		BaseURL: "http://localhost:8080",
		Timeout: 200 * time.Millisecond,
	}
)

type accessToken struct {
	ID       string `json:"id"`
	UserID   int64  `json:"user_id"`
	ClientID int64  `json:"client_id"`
}

//IsPublic : To validate the request whether the request is public or not
func IsPublic(request *http.Request) bool {
	if request == nil {
		return true
	}
	return request.Header.Get(headerXPublic) == "true"
}

//GetCallerID : To get the caller id from the url
func GetCallerID(request *http.Request) int {
	if request == nil {
		return 0
	}
	callerID, err := strconv.Atoi(request.Header.Get(headerXCallerID))
	if err != nil {
		return 0
	}
	return callerID
}

//GetClientID : To get the client id
func GetClientID(request *http.Request) int {
	if request == nil {
		return 0
	}
	clientID, err := strconv.Atoi(request.Header.Get(headerXClientID))
	if err != nil {
		return 0
	}
	return clientID
}

//AuthenticateRequest : To authenticate the given request
func AuthenticateRequest(request *http.Request) *errors.RestError {
	if request == nil {
		return nil
	}

	cleanRequest(request)

	accessTokenID := strings.TrimSpace(request.URL.Query().Get(paramAccessToken))
	if accessTokenID == "" {
		return nil
	}

	at, err := getAccessToken(accessTokenID)
	if err != nil {
		if err.Status == http.StatusNotFound {
			return nil
		}
		return err
	}
	request.Header.Add(headerXClientID, fmt.Sprintf("%v", at.ClientID))
	request.Header.Add(headerXCallerID, fmt.Sprintf("%v", at.UserID))
	return nil
}

func cleanRequest(request *http.Request) {
	if request == nil {
		return
	}
	request.Header.Del(headerXClientID)
	request.Header.Del(headerXCallerID)
}

func getAccessToken(accessTokenID string) (*accessToken, *errors.RestError) {
	response := oauthRestClient.Get(fmt.Sprintf("/oauth/access_token/%s", accessTokenID))
	if response == nil || response.Response == nil {
		return nil, errors.NewNotFound("Not found")
	}

	if response.StatusCode > 299 {
		err := errors.NewInternalServerError("internal server err")
		return nil, err
	}

	var at accessToken
	if err := json.Unmarshal(response.Bytes(), &at); err != nil {
		return nil, errors.NewInternalServerError("error when trying to unmarshal access token response")
	}
	return &at, nil
}
