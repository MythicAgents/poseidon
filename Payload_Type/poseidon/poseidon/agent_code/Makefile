HTTP_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
HTTP_AGENT_CONFIG_FILE=test_agent_config_http.json
HTTP_CONFIG=$(shell base64 -i ${HTTP_AGENT_CONFIG_FILE})

WEBSOCKET_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
WEBSOCKET_AGENT_CONFIG_FILE=test_agent_config_websocket.json
WEBSOCKET_CONFIG=$(shell base64 -i ${WEBSOCKET_AGENT_CONFIG_FILE})

TCP_UUID=e053199c-3be8-4901-b8d2-fecb444fcd9e
TCP_AGENT_CONFIG_FILE=test_agent_config_tcp.json
TCP_CONFIG=$(shell base64 -i ${TCP_AGENT_CONFIG_FILE})

DYNAMICHTTP_UUID=fbd34459-f783-4b78-9d4c-fd26a5289fee
DYNAMICHTTP_AGENT_CONFIG_FILE=test_agent_config_dynamichttp.json
DYNAMICHTTP_CONFIG=$(shell base64 -i ${DYNAMICHTTP_AGENT_CONFIG_FILE})

HTTPX_UUID=6d1dce05-6f50-4043-9dd0-231f771b20ce
HTTPX_AGENT_CONFIG_FILE=test_agent_config_httpx.json
HTTPX_CONFIG=$(shell base64 -i ${HTTPX_AGENT_CONFIG_FILE})

DNS_UUID=7461e218-d766-4937-87a1-957d9b718455
DNS_AGENT_CONFIG_FILE=test_agent_config_dns.json
DNS_CONFIG=$(shell base64 -i ${DNS_AGENT_CONFIG_FILE})

BINARY_NAME=poseidon
C2_HTTP=http
C2_WEBSOCKET=websocket
C2_TCP=tcp
C2_DYNAMICHTTP=dynamichttp
C2_HTTPX=httpx
C2_DNS=dns

EGRESS_ORDER=$(shell echo [\"http\",\"websocket\",\"dynamichttp\",\"dns\"] | base64 -i - )
EGRESS_FAILOVER="failover"
DEBUG=true
failedConnectionCountThresholdString=10
goCmd=CGO_ENABLED=1 go build
#goCmd=CGO_ENABLED=1 GOGARBLE=* /Users/itsafeature/go/bin/garble -literals -tiny -debug -seed random build


build_protobuf_go:
	protoc --go_out=`pwd`/pkg/profiles/dnsgrpc --go_opt=paths=source_relative \
--go-grpc_out=`pwd`/pkg/profiles/dnsgrpc --go-grpc_opt=paths=source_relative \
--proto_path=`pwd`/pkg/profiles/dnsgrpc \
`pwd`/pkg/profiles/dnsgrpc/*.proto


build_http:
	${goCmd} -o ${BINARY_NAME}_${C2_HTTP}.bin -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
" .

build_tcp:
	${goCmd} -o ${BINARY_NAME}_${C2_TCP} -tags="${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${TCP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_initial_config=${TCP_CONFIG}'\
" .

build_http_tcp:
	${goCmd} -o ${BINARY_NAME}_${C2_HTTP}_${C2_TCP} -tags="${C2_HTTP},${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_initial_config=${TCP_CONFIG}'\
" .

build_websocket:
	${goCmd} -o ${BINARY_NAME}_${C2_WEBSOCKET}.bin -tags="${C2_WEBSOCKET}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_initial_config=${WEBSOCKET_CONFIG}'\
" .

build_websocket_http:
	${goCmd} -o ${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin -tags="${C2_WEBSOCKET},${C2_HTTP}" -ldflags=" \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_initial_config=${WEBSOCKET_CONFIG}'\
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
    " .

build_dynamichttp:
	${goCmd} -o ${BINARY_NAME}_${C2_DYNAMICHTTP}.bin -tags="${C2_DYNAMICHTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${DYNAMICHTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_initial_config=${DYNAMICHTTP_CONFIG}'\
" .

build_httpx:
	${goCmd} -o ${BINARY_NAME}_${C2_HTTPX}.bin -tags="${C2_HTTPX}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTPX_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTPX}_initial_config=${HTTPX_CONFIG}'\
" .

build_dns:
	${goCmd} -o ${BINARY_NAME}_${C2_DNS}.bin -tags="${C2_DNS}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${DNS_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DNS}_initial_config=${DNS_CONFIG}'\
" .

run_http:
	./${BINARY_NAME}_${C2_HTTP}.bin

run_tcp:
	./${BINARY_NAME}_${C2_TCP}

run_websocket:
	./${BINARY_NAME}_${C2_WEBSOCKET}.bin

run_dynamichttp:
	./${BINARY_NAME}_${C2_DYNAMICHTTP}.bin

run_websocket_http:
	./${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin

run_httpx:
	./${BINARY_NAME}_${C2_HTTPX}.bin

run_dns:
	./${BINARY_NAME}_${C2_DNS}.bin

build_and_run_http: build_http run_http

build_and_run_tcp: build_tcp run_tcp

build_and_run_dynamichttp: build_dynamichttp run_dynamichttp

build_and_run_websocket_http: build_websocket_http run_websocket_http

build_and_run_websocket: build_websocket run_websocket

build_and_run_httpx: build_httpx run_httpx

build_and_run_dns: build_dns run_dns

build_all: build_http build_tcp build_websocket

clean_http:
	go clean
	rm ${BINARY_NAME}_${C2_HTTP}.bin
	rm ${BINARY_NAME}_${C2_TCP}
	rm ${BINARY_NAME}_${C2_WEBSOCKET}.bin