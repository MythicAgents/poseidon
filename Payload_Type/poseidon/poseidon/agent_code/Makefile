HTTP_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
HTTP_CALLBACK_HOST=http://127.0.0.1
HTTP_CALLBACK_PORT=80
HTTP_KILLDATE=2024-12-31
HTTP_CALLBACK_INTERVAL=2
HTTP_CALLBACK_JITTER=0
HTTP_POST_URI=data
HTTP_ENCRYPTED_EXCHANGE_CHECK=false
HTTP_HEADERS="{}"
HTTP_AESPSK=hfN9Nk29S8LsjrE9ffbT9KONue4uozk+/TVMyrxDvvM=
HTTP_PROXY_PORT=
HTTP_PROXY_USER=
HTTP_PROXY_HOST=
HTTP_PROXY_BYPASS=False

WEBSOCKET_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
WEBSOCKET_CALLBACK_HOST=ws://127.0.0.1
WEBSOCKET_CALLBACK_PORT=8081
WEBSOCKET_AESPSK=qgbrtDu3eaYYUXZmSuaeMpkzI+5mkVjr00iQ4p/xtpU=
WEBSOCKET_ENDPOINT_REPLACE=socket
WEBSOCKET_USER_AGENT=
WEBSOCKET_CALLBACK_INTERVAL=2
WEBSOCKET_ENCRYPTED_EXCHANGE_CHECK=true
WEBSOCKET_DOMAIN_FRONT=
WEBSOCKET_CALLBACK_JITTER=20
WEBSOCKET_TASKING_TYPE="Push"
WEBSOCKET_KILLDATE="2024-12-31"

TCP_UUID=80844d19-9bfc-47f9-b9af-c6b9144c0fdc
TCP_PORT=8085
TCP_KILLDATE=2024-12-31
TCP_ENCRYPTED_EXCHANGE_CHECK=true
TCP_AESPSK=7JSBbGON1cHI4xtpxR0M41qQulCBD+DgyABLr6hpjFc=

BINARY_NAME=poseidon
C2_HTTP=http
C2_WEBSOCKET=websocket
C2_TCP=poseidon_tcp
C2_DYNAMICHTTP=dynamichttp

EGRESS_ORDER=[\"http\",\"websocket\",\"dynamichttp\"]
EGRESS_FAILOVER="round-robin"
DEBUG=true
failedConnectionCountThresholdString=10

DYNAMICHTTP_UUID=fbd34459-f783-4b78-9d4c-fd26a5289fee
DYNAMICHTTP_KILL_DATE=2024-12-31
DYNAMICHTTP_INTERVAL=2
DYNAMICHTTP_JITTER=0
DYNAMICHTTP_ENCRYPTED_EXCHANGE_CHECK=true
DYNAMICHTTP_AESPSK=YBRX8nRTYUrkKue/HhEgm+F7gG304uqPdZQwzZ+vaL8=
DYNAMICHTTP_AGENT_CONFIG_FILE=dynamichttp_test_agent_config.json
DYNAMICHTTP_AGENT_CONFIG=$(shell cat ${DYNAMICHTTP_AGENT_CONFIG_FILE} |  sed 's/\\/\\\\/g' |  sed 's/"/\\"/g' | tr -d '\n')

build_http:
	go build -o ${BINARY_NAME}_${C2_HTTP}.bin -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_host=${HTTP_CALLBACK_HOST}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_port=${HTTP_CALLBACK_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_killdate=${HTTP_KILLDATE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_interval=${HTTP_CALLBACK_INTERVAL}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_jitter=${HTTP_CALLBACK_JITTER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_post_uri=${HTTP_POST_URI}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_encrypted_exchange_check=${HTTP_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_AESPSK=${HTTP_AESPSK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_port=${HTTP_PROXY_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_user=${HTTP_PROXY_USER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_host=${HTTP_PROXY_HOST}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_bypass=${HTTP_PROXY_BYPASS}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_headers=${HTTP_HEADERS}' \
" .

build_tcp:
	go build -o ${BINARY_NAME}_${C2_TCP}.bin -tags="${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${TCP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_port=${TCP_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_killdate=${TCP_KILLDATE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_encrypted_exchange_check=${TCP_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_AESPSK=${TCP_AESPSK}' \
" .

build_http_tcp:
	go build -o ${BINARY_NAME}_${C2_HTTP}_${C2_TCP}.bin -tags="${C2_HTTP},${C2_TCP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_host=${HTTP_CALLBACK_HOST}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_port=${HTTP_CALLBACK_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_killdate=${HTTP_KILLDATE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_interval=${HTTP_CALLBACK_INTERVAL}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_jitter=${HTTP_CALLBACK_JITTER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_post_uri=${HTTP_POST_URI}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_encrypted_exchange_check=${HTTP_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_AESPSK=${HTTP_AESPSK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_port=${HTTP_PROXY_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_user=${HTTP_PROXY_USER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_host=${HTTP_PROXY_HOST}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_bypass=${HTTP_PROXY_BYPASS}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_headers=${HTTP_HEADERS}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${TCP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_port=${TCP_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_killdate=${TCP_KILLDATE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_encrypted_exchange_check=${TCP_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_TCP}_AESPSK=${TCP_AESPSK}' \
" .

build_websocket:
	go build -o ${BINARY_NAME}_${C2_WEBSOCKET}.bin -tags="${C2_WEBSOCKET}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_host=${WEBSOCKET_CALLBACK_HOST}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_port=${WEBSOCKET_CALLBACK_PORT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_interval=${WEBSOCKET_CALLBACK_INTERVAL}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_jitter=${WEBSOCKET_CALLBACK_JITTER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_encrypted_exchange_check=${WEBSOCKET_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_AESPSK=${WEBSOCKET_AESPSK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_domain_front=${WEBSOCKET_DOMAIN_FRONT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_ENDPOINT_REPLACE=${WEBSOCKET_ENDPOINT_REPLACE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_USER_AGENT=${WEBSOCKET_USER_AGENT}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_tasking_type=${WEBSOCKET_TASKING_TYPE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_killdate=${WEBSOCKET_KILLDATE}' \
" .

build_websocket_http:
	go build -o ${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin -tags="${C2_WEBSOCKET},${C2_HTTP}" -ldflags=" \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_host=${WEBSOCKET_CALLBACK_HOST}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${WEBSOCKET_UUID}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
	-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_port=${WEBSOCKET_CALLBACK_PORT}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_interval=${WEBSOCKET_CALLBACK_INTERVAL}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_callback_jitter=${WEBSOCKET_CALLBACK_JITTER}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_encrypted_exchange_check=${WEBSOCKET_ENCRYPTED_EXCHANGE_CHECK}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_AESPSK=${WEBSOCKET_AESPSK}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_domain_front=${WEBSOCKET_DOMAIN_FRONT}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_ENDPOINT_REPLACE=${WEBSOCKET_ENDPOINT_REPLACE}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_USER_AGENT=${WEBSOCKET_USER_AGENT}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_tasking_type=${WEBSOCKET_TASKING_TYPE}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_WEBSOCKET}_killdate=${WEBSOCKET_KILLDATE}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_host=${HTTP_CALLBACK_HOST}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_port=${HTTP_CALLBACK_PORT}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_killdate=${HTTP_KILLDATE}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_interval=${HTTP_CALLBACK_INTERVAL}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_callback_jitter=${HTTP_CALLBACK_JITTER}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_post_uri=${HTTP_POST_URI}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_encrypted_exchange_check=${HTTP_ENCRYPTED_EXCHANGE_CHECK}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_AESPSK=${HTTP_AESPSK}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_port=${HTTP_PROXY_PORT}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_user=${HTTP_PROXY_USER}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_host=${HTTP_PROXY_HOST}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_proxy_bypass=${HTTP_PROXY_BYPASS}' \
    -X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_HTTP}_headers=${HTTP_HEADERS}' \
    " .

build_dynamichttp:
	go build -o ${BINARY_NAME}_${C2_DYNAMICHTTP}.bin -tags="${C2_DYNAMICHTTP}" -ldflags=" \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.UUID=${DYNAMICHTTP_UUID}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_killdate=${DYNAMICHTTP_KILL_DATE}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_callback_interval=${DYNAMICHTTP_INTERVAL}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_callback_jitter=${DYNAMICHTTP_JITTER}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_encrypted_exchange_check=${DYNAMICHTTP_ENCRYPTED_EXCHANGE_CHECK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_AESPSK=${DYNAMICHTTP_AESPSK}' \
-X 'github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles.${C2_DYNAMICHTTP}_raw_c2_config=${DYNAMICHTTP_AGENT_CONFIG}' \
" .

run_http:
	./${BINARY_NAME}_${C2_HTTP}.bin

run_tcp:
	./${BINARY_NAME}_${C2_TCP}.bin

run_dynamichttp:
	./${BINARY_NAME}_${C2_DYNAMICHTTP}.bin

run_websocket_http:
	./${BINARY_NAME}_${C2_WEBSOCKET}_${C2_HTTP}.bin

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