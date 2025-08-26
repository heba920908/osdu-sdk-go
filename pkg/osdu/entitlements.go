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

func (a OsduApiRequest) EntitlementsBootstrap() error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	bootstrap_url := fmt.Sprintf("%s/tenant-provisioning", a.osduSettings.EntitlementsUrl)
	boostrap_request := models.EntitlementsBootstrapRequest{
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
			headers, err := a._build_headers_with_partition()
			if err != nil {
				return err
			}

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

func (a OsduApiRequest) EntitlementsCreateAdminUser(user_email string) error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	slog.InfoContext(ctx, fmt.Sprintf("[CreateEntitlementsAdminUser] User: %s",
		user_email))

	add_user_request := models.EntitlementsAddUserRequest{
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

func (a OsduApiRequest) EntitlementsCreateGroup(group_id string, user_ids []string) error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	slog.InfoContext(ctx, fmt.Sprintf("Create Group %s", group_id))
	create_group_url := fmt.Sprintf("%s/groups", a.osduSettings.EntitlementsUrl)

	request_body := models.EntitlementsCreateGroupRequest{
		GroupName:   group_id,
		Description: fmt.Sprintf("Group %s bootstrapped", group_id),
	}

	json_content, err := json.Marshal(request_body)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(request_body, "", "  ")
	slog.InfoContext(ctx, fmt.Sprintf("Create Group URL: %s", create_group_url))
	slog.DebugContext(ctx, string(j))

	err = retry.Do(
		func() error {
			req, err := http.NewRequest("POST", create_group_url, bytes.NewBuffer(json_content))
			if err != nil {
				return err
			}

			headers, err := a._build_headers_with_partition()
			if err != nil {
				return err
			}
			req.Header = headers

			http_client := http.Client{}
			res, err := http_client.Do(req)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
			slog.DebugContext(ctx, string(body))

			slog.InfoContext(ctx, fmt.Sprintf("Created GroupId: %s | Entitlements Response: %d", group_id, res.StatusCode))

			if res.StatusCode == http.StatusConflict {
				slog.WarnContext(ctx, fmt.Sprintf("Group %s already exists", group_id))
				return nil
			}

			if res.StatusCode > http.StatusCreated {
				return fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.WarnContext(ctx, fmt.Sprintf("retry #%d: %s", n, err))
		}),
	)

	if err != nil {
		return err
	}

	// Add users to the group
	for _, user := range user_ids {
		if err := a._create_owner_member_group(group_id, user); err != nil {
			return err
		}
	}
	return nil
}

func (a OsduApiRequest) _create_owner_member_group(group_id, user_id string) error {
	ctx := context.WithValue(a.Context(), OsduApi, "entitlements.go")
	entitlements_group := fmt.Sprintf("%s@%s.%s",
		group_id,
		a.osduSettings.PartitionId,
		a.osduSettings.EntitlementsDomain,
	)

	slog.InfoContext(ctx, fmt.Sprintf("[Entitlements] Create Owner Member routine - UserId: %s | GroupId: %s", user_id, entitlements_group))

	add_user_url := fmt.Sprintf("%s/groups/%s/members",
		a.osduSettings.EntitlementsUrl,
		entitlements_group,
	)
	request_body := models.EntitlementsAddUserRequest{
		Email: user_id,
		Role:  "OWNER",
	}

	json_content, err := json.Marshal(request_body)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(request_body, "", "  ")
	slog.InfoContext(ctx, fmt.Sprintf("Add user URL: %s", add_user_url))
	slog.DebugContext(ctx, string(j))

	return retry.Do(
		func() error {
			req, err := http.NewRequest("POST", add_user_url, bytes.NewBuffer(json_content))
			if err != nil {
				return err
			}

			headers, err := a._build_headers_with_partition()
			if err != nil {
				return err
			}
			req.Header = headers

			http_client := http.Client{}
			res, err := http_client.Do(req)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
				return err
			}
			defer res.Body.Close()

			body, err := io.ReadAll(res.Body)
			if err != nil {
				slog.ErrorContext(ctx, err.Error())
			}
			slog.DebugContext(ctx, string(body))

			slog.InfoContext(ctx, fmt.Sprintf("[Entitlements] OWNER Member Created - UserId: %s | GroupId: %s | Entitlements Response: %d", user_id, entitlements_group, res.StatusCode))

			if res.StatusCode == http.StatusConflict {
				slog.WarnContext(ctx, fmt.Sprintf("User %s already exists in group %s", user_id, entitlements_group))
				return nil
			}

			if res.StatusCode > http.StatusCreated {
				return fmt.Errorf("unexpected status code: %d", res.StatusCode)
			}
			return nil
		},
		retry.Attempts(3),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.WarnContext(ctx, fmt.Sprintf("retry #%d: %s", n, err))
		}),
	)
}
