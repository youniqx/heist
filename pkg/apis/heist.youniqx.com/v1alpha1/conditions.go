package v1alpha1

var Conditions = &ConditionsWrapper{
	Reasons: &ConditionReason{
		Provisioned:     "provisioned",
		Terminating:     "terminating",
		ErrorVault:      "vault_error",
		Initializing:    "initializing",
		ErrorConfig:     "config_error",
		ErrorKubernetes: "kubernetes_error",
	},
	Types: &ConditionType{
		Provisioned: "Provisioned",
		Active:      "Active",
	},
}

type ConditionReason struct {
	Provisioned     string
	Terminating     string
	ErrorVault      string
	Initializing    string
	ErrorConfig     string
	ErrorKubernetes string
}

type ConditionType struct {
	Provisioned string
	Active      string
}

type ConditionsWrapper struct {
	Reasons *ConditionReason
	Types   *ConditionType
}
