package osdu

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/heba920908/osdu-sdk-go/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPartitionProvisioning(t *testing.T) {
	partition_name := petname.Generate(1, " ")
	os.Setenv("CONFIG_FILE", "../../config/default.yaml")
	osdu_client := NewClient()
	partitionProperties := models.GetDefaultPartitionPropertiesCI(partition_name)
	partition := models.Partition{
		Properties: partitionProperties,
	}
	defer osdu_client._clean_up_partition(partition_name)
	err := osdu_client.RegisterPartition(partition, false)
	assert.NoError(t, err)
	// Also test entitlements bootstrap
	err = osdu_client.EntitlementsBootstrap()
	assert.NoError(t, err)
}

func TestCreateEntitlementsAdminUser(t *testing.T) {
	os.Setenv("CONFIG_FILE", "../../config/default.yaml")
	osdu_client := NewClient()
	err := osdu_client.CreateEntitlementsAdminUser(fmt.Sprintf("%s@example.com", petname.Generate(1, "")))
	assert.NoError(t, err)
}

func TestPartitionWithOverride(t *testing.T) {
	partition_name := petname.Generate(1, " ")
	os.Setenv("CONFIG_FILE", "../../config/default.yaml")
	partitionOverride := `{
			"properties": {
				"obm.minio.endpoint": {
					"sensitive": false,
					"value": "http://test-minion:9000"
				},
				"obm.minio.external.endpoint": {
					"sensitive": false,
					"value": "https://s3-some-external.test-minion:9000"
				}
			}
		}`
	partitionProperties := models.GetDefaultPartitionPropertiesCI(partition_name)

	partition := models.Partition{
		Properties: partitionProperties,
	}

	err := json.NewDecoder(strings.NewReader(partitionOverride)).Decode(&partition)
	assert.NoError(t, err)

	osdu_client := NewClient()

	defer osdu_client._clean_up_partition(partition_name)
	assert.NoError(t, osdu_client.RegisterPartition(partition, false))
}
