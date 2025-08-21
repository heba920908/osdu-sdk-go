package osdu

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	retry "github.com/avast/retry-go"
)

var api_schema_system_put = "schemas/system"

func (a OsduApiRequest) PutSystemSchema(schemaPayload []byte) error {
	schema_url := fmt.Sprintf("%s/%s", a.osduSettings.SchemaUrl, api_schema_system_put)

	var schema struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(schemaPayload, &schema); err != nil {
		slog.Warn(fmt.Sprintf("Failed to parse schema ID: %v", err))
		schema.ID = "unknown"
	}

	err := retry.Do(
		func() error {
			res, err := a.HttpRequestWithoutPartition("PUT", schema_url, schemaPayload)
			if err != nil {
				return err
			}

			if res.StatusCode > http.StatusBadRequest {
				if bodyBytes, err := io.ReadAll(res.Body); err == nil {
					slog.Warn(string(bodyBytes))
				}
				return fmt.Errorf("[%s] schema unexpected status code: %d", schema.ID, res.StatusCode)
			}

			if res.StatusCode == http.StatusBadRequest {
				slog.Warn(fmt.Sprintf("Schema %s most likely exists already", schema.ID))
			}

			slog.Info(fmt.Sprintf("DONE SchemaUpload %s StatusCode : %d", schema.ID, res.StatusCode))
			return nil
		},
		retry.Attempts(3),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			slog.Warn(fmt.Sprintf("[%s] Schema Upload retry #%d: %s\n", schema.ID, n, err))
		}),
	)

	return err
}
