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

	"github.com/heba920908/osdu-sdk-go/pkg/models"
)

func (a OsduApiRequest) RegisterPartition(partition models.Partition, is_system bool) error {
	partition_id := partition.Properties.DataPartitionId.Value
	ctx := context.WithValue(a.Context(), OsduApi, "register_partition")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	post_partition_url := fmt.Sprintf("%s/partitions/%s", a.osduSettings.PartitionUrl, partition_id)

	if is_system {
		post_partition_url = fmt.Sprintf("%s/partitions/system", a.osduSettings.PartitionUrl)
		// Looks that this endpoint got deprecated
		// post_partition_url = fmt.Sprintf("%s/partition/system", a.OsduSettings.PartitionUrl)
	}

	json_content, err := json.Marshal(partition)
	if err != nil {
		return err
	}

	j, _ := json.MarshalIndent(partition, "", "  ")
	slog.InfoContext(ctx, "Registering partition ---")
	slog.DebugContext(ctx, string(j))

	req, _ := http.NewRequest("POST", post_partition_url, bytes.NewBuffer([]byte(json_content)))
	/* Partition from internal service does not need to use headers */
	headers, _ := a._build_headers_without_partition()
	req.Header = headers
	//

	http_client := http.Client{}

	res, err := http_client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}

	slog.Info(fmt.Sprintf("%d", res.StatusCode))

	defer res.Body.Close()

	if res.StatusCode == http.StatusConflict {
		slog.Warn("Partition already created, trying to patch")
		req, _ = http.NewRequest("PATCH", post_partition_url, bytes.NewBuffer([]byte(json_content)))
		req.Header = headers
		res, err = http_client.Do(req)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
			return err
		}
	}

	if res.StatusCode > 205 {
		body_bytes, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error(err.Error())
		}
		status_err := fmt.Errorf("partition service response - %d : %s", res.StatusCode, string(body_bytes))
		slog.ErrorContext(ctx, status_err.Error())
		return status_err
	}

	slog.InfoContext(ctx, fmt.Sprintf("Partition %s registered", partition_id))
	return nil
}

func (a OsduApiRequest) _clean_up_partition(partitionid string) error {
	ctx := context.WithValue(a.Context(), OsduApi, "cleanup_partition")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	delete_partition_url := fmt.Sprintf("%s/partitions/%s", a.osduSettings.PartitionUrl, partitionid)

	req, _ := http.NewRequest(http.MethodDelete, delete_partition_url, nil)
	headers, _ := a._build_headers_without_partition()
	req.Header = headers

	http_client := http.Client{}

	res, err := http_client.Do(req)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("DELETE partition %s StatusCode: %d", partitionid, res.StatusCode))

	return nil
}
