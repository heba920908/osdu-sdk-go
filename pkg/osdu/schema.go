package osdu

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var api_schema_system_put = "schemas/system"

func (a OsduApiRequest) PutSystemSchema(schemaPayload []byte) error {
	schema_url := fmt.Sprintf("%s/%s", a.osduSettings.SchemaUrl, api_schema_system_put)

	var schema struct {
		SchemaInfo struct {
			SchemaIdentity struct {
				ID string `json:"id"`
			} `json:"schemaIdentity"`
		} `json:"schemaInfo"`
	}

	if err := json.Unmarshal(schemaPayload, &schema); err != nil {
		slog.Warn(fmt.Sprintf("Failed to parse schema ID: %v", err))
		schema.SchemaInfo.SchemaIdentity.ID = "unknown"
	}

	res, err := a.HttpRequestWithoutPartition("PUT", schema_url, schemaPayload)
	if err != nil {
		return err
	}

	if res.StatusCode > http.StatusBadRequest {
		if bodyBytes, err := io.ReadAll(res.Body); err == nil {
			slog.Warn(string(bodyBytes))
		}
		return fmt.Errorf("[%s] schema upload failed with status code: %d", schema.SchemaInfo.SchemaIdentity.ID, res.StatusCode)
	}

	if res.StatusCode == http.StatusBadRequest {
		slog.Warn(fmt.Sprintf("Schema %s most likely exists already", schema.SchemaInfo.SchemaIdentity.ID))
	}

	slog.Info(fmt.Sprintf("DONE SchemaUpload %s StatusCode : %d", schema.SchemaInfo.SchemaIdentity.ID, res.StatusCode))
	return nil
}
