package c2functions

import (
	"fmt"
	c2structs "github.com/MythicMeta/MythicContainer/c2_structs"
)

var poseidonTCPC2definition = c2structs.C2Profile{
	Name:           "poseidon_tcp",
	Author:         "@its_a_feature_",
	Description:    "TCP-based P2P protocol that uses a bind connection",
	IsP2p:          true,
	IsServerRouted: true,
	GetIOCFunction: func(message c2structs.C2GetIOCMessage) c2structs.C2GetIOCMessageResponse {
		response := c2structs.C2GetIOCMessageResponse{
			Success: true,
		}
		return response
	},
	OPSECCheckFunction: func(message c2structs.C2OPSECMessage) c2structs.C2OPSECMessageResponse {
		response := c2structs.C2OPSECMessageResponse{
			Success: true,
			Message: "",
		}
		port, err := message.GetNumberArg("port")
		if err != nil {
			response.Error = err.Error()
			return response
		}
		if port == 4444 {
			response.Success = false
			response.Error = "Port 4444 is a bad choice"
			return response
		}
		response.Message = "Port OPSEC check passed"
		return response
	},
	SampleMessageFunction: func(message c2structs.C2SampleMessageMessage) c2structs.C2SampleMessageResponse {
		response := c2structs.C2SampleMessageResponse{
			Success: true,
		}
		response.Message = fmt.Sprintf("poseidon_tcp messages are formatted as follows:\n")
		response.Message += "uint32 in BigEndian format to represent the size of the message about to go on the wire.\n"
		response.Message += "base64 normal Mythic message.\n"
		response.Message += "\twhere the normal Mythic message is a 36 char UUIDv4 followed by an encrypted message."
		return response
	},
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
