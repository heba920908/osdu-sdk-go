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
	ProjectId                     PartitionProperty `json:"projectId"`
	ServiceAccount                PartitionProperty `json:"serviceAccount"`
	ComplianceRuleSet             PartitionProperty `json:"complianceRuleSet"`
	DataPartitionId               PartitionProperty `json:"dataPartitionId"`
	Name                          PartitionProperty `json:"name"`
	Bucket                        PartitionProperty `json:"bucket"`
	CrmAccountID                  PartitionProperty `json:"crmAccountID"`
	OsmPostgresDatasourceUrl      PartitionProperty `json:"osm.postgres.datasource.url"`
	OsmPostgresDatasourceUsername PartitionProperty `json:"osm.postgres.datasource.username"`
	OsmPostgresDatasourcePassword PartitionProperty `json:"osm.postgres.datasource.password"`
	ObmMinioEndpoint              PartitionProperty `json:"obm.minio.endpoint"`
	ObmMinioAccessKey             PartitionProperty `json:"obm.minio.accessKey"`
	ObmMinioSecretKey             PartitionProperty `json:"obm.minio.secretKey"`
	ObmMinioIgnoreCertCheck       PartitionProperty `json:"obm.minio.ignoreCertCheck"`
	ObmMinioUiEndpoint            PartitionProperty `json:"obm.minio.ui.endpoint"`
	KubernetesSecretName          PartitionProperty `json:"kubernetes-secret-name"`
	OqmRabbitmqAmqpHost           PartitionProperty `json:"oqm.rabbitmq.amqp.host"`
	OqmRabbitmqAmqpPort           PartitionProperty `json:"oqm.rabbitmq.amqp.port"`
	OqmRabbitmqAmqpPath           PartitionProperty `json:"oqm.rabbitmq.amqp.path"`
	OqmRabbitmqAmqpUsername       PartitionProperty `json:"oqm.rabbitmq.amqp.username"`
	OqmRabbitmqAmqpPassword       PartitionProperty `json:"oqm.rabbitmq.amqp.password"`
	OqmRabbitmqAdminSchema        PartitionProperty `json:"oqm.rabbitmq.admin.schema"`
	OqmRabbitmqAdminHost          PartitionProperty `json:"oqm.rabbitmq.admin.host"`
	OqmRabbitmqAdminPort          PartitionProperty `json:"oqm.rabbitmq.admin.port"`
	OqmRabbitmqAdminPath          PartitionProperty `json:"oqm.rabbitmq.admin.path"`
	OqmRabbitmqAdminUsername      PartitionProperty `json:"oqm.rabbitmq.admin.username"`
	OqmRabbitmqAdminPassword      PartitionProperty `json:"oqm.rabbitmq.admin.password"`
	ElasticsearchHost             PartitionProperty `json:"elasticsearch.host"`
	ElasticsearchPort             PartitionProperty `json:"elasticsearch.port"`
	ElasticsearchUser             PartitionProperty `json:"elasticsearch.user"`
	ElasticsearchPassword         PartitionProperty `json:"elasticsearch.password"`
	ElasticsearchHttps            PartitionProperty `json:"elasticsearch.https"`
	ElasticsearchTls              PartitionProperty `json:"elasticsearch.tls"`
	IndexAugmenterEnabled         PartitionProperty `json:"index-augmenter-enabled"`
	PolicyServiceEnabled          PartitionProperty `json:"policy-service-enabled"`
	ObmMinioExternalEndpoint      PartitionProperty `json:"obm.minio.external.endpoint"`
	WellboreDmsBucket             PartitionProperty `json:"wellbore-dms-bucket"`
	ReservoirConnection           PartitionProperty `json:"reservoir-connection"`
	SlbFeatureAggregateApi        PartitionProperty `json:"slb-feature-aggregate-api,omitempty"`
	FeatureFlagOpaEnabled         PartitionProperty `json:"featureFlag.opa.enabled"`
	LegalBucketName               PartitionProperty `json:"legal.bucket.name"`
	StorageBucketName             PartitionProperty `json:"storage.bucket.name"`
	SchemaBucketName              PartitionProperty `json:"schema.bucket.name"`
	// Needed by system partition
	EntitlementsDatasourceUrl      PartitionProperty `json:"entitlements.datasource.url"`
	EntitlementsDatasourceUsername PartitionProperty `json:"entitlements.datasource.username"`
	EntitlementsDatasourcePassword PartitionProperty `json:"entitlements.datasource.password"`
	EntitlementsDatasourceSchema   PartitionProperty `json:"entitlements.datasource.schema"`
	SystemSchemaBucketName         PartitionProperty `json:"system.schema.bucket.name"`
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

	root.PolicyServiceEnabled.Value = "false"
	root.PolicyServiceEnabled.Sensitive = false

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

	root.SlbFeatureAggregateApi.Value = "true"
	root.SlbFeatureAggregateApi.Sensitive = false

	root.FeatureFlagOpaEnabled.Value = "false"
	root.FeatureFlagOpaEnabled.Sensitive = false

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
