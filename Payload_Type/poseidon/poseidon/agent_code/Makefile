HTTP_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
HTTP_AGENT_CONFIG_FILE=test_agent_config_http.json
HTTP_CONFIG=$(shell base64 -i ${HTTP_AGENT_CONFIG_FILE})

WEBSOCKET_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
WEBSOCKET_AGENT_CONFIG_FILE=test_agent_config_websocket.json
WEBSOCKET_CONFIG=$(shell base64 -i ${WEBSOCKET_AGENT_CONFIG_FILE})

TCP_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
TCP_AGENT_CONFIG_FILE=test_agent_config_tcp.json
TCP_CONFIG=$(shell base64 -i ${TCP_AGENT_CONFIG_FILE})

DYNAMICHTTP_UUID=fbd34459-f783-4b78-9d4c-fd26a5289fee
DYNAMICHTTP_AGENT_CONFIG_FILE=test_agent_config_dynamichttp.json
DYNAMICHTTP_CONFIG=$(shell base64 -i ${DYNAMICHTTP_AGENT_CONFIG_FILE})

HTTPX_UUID=215a0336-420e-4845-93c8-2fc194637185
HTTPX_AGENT_CONFIG_FILE=test_agent_config_httpx.json
HTTPX_CONFIG=$(shell base64 -i ${HTTPX_AGENT_CONFIG_FILE})

BINARY_NAME=poseidon
C2_HTTP=http
C2_WEBSOCKET=websocket
C2_TCP=poseidon_tcp
C2_DYNAMICHTTP=dynamichttp
C2_HTTPX=httpx

EGRESS_ORDER=$(shell echo [\"http\",\"websocket\",\"dynamichttp\"] | base64 -i - )
EGRESS_FAILOVER="failover"
DEBUG=true
failedConnectionCountThresholdString=10



build_http:
	go build -o ${BINARY_NAME}_${C2_HTTP}.bin -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
" .

build_tcp:
	go build -o ${BINARY_NAME}_${C2_TCP}.bin -tags="${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${TCP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_initial_config=${TCP_CONFIG}'\
" .

build_http_tcp:
	go build -o ${BINARY_NAME}_${C2_HTTP}_${C2_TCP}.bin -tags="${C2_HTTP},${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_initial_config=${TCP_CONFIG}'\
" .

build_websocket:
	go build -o ${BINARY_NAME}_${C2_WEBSOCKET}.bin -tags="${C2_WEBSOCKET}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_initial_config=${WEBSOCKET_CONFIG}'\
" .

build_websocket_http:
	go build -o ${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin -tags="${C2_WEBSOCKET},${C2_HTTP}" -ldflags=" \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_initial_config=${WEBSOCKET_CONFIG}'\
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
    " .

build_dynamichttp:
	go build -o ${BINARY_NAME}_${C2_DYNAMICHTTP}.bin -tags="${C2_DYNAMICHTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${DYNAMICHTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_initial_config=${DYNAMICHTTP_CONFIG}'\
" .

build_httpx:
	go build -o ${BINARY_NAME}_${C2_HTTPX}.bin -tags="${C2_HTTPX}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTPX_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTPX}_initial_config=${HTTPX_CONFIG}'\
" .

run_http:
	./${BINARY_NAME}_${C2_HTTP}.bin

run_tcp:
	./${BINARY_NAME}_${C2_TCP}.bin

run_dynamichttp:
	./${BINARY_NAME}_${C2_DYNAMICHTTP}.bin

run_websocket_http:
	./${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin

run_httpx:
	./${BINARY_NAME}_${C2_HTTPX}.bin

build_and_run_http: build_http run_http

build_and_run_tcp: build_tcp run_tcp

build_and_run_dynamichttp: build_dynamichttp run_dynamichttp

build_and_run_websocket_http: build_websocket_http run_websocket_http

build_all: build_http build_tcp build_websocket

clean_http:
	go clean
	rm ${BINARY_NAME}_${C2_HTTP}.bin
	rm ${BINARY_NAME}_${C2_TCP}.bin
	rm ${BINARY_NAME}_${C2_WEBSOCKET}.bin