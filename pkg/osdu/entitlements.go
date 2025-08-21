package osdu

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/heba920908/osdu-sdk-go/pkg/models"
)

type EntitlementsBootstrapRequest struct {
	AliasMappings []models.EntitlementsBoostrapUser `json:"aliasMappings"`
}

type EntitlementsAddUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (a OsduApiRequest) EntitlementsBootstrap() error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	bootstrap_url := fmt.Sprintf("%s/tenant-provisioning", a.osduSettings.EntitlementsUrl)
	boostrap_request := EntitlementsBootstrapRequest{
		AliasMappings: models.DefaultEntitlementsBootstrapUsers(),
	}

	json_content, err := json.Marshal(boostrap_request)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(boostrap_request, "", "  ")
	slog.Info(string(j))

	err = retry.Do(
		func() error {
			req, _ := http.NewRequest("POST", bootstrap_url, bytes.NewBuffer([]byte(json_content)))
			headers, _ := a._build_headers_with_partition()
			req.Header = headers

			http_client := http.Client{}

			res, err := http_client.Do(req)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}
			slog.InfoContext(ctx, fmt.Sprintf("Entitlements Boostrap Code: %d", res.StatusCode))
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
			slog.DebugContext(ctx, string(body))
			if res.StatusCode != http.StatusOK {
				return errors.New("not 200 response")
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(10*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.WarnContext(ctx, fmt.Sprintf("retry #%d: %s\n", n, err))
		}),
	)

	return err
}

func (a OsduApiRequest) CreateEntitlementsAdminUser(user_email string) error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	slog.InfoContext(ctx, fmt.Sprintf("[CreateEntitlementsAdminUser] User: %s",
		user_email))

	add_user_request := EntitlementsAddUserRequest{
		Role:  "OWNER",
		Email: user_email,
	}

	json_content, err := json.Marshal(add_user_request)
	if err != nil {
		return err
	}

	entitlement_groups := []string{"users", "users.datalake.ops", "users.datalake.admins"}

	j, _ := json.MarshalIndent(add_user_request, "", "  ")
	slog.InfoContext(ctx, fmt.Sprintf("[CreateEntitlementsAdminUser] Payload: %s", string(j)))

	headers, _ := a._build_headers_with_partition()

	err = retry.Do(
		func() error {
			for _, group := range entitlement_groups {
				entitlements_url := fmt.Sprintf("%s/groups/%s@%s.%s/members",
					a.osduSettings.EntitlementsUrl,
					group,
					a.osduSettings.PartitionId,
					a.osduSettings.EntitlementsDomain)
				req, _ := http.NewRequest("POST", entitlements_url, bytes.NewBuffer([]byte(json_content)))
				slog.InfoContext(ctx, fmt.Sprintf("[CreateEntitlementsAdminUser] POST: %s", entitlements_url))
				req.Header = headers

				http_client := http.Client{}

				res, err := http_client.Do(req)
				if err != nil {
					slog.Error(err.Error())
					return err
				}
				slog.InfoContext(ctx, fmt.Sprintf("[CreateEntitlementsAdminUser] User: %s | Group: %s | Code: %d",
					user_email,
					group,
					res.StatusCode))
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				if err != nil {
					slog.Error(err.Error())
				}
				slog.DebugContext(ctx, string(body))
				if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
					return errors.New("not 200 nor 409 response")
				}
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(10*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.WarnContext(ctx, fmt.Sprintf("retry #%d: %s\n", n, err))
		}),
	)

	return err
}
