package osdu

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
	"github.com/heba920908/osdu-sdk-go/pkg/models"
)

func (a OsduApiRequest) RegisterWorkflow(wr models.RegisterWorkflow) error {
	ctx := context.WithValue(a.Context(), OsduApi, "register_workflow")
	create_workflow_url := fmt.Sprintf("%s/workflow", a.osduSettings.WorkflowUrl)

	json_content, err := json.Marshal(wr)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(wr, "", "  ")
	slog.InfoContext(ctx, fmt.Sprintf("Registering workflow %s", wr.WorkflowName))
	slog.DebugContext(ctx, string(j))

	return retry.Do(
		func() error {
			http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

			req, err := http.NewRequest("POST", create_workflow_url, bytes.NewBuffer(json_content))
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

			slog.InfoContext(ctx, fmt.Sprintf("Workflow registration StatusCode: %d", res.StatusCode))

			if res.StatusCode == http.StatusConflict {
				slog.WarnContext(ctx, fmt.Sprintf("Workflow %s already registered", wr.WorkflowName))
				return nil
			}

			if res.StatusCode > 205 {
				body_bytes, err := io.ReadAll(res.Body)
				if err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Failed to read response body: %v", err))
				}
				status_err := fmt.Errorf("workflow service response - %d : %s", res.StatusCode, string(body_bytes))
				slog.ErrorContext(ctx, status_err.Error())
				return status_err
			}

			slog.InfoContext(ctx, fmt.Sprintf("Workflow %s registered successfully", wr.WorkflowName))
			return nil
		},
		retry.Attempts(3),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.WarnContext(ctx, fmt.Sprintf("Workflow registration retry #%d: %s", n, err))
		}),
	)
}
