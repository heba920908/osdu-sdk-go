package osdu

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

type osdu_api string

const (
	OsduApi osdu_api = "osdu_api"
)

/*
TODO: In favor or using the keycloak auth instead of own implementation
*/
type OsduApiRequest struct {
	AuthToken    AuthToken
	OsduSettings config.OsduSettings
	AuthSettings config.AuthSettings
}

func NewClient() OsduApiRequest {
	osdusettings, _ := config.GetOsduSettings()
	authsettings, _ := config.GetAuthSettings()
	authtoken := NewAuthToken()
	return OsduApiRequest{
		AuthToken:    *authtoken,
		OsduSettings: osdusettings,
		AuthSettings: authsettings,
	}
}

func (a OsduApiRequest) Context() context.Context {
	/* if r.ctx != nil {
		return r.ctx
	} */
	return context.Background()
}

func (a OsduApiRequest) NewRequest(operation string, url string, partitionid string, body []byte) ([]byte, error) {
	/*
		log.Println("Request:")
		_print_pretty_json(body)
	*/
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req, _ := http.NewRequest(operation, url, bytes.NewBuffer(body))
	headers, _ := a._build_headers_with_partition()
	req.Header = headers
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return resBody, err
	}
	res.Body.Close()
	slog.Debug("Response:")
	// utils.PrintPrettyJson(resBody, 30)
	return resBody, nil
}

func (a OsduApiRequest) _build_headers_with_partition() (http.Header, error) {
	atoken, err := a.AuthToken.GetAccessToken(a.AuthSettings)
	if err != nil {
		return http.Header{}, err
	}
	slog.Debug(fmt.Sprintf("Authorization Header : Bearer %s", atoken.AccessToken))
	return http.Header{
		"Content-Type":      {"application/json"},
		"Authorization":     {fmt.Sprintf("Bearer %s", atoken.AccessToken)},
		"data-partition-id": {a.OsduSettings.PartitionId},
	}, nil
}

func (a OsduApiRequest) _build_headers_without_partition() (http.Header, error) {
	if a.AuthSettings.InternalService {
		ctx := context.Background()
		slog.InfoContext(ctx, "Internal service setup, skipping token generation")
		return http.Header{
			"Content-Type": {"application/json"},
		}, nil
	}
	atoken, err := a.AuthToken.GetAccessToken(a.AuthSettings)
	if err != nil {
		log.Println(err)
	}
	return http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {fmt.Sprintf("Bearer %s", atoken.AccessToken)},
	}, nil
}
