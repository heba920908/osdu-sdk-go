package models

import (
	"fmt"
	"log/slog"
	"strings"
)

type PartitionProperty struct {
	Sensitive bool   `json:"sensitive"`
	Value     string `json:"value"`
}

type Partition struct {
	Properties PartitionProperties `json:"properties"`
}

type PartitionProperties struct {
	/*
	  Known properties for CI implementation
	*/
	ProjectId                     PartitionProperty `json:"projectId,omitempty"`
	ServiceAccount                PartitionProperty `json:"serviceAccount,omitempty"`
	ComplianceRuleSet             PartitionProperty `json:"complianceRuleSet,omitempty"`
	DataPartitionId               PartitionProperty `json:"dataPartitionId,omitempty"`
	Name                          PartitionProperty `json:"name,omitempty"`
	Bucket                        PartitionProperty `json:"bucket,omitempty"`
	CrmAccountID                  PartitionProperty `json:"crmAccountID,omitempty"`
	OsmPostgresDatasourceUrl      PartitionProperty `json:"osm.postgres.datasource.url,omitempty"`
	OsmPostgresDatasourceUsername PartitionProperty `json:"osm.postgres.datasource.username,omitempty"`
	OsmPostgresDatasourcePassword PartitionProperty `json:"osm.postgres.datasource.password,omitempty"`
	ObmMinioEndpoint              PartitionProperty `json:"obm.minio.endpoint,omitempty"`
	ObmMinioAccessKey             PartitionProperty `json:"obm.minio.accessKey,omitempty"`
	ObmMinioSecretKey             PartitionProperty `json:"obm.minio.secretKey,omitempty"`
	ObmMinioIgnoreCertCheck       PartitionProperty `json:"obm.minio.ignoreCertCheck,omitempty"`
	ObmMinioUiEndpoint            PartitionProperty `json:"obm.minio.ui.endpoint,omitempty"`
	KubernetesSecretName          PartitionProperty `json:"kubernetes-secret-name,omitempty"`
	OqmRabbitmqAmqpHost           PartitionProperty `json:"oqm.rabbitmq.amqp.host,omitempty"`
	OqmRabbitmqAmqpPort           PartitionProperty `json:"oqm.rabbitmq.amqp.port,omitempty"`
	OqmRabbitmqAmqpPath           PartitionProperty `json:"oqm.rabbitmq.amqp.path,omitempty"`
	OqmRabbitmqAmqpUsername       PartitionProperty `json:"oqm.rabbitmq.amqp.username,omitempty"`
	OqmRabbitmqAmqpPassword       PartitionProperty `json:"oqm.rabbitmq.amqp.password,omitempty"`
	OqmRabbitmqAdminSchema        PartitionProperty `json:"oqm.rabbitmq.admin.schema,omitempty"`
	OqmRabbitmqAdminHost          PartitionProperty `json:"oqm.rabbitmq.admin.host,omitempty"`
	OqmRabbitmqAdminPort          PartitionProperty `json:"oqm.rabbitmq.admin.port,omitempty"`
	OqmRabbitmqAdminPath          PartitionProperty `json:"oqm.rabbitmq.admin.path,omitempty"`
	OqmRabbitmqAdminUsername      PartitionProperty `json:"oqm.rabbitmq.admin.username,omitempty"`
	OqmRabbitmqAdminPassword      PartitionProperty `json:"oqm.rabbitmq.admin.password,omitempty"`
	ElasticsearchHost             PartitionProperty `json:"elasticsearch.8.host,omitempty"`
	ElasticsearchPort             PartitionProperty `json:"elasticsearch.8.port,omitempty"`
	ElasticsearchUser             PartitionProperty `json:"elasticsearch.8.user,omitempty"`
	ElasticsearchPassword         PartitionProperty `json:"elasticsearch.8.password,omitempty"`
	ElasticsearchHttps            PartitionProperty `json:"elasticsearch.8.https,omitempty"`
	ElasticsearchTls              PartitionProperty `json:"elasticsearch.8.tls,omitempty"`
	ElasticsearchSevenHost        PartitionProperty `json:"elasticsearch.host,omitempty"`
	ElasticsearchSevenPort        PartitionProperty `json:"elasticsearch.port,omitempty"`
	ElasticsearchSevenUser        PartitionProperty `json:"elasticsearch.user,omitempty"`
	ElasticsearchSevenPassword    PartitionProperty `json:"elasticsearch.password,omitempty"`
	ElasticsearchSevenHttps       PartitionProperty `json:"elasticsearch.https,omitempty"`
	ElasticsearchSevenTls         PartitionProperty `json:"elasticsearch.tls,omitempty"`
	IndexAugmenterEnabled         PartitionProperty `json:"index-augmenter-enabled,omitempty"`
	FeatureFlagPolicyEnabled      PartitionProperty `json:"featureFlag.policy.enabled,omitempty"`
	FeatureFlagOpaEnabled         PartitionProperty `json:"featureFlag.opa.enabled,omitempty"`
	ObmMinioExternalEndpoint      PartitionProperty `json:"obm.minio.external.endpoint,omitempty"`
	WellboreDmsBucket             PartitionProperty `json:"wellbore-dms-bucket,omitempty"`
	ReservoirConnection           PartitionProperty `json:"reservoir-connection,omitempty"`
	LegalBucketName               PartitionProperty `json:"legal.bucket.name,omitempty"`
	StorageBucketName             PartitionProperty `json:"storage.bucket.name,omitempty"`
	SchemaBucketName              PartitionProperty `json:"schema.bucket.name,omitempty"`
	FileStagingLocation           PartitionProperty `json:"file.staging.location,omitempty"`
	FilePersistenLocation         PartitionProperty `json:"file.persistent.location,omitempty"`
	// Needed by system partition
	EntitlementsDatasourceUrl      PartitionProperty `json:"entitlements.datasource.url,omitempty"`
	EntitlementsDatasourceUsername PartitionProperty `json:"entitlements.datasource.username,omitempty"`
	EntitlementsDatasourcePassword PartitionProperty `json:"entitlements.datasource.password,omitempty"`
	EntitlementsDatasourceSchema   PartitionProperty `json:"entitlements.datasource.schema,omitempty"`
	SystemSchemaBucketName         PartitionProperty `json:"system.schema.bucket.name,omitempty"`

	/*
	  Known properties for azure
	*/
	ComplianceRuleset                         PartitionProperty `json:"compliance-ruleset,omitempty"`
	ElasticSevenEndpoint                      PartitionProperty `json:"elastic-endpoint,omitempty"`
	ElasticSevenUsername                      PartitionProperty `json:"elastic-username,omitempty"`
	ElasticSevenPassword                      PartitionProperty `json:"elastic-password,omitempty"`
	ElasticSevenSslEnabled                    PartitionProperty `json:"elastic-ssl-enabled,omitempty"`
	CosmosConnection                          PartitionProperty `json:"cosmos-connection,omitempty"`
	CosmosEndpoint                            PartitionProperty `json:"cosmos-endpoint,omitempty"`
	CosmosPrimaryKey                          PartitionProperty `json:"cosmos-primary-key,omitempty"`
	SbConnection                              PartitionProperty `json:"sb-connection,omitempty"`
	SbNamespace                               PartitionProperty `json:"sb-namespace,omitempty"`
	StorageAccountKey                         PartitionProperty `json:"storage-account-key,omitempty"`
	StorageAccountName                        PartitionProperty `json:"storage-account-name,omitempty"`
	StorageAccountBlobEndpoint                PartitionProperty `json:"storage-account-blob-endpoint,omitempty"`
	IngestStorageAccountName                  PartitionProperty `json:"ingest-storage-account-name,omitempty"`
	IngestStorageAccountKey                   PartitionProperty `json:"ingest-storage-account-key,omitempty"`
	HierarchicalStorageAccountName            PartitionProperty `json:"hierarchical-storage-account-name,omitempty"`
	HierarchicalStorageAccountKey             PartitionProperty `json:"hierarchical-storage-account-key,omitempty"`
	EventgridRecordstopic                     PartitionProperty `json:"eventgrid-recordstopic,omitempty"`
	EventgridRecordstopicAccesskey            PartitionProperty `json:"eventgrid-recordstopic-accesskey,omitempty"`
	EventgridLegaltagschangedtopic            PartitionProperty `json:"eventgrid-legaltagschangedtopic,omitempty"`
	EventgridLegaltagschangedtopicAccesskey   PartitionProperty `json:"eventgrid-legaltagschangedtopic-accesskey,omitempty"`
	EventgridResourcegroup                    PartitionProperty `json:"eventgrid-resourcegroup,omitempty"`
	EncryptionKeyIdentifier                   PartitionProperty `json:"encryption-key-identifier,omitempty"`
	SdmsStorageAccountName                    PartitionProperty `json:"sdms-storage-account-name,omitempty"`
	SdmsStorageAccountKey                     PartitionProperty `json:"sdms-storage-account-key,omitempty"`
	EventgridSchemanotificationtopic          PartitionProperty `json:"eventgrid-schemanotificationtopic,omitempty"`
	EventgridSchemanotificationtopicAccesskey PartitionProperty `json:"eventgrid-schemanotificationtopic-accesskey,omitempty"`
	EventgridGsmtopic                         PartitionProperty `json:"eventgrid-gsmtopic,omitempty"`
	EventgridGsmtopicAccesskey                PartitionProperty `json:"eventgrid-gsmtopic-accesskey,omitempty"`
	EventgridStatuschangedtopic               PartitionProperty `json:"eventgrid-statuschangedtopic,omitempty"`
	EventgridStatuschangedtopicAccesskey      PartitionProperty `json:"eventgrid-statuschangedtopic-accesskey,omitempty"`
	EventgridSchemachangedtopic               PartitionProperty `json:"eventgrid-schemachangedtopic,omitempty"`
	EventgridSchemachangedtopicAccesskey      PartitionProperty `json:"eventgrid-schemachangedtopic-accesskey,omitempty"`
	IndexerDecimationEnabled                  PartitionProperty `json:"indexer-decimation-enabled,omitempty"`
}

func GetDefaultPartitionPropertiesCI(partition_id string) PartitionProperties {
	slog.Info(fmt.Sprintf("Gathering default values for partition CI %s", partition_id))
	root := PartitionProperties{}
	suffix_capital := strings.ToUpper(partition_id)
	// Used in the past for elasticsearch system single instance
	system_partition_capital := strings.ToUpper("system")

	bucket_prefix := fmt.Sprintf("refi-%s", partition_id)

	root.ProjectId.Value = "refi"
	root.ProjectId.Sensitive = false

	root.ServiceAccount.Value = "datafier@service.local"
	root.ServiceAccount.Sensitive = false

	root.ComplianceRuleSet.Value = "shared"
	root.ComplianceRuleSet.Sensitive = false

	root.DataPartitionId.Value = partition_id
	root.DataPartitionId.Sensitive = false

	root.Name.Value = partition_id
	root.Name.Sensitive = false

	root.Bucket.Value = fmt.Sprintf("%s-records", bucket_prefix)
	root.Bucket.Sensitive = false

	root.CrmAccountID.Value = fmt.Sprintf("[%s,%s]", partition_id, partition_id)
	root.CrmAccountID.Sensitive = false

	root.OsmPostgresDatasourceUrl.Value = fmt.Sprintf("POSTGRES_DATASOURCE_URL_%s", suffix_capital)
	root.OsmPostgresDatasourceUrl.Sensitive = true

	root.OsmPostgresDatasourceUsername.Value = fmt.Sprintf("POSTGRES_DB_USERNAME_%s", suffix_capital)
	root.OsmPostgresDatasourceUsername.Sensitive = true

	root.OsmPostgresDatasourcePassword.Value = fmt.Sprintf("POSTGRES_DB_PASSWORD_%s", suffix_capital)
	root.OsmPostgresDatasourcePassword.Sensitive = true

	root.ObmMinioEndpoint.Value = "http://minio:9001"
	root.ObmMinioEndpoint.Sensitive = false

	root.ObmMinioAccessKey.Value = fmt.Sprintf("MINIO_ACCESS_KEY_%s", suffix_capital)
	root.ObmMinioAccessKey.Sensitive = true

	root.ObmMinioSecretKey.Value = fmt.Sprintf("MINIO_SECRET_KEY_%s", suffix_capital)
	root.ObmMinioSecretKey.Sensitive = true

	root.ObmMinioIgnoreCertCheck.Value = "true"
	root.ObmMinioIgnoreCertCheck.Sensitive = false

	root.ObmMinioUiEndpoint.Value = "s3"
	root.ObmMinioUiEndpoint.Sensitive = false

	root.KubernetesSecretName.Value = "eds-osdu"
	root.KubernetesSecretName.Sensitive = false

	root.OqmRabbitmqAmqpHost.Value = "rabbitmq"
	root.OqmRabbitmqAmqpHost.Sensitive = false

	root.OqmRabbitmqAmqpPort.Value = "5672"
	root.OqmRabbitmqAmqpPort.Sensitive = false

	root.OqmRabbitmqAmqpPath.Value = ""
	root.OqmRabbitmqAmqpPath.Sensitive = false

	root.OqmRabbitmqAmqpUsername.Value = "RABBITMQ_ADMIN_USERNAME"
	root.OqmRabbitmqAmqpUsername.Sensitive = true

	root.OqmRabbitmqAmqpPassword.Value = "RABBITMQ_ADMIN_PASSWORD"
	root.OqmRabbitmqAmqpPassword.Sensitive = true

	root.OqmRabbitmqAdminSchema.Value = "http"
	root.OqmRabbitmqAdminSchema.Sensitive = false

	root.OqmRabbitmqAdminHost.Value = "rabbitmq"
	root.OqmRabbitmqAdminHost.Sensitive = false

	root.OqmRabbitmqAdminPort.Value = "15672"
	root.OqmRabbitmqAdminPort.Sensitive = false

	root.OqmRabbitmqAdminPath.Value = "/api"
	root.OqmRabbitmqAdminPath.Sensitive = false

	root.OqmRabbitmqAdminUsername.Value = "RABBITMQ_ADMIN_USERNAME"
	root.OqmRabbitmqAdminUsername.Sensitive = true

	root.OqmRabbitmqAdminPassword.Value = "RABBITMQ_ADMIN_PASSWORD"
	root.OqmRabbitmqAdminPassword.Sensitive = true

	root.ElasticsearchHost.Value = fmt.Sprintf("ELASTIC_HOST_%s", suffix_capital)
	root.ElasticsearchHost.Sensitive = true

	root.ElasticsearchPort.Value = fmt.Sprintf("ELASTIC_PORT_%s", system_partition_capital)
	root.ElasticsearchPort.Sensitive = true

	root.ElasticsearchUser.Value = fmt.Sprintf("ELASTIC_USER_%s", suffix_capital)
	root.ElasticsearchUser.Sensitive = true

	root.ElasticsearchPassword.Value = fmt.Sprintf("ELASTIC_PASS_%s", suffix_capital)
	root.ElasticsearchPassword.Sensitive = true

	root.ElasticsearchHttps.Value = "false"
	root.ElasticsearchPassword.Sensitive = false

	root.ElasticsearchTls.Value = "false"
	root.ElasticsearchPassword.Sensitive = false

	root.IndexAugmenterEnabled.Value = "false"
	root.IndexAugmenterEnabled.Sensitive = false

	root.FeatureFlagPolicyEnabled.Value = "false"
	root.FeatureFlagPolicyEnabled.Sensitive = false

	root.FeatureFlagOpaEnabled.Value = "false"
	root.FeatureFlagOpaEnabled.Sensitive = false

	root.ObmMinioExternalEndpoint.Value = "${}"
	root.ObmMinioExternalEndpoint.Sensitive = false

	root.EntitlementsDatasourceUrl.Value = "ENT_PG_URL_SYSTEM"
	root.EntitlementsDatasourceUrl.Sensitive = true

	root.EntitlementsDatasourceUsername.Value = "ENT_PG_USER_SYSTEM"
	root.EntitlementsDatasourceUsername.Sensitive = true

	root.EntitlementsDatasourcePassword.Value = "ENT_PG_PASS_SYSTEM"
	root.EntitlementsDatasourcePassword.Sensitive = true

	root.EntitlementsDatasourceSchema.Value = "ENT_PG_SCHEMA_OSDU"
	root.EntitlementsDatasourceSchema.Sensitive = true

	root.WellboreDmsBucket.Value = "refi-osdu-logstore"
	root.WellboreDmsBucket.Sensitive = false

	root.ReservoirConnection.Value = "POSTGRESQL_CONN_STRING"
	root.ReservoirConnection.Sensitive = true

	/*
	  Default bucket naming convention
	*/

	root.LegalBucketName.Value = fmt.Sprintf("%s-legal-config", bucket_prefix)
	root.LegalBucketName.Sensitive = false

	root.StorageBucketName.Value = fmt.Sprintf("%s-records", bucket_prefix)
	root.StorageBucketName.Sensitive = false

	root.SchemaBucketName.Value = fmt.Sprintf("%s-schema", bucket_prefix)
	root.SchemaBucketName.Sensitive = false

	root.SystemSchemaBucketName.Value = fmt.Sprintf("%s-system-schema", bucket_prefix)
	root.SystemSchemaBucketName.Sensitive = false

	return root
}
