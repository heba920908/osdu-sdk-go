package osdu

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/heba920908/osdu-sdk-go/pkg/config"
)

type AuthToken struct {
	AccessToken    string `json:"access_token"`
	TokenType      string `json:"token_type"`
	RefreshToken   string `json:"refresh_token"`
	ExpiresIn      int    `json:"expires_in"`
	InternalExpire time.Time
}

func NewAuthToken() *AuthToken {
	auths, _ := config.GetAuthSettings()
	var a AuthToken
	aa, _ := a.GetAccessToken(auths)
	return aa
}

// Got example from https://github.com/mcordell/go-ms-graph/blob/master/auth/auth.go
func (a *AuthToken) GetAccessToken(authsettings config.AuthSettings) (*AuthToken, error) {
	ctx := context.WithValue(context.Background(), OsduApi, "GetAccessToken")
	slog.DebugContext(ctx, fmt.Sprintf("Auth: Expire - %s", a.InternalExpire))
	if len(a.AccessToken) > 5 {
		now := time.Now()
		if now.Before(a.InternalExpire) {
			slog.DebugContext(ctx, fmt.Sprintf("Token still active ... %s > %s", a.InternalExpire, now))
			return a, nil
		}
	}

	slog.InfoContext(ctx, "Auth - Generating new token")

	formVals := url.Values{}
	formVals.Set("client_id", authsettings.ClientId)
	formVals.Set("grant_type", authsettings.GrantType)
	if authsettings.GrantType == "refresh_token" {
		formVals.Set("refresh_token", authsettings.RefreshToken)
		if a.RefreshToken != "" {
			formVals.Set("refresh_token", a.RefreshToken)
		}
	}
	formVals.Set("scope", authsettings.Scopes)
	if len(authsettings.ClientSecret) > 0 {
		formVals.Set("client_secret", authsettings.ClientSecret)
	}
	slog.InfoContext(ctx, fmt.Sprintf("Trying: %s", authsettings.TokenUrl))
	slog.InfoContext(ctx, fmt.Sprintf("grant_type: %s", authsettings.GrantType))
	slog.InfoContext(ctx, fmt.Sprintf("client_id: %s", authsettings.ClientId))
	slog.InfoContext(ctx, fmt.Sprintf("scope: %s", authsettings.Scopes))

	response, err := http.PostForm(authsettings.TokenUrl, formVals)

	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error while obtaining token: %s", err))
		return a, err
	}

	if response.StatusCode > 302 {
		slog.ErrorContext(ctx, fmt.Sprintf("Unexpected auth code: %d", response.StatusCode))
		return a, err
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return a, err
	}

	var aa AuthToken
	err = json.Unmarshal(body, &aa)
	aa.InternalExpire = time.Now().Add(60 * time.Minute)
	*a = aa
	if err != nil {
		return a, err
	}
	slog.InfoContext(ctx, fmt.Sprintf("Auth - Done - code: %s | AT: %d", response.Status, len(a.AccessToken)))
	return a, nil
}
