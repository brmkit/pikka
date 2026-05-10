# change HTTP_UUID for local testing
HTTP_UUID=866ea08b-41f2-4d86-a6a6-af1bbebc5243
HTTP_AGENT_CONFIG_FILE=test_agent_config_http.json
HTTP_CONFIG=$(shell base64 -i ${HTTP_AGENT_CONFIG_FILE})
HTTP_CONFIG_WIN=$(shell powershell -NoProfile -Command "[Convert]::ToBase64String([IO.File]::ReadAllBytes('$(HTTP_AGENT_CONFIG_FILE)'))")

BINARY_NAME=pikka
C2_HTTP=http

EGRESS_ORDER=$(shell echo [\"http\",\"dynamichttp\",\"dns\"] | base64 -i - )
EGRESS_ORDER_WIN=$(shell powershell -NoProfile -Command "[Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes('[\"http\",\"dynamichttp\",\"dns\"]'))")
EGRESS_FAILOVER="failover"
DEBUG=true
failedConnectionCountThresholdString=10
goCmdWin=CGO_ENABLED=0 go build
goCmdLinux=CGO_ENABLED=1 GOOS=linux go build
goCmdMacOS=CGO_ENABLED=1 GOOS=darwin go build

#goCmd=CGO_ENABLED=1 GOGARBLE=* /Users/itsafeature/go/bin/garble -literals -tiny -debug -seed random build
all: windows_http linux_http macos_http

windows_http:
	${goCmdWin} -o ${BINARY_NAME}_${C2_HTTP}.exe -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER_WIN}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG_WIN}'\
" .

linux_http:
	CGO_ENABLED=1 ${goCmdLinux} -o ${BINARY_NAME}_${C2_HTTP}.nix.bin -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
" .


macos_http:
	CGO_ENABLED=1 ${goCmdMacOS} -o ${BINARY_NAME}_${C2_HTTP}.mac.bin -tags="${C2_HTTP}" -ldflags=" \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.UUID=${HTTP_UUID}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_order=${EGRESS_ORDER}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.egress_failover=${EGRESS_FAILOVER}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils.debugString=${DEBUG}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.failedConnectionCountThresholdString=${failedConnectionCountThresholdString}' \
-X 'github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/profiles.${C2_HTTP}_initial_config=${HTTP_CONFIG}'\
" .


clean_http:
	go clean
	rm ${BINARY_NAME}_${C2_HTTP}.bin
	rm ${BINARY_NAME}_${C2_HTTP}.exe
