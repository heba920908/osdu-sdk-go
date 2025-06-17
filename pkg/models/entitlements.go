package models

type EntitlementsBoostrapUser struct {
	AliasId string `json:"aliasId"`
	UserId  string `json:"userId"`
}

/*
https://community.opengroup.org/osdu/platform/security-and-compliance/entitlements/-/blob/release/0.27/provider/entitlements-v2-jdbc/bootstrap/bootstrap.sh?ref_type=heads
*/

func DefaultEntitlementsBootstrapUsers() []EntitlementsBoostrapUser {
	m := make(map[string]string)
	m["SERVICE_PRINCIPAL"] = "osdu-admin@service.local"
	m["SERVICE_PRINCIPAL_AIRFLOW"] = "airflow@service.local"
	m["SERVICE_PRINCIPAL_INDEXER"] = "indexer@service.local"
	m["SERVICE_PRINCIPAL_GCZ_TRANSFORMER"] = "gcz-transformer@service.local"
	m["SERVICE_PRINCIPAL_REGISTER"] = "register@service.local"
	m["SERVICE_PRINCIPAL_NOTIFICATION"] = "notification@service.local"
	m["SERVICE_PRINCIPAL_STORAGE"] = "storage@service.local"
	m["SERVICE_PRINCIPAL_SEISMIC"] = "seismic@service.local"

	var bootstrap_users []EntitlementsBoostrapUser

	for k, v := range m {
		bootstrap_users = append(bootstrap_users, EntitlementsBoostrapUser{
			AliasId: k,
			UserId:  v,
		})
	}
	return bootstrap_users
}
