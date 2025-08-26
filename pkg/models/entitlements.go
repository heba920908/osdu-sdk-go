package models

type EntitlementsBoostrapUser struct {
	AliasId string `json:"aliasId"`
	UserId  string `json:"userId"`
}

type EntitlementsBootstrapRequest struct {
	AliasMappings []EntitlementsBoostrapUser `json:"aliasMappings"`
}

type EntitlementsAddUserRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type EntitlementsCreateGroupRequest struct {
	GroupName   string `json:"name"`
	Description string `json:"description"`
}

/*
https://community.opengroup.org/osdu/platform/security-and-compliance/entitlements/-/blob/release/0.27/provider/entitlements-v2-jdbc/bootstrap/bootstrap.sh?ref_type=heads
*/

func DefaultEntitlementsBootstrapUsers() []EntitlementsBoostrapUser {
	extra_service_principals := []string{
		"datafier@service.local",
		"osdu-admin@service.local",
	}

	/*
	  These should match the file groups in entitlements service for core-plus
	  https://community.opengroup.org/osdu/platform/security-and-compliance/entitlements/-/tree/ce11b9780cc2130f939cbe1811511e58cb674d89/entitlements-v2-core-plus/src/main/resources/provisioning/accounts
	  i.e groups_of_<>.json
	*/

	m := make(map[string]string)
	m["SERVICE_PRINCIPAL_AIRFLOW"] = "airflow@service.local"
	m["SERVICE_PRINCIPAL_INDEXER"] = "indexer@service.local"
	m["SERVICE_PRINCIPAL_GCZ"] = "gcz-transformer@service.local"
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

	for _, sp := range extra_service_principals {
		bootstrap_users = append(bootstrap_users, EntitlementsBoostrapUser{
			AliasId: "SERVICE_PRINCIPAL",
			UserId:  sp,
		})
	}

	return bootstrap_users
}
