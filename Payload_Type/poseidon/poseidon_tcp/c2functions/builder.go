package c2functions

import (
	c2structs "github.com/MythicMeta/MythicContainer/c2_structs"
)

var poseidonTCPC2definition = c2structs.C2Profile{
	Name:           "poseidon_tcp",
	Author:         "@its_a_feature_",
	Description:    "TCP-based P2P protocol that uses a bind connection",
	IsP2p:          true,
	IsServerRouted: true,
}
var poseidonTCPC2parameters = []c2structs.C2Parameter{
	{
		Name:          "port",
		Description:   "Port to open",
		DefaultValue:  8085,
		ParameterType: c2structs.C2_PARAMETER_TYPE_NUMBER,
		Required:      false,
		FormatString:  "[0-65535]{1}",
		VerifierRegex: "^[0-9]+$",
	},
	{
		Name:          "killdate",
		Description:   "Kill Date",
		DefaultValue:  365,
		ParameterType: c2structs.C2_PARAMETER_TYPE_DATE,
		Required:      false,
	},
	{
		Name:          "encrypted_exchange_check",
		Description:   "Perform Key Exchange",
		DefaultValue:  true,
		ParameterType: c2structs.C2_PARAMETER_TYPE_BOOLEAN,
		Required:      false,
	},
	{
		Name:          "AESPSK",
		Description:   "Encryption Type",
		DefaultValue:  "aes256_hmac",
		ParameterType: c2structs.C2_PARAMETER_TYPE_CHOOSE_ONE,
		Required:      false,
		IsCryptoType:  true,
		Choices: []string{
			"aes256_hmac",
			"none",
		},
	},
}

func Initialize() {
	c2structs.AllC2Data.Get("poseidon_tcp").AddC2Definition(poseidonTCPC2definition)
	c2structs.AllC2Data.Get("poseidon_tcp").AddParameters(poseidonTCPC2parameters)
}
