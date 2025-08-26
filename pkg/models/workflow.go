package models

type RegisterWorkflow struct {
	WorkflowName             string                   `json:"workflowName"`
	RegistrationInstructions RegistrationInstructions `json:"registrationInstructions"`
	Description              string                   `json:"description"`
}

type RegistrationInstructions struct {
	DagName string `json:"dagName"`
}
