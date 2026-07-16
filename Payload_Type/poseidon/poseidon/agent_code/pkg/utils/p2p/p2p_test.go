package p2p

import (
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/profiles"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/responses"
	"github.com/MythicAgents/poseidon/Payload_Type/poseidon/agent_code/pkg/utils/structs"
)

const (
	testSourceUUID       = "source-callback-uuid"
	testTemporaryUUID    = "temporary-connection-uuid"
	testIntermediateUUID = "intermediate-connection-uuid"
	testCallbackUUID     = "destination-callback-uuid"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type failingReadCloser struct{}

func (failingReadCloser) Read([]byte) (int, error) {
	return 0, errors.New("response body read failed")
}

func (failingReadCloser) Close() error {
	return nil
}

func resetP2PTestState(t *testing.T) {
	t.Helper()

	oldMythicID := profiles.GetMythicID()
	oldClient := client

	clearTCPConnections()
	internalWebshellConnectionMutex.Lock()
	internalWebshellConnections = make(map[string]Arguments)
	internalWebshellConnectionMutex.Unlock()
	uuidMappingsSync.Lock()
	uuidMappings = make(map[string]string)
	uuidMappingsSync.Unlock()
	drainP2PTestChannels()
	profiles.SetMythicID(testSourceUUID)

	t.Cleanup(func() {
		clearTCPConnections()
		internalWebshellConnectionMutex.Lock()
		internalWebshellConnections = make(map[string]Arguments)
		internalWebshellConnectionMutex.Unlock()
		uuidMappingsSync.Lock()
		uuidMappings = make(map[string]string)
		uuidMappingsSync.Unlock()
		client = oldClient
		profiles.SetMythicID(oldMythicID)
		drainP2PTestChannels()
	})
}

func clearTCPConnections() {
	internalTCPConnectionMutex.Lock()
	defer internalTCPConnectionMutex.Unlock()
	for _, connection := range internalTCPConnections {
		(*connection).Close()
	}
	internalTCPConnections = make(map[string]*net.Conn)
}

func drainP2PTestChannels() {
	for {
		select {
		case <-RemoveInternalConnectionChannel:
		default:
			goto edges
		}
	}

edges:
	for {
		select {
		case <-responses.P2PConnectionMessageChannel:
		default:
			goto alerts
		}
	}

alerts:
	for {
		select {
		case <-responses.NewAlertChannel:
		default:
			return
		}
	}
}

func receiveRemovalRequest(t *testing.T) structs.RemoveInternalConnectionMessage {
	t.Helper()
	select {
	case request := <-RemoveInternalConnectionChannel:
		return request
	case <-time.After(time.Second):
		t.Fatal("expected a P2P removal request")
		return structs.RemoveInternalConnectionMessage{}
	}
}

func receiveRemovalEdge(t *testing.T) structs.P2PConnectionMessage {
	t.Helper()
	select {
	case edge := <-responses.P2PConnectionMessageChannel:
		return edge
	default:
		t.Fatal("expected a P2P removal edge")
		return structs.P2PConnectionMessage{}
	}
}

func assertNoRemovalRequest(t *testing.T) {
	t.Helper()
	select {
	case request := <-RemoveInternalConnectionChannel:
		t.Fatalf("unexpected P2P removal request: %+v", request)
	default:
	}
}

func assertNoRemovalEdge(t *testing.T) {
	t.Helper()
	select {
	case edge := <-responses.P2PConnectionMessageChannel:
		t.Fatalf("unexpected P2P removal edge: %+v", edge)
	default:
	}
}

func assertRemovalEdge(t *testing.T, edge structs.P2PConnectionMessage, profileName string) {
	t.Helper()
	if edge.Source != testSourceUUID {
		t.Fatalf("unexpected edge source: %q", edge.Source)
	}
	if edge.Destination != testCallbackUUID {
		t.Fatalf("unexpected edge destination: %q", edge.Destination)
	}
	if edge.Action != "remove" {
		t.Fatalf("unexpected edge action: %q", edge.Action)
	}
	if edge.C2ProfileName != profileName {
		t.Fatalf("unexpected edge profile: %q", edge.C2ProfileName)
	}
}

func tcpConnectionExists(connectionUUID string) bool {
	internalTCPConnectionMutex.RLock()
	defer internalTCPConnectionMutex.RUnlock()
	_, ok := internalTCPConnections[connectionUUID]
	return ok
}

func webshellConnectionExists(connectionUUID string) bool {
	internalWebshellConnectionMutex.RLock()
	defer internalWebshellConnectionMutex.RUnlock()
	_, ok := internalWebshellConnections[connectionUUID]
	return ok
}

func TestHandleRemoveInternalP2PConnectionResolvesCanonicalUUID(t *testing.T) {
	resetP2PTestState(t)

	addInternalConnectionUUID(testTemporaryUUID, testIntermediateUUID)
	addInternalConnectionUUID(testIntermediateUUID, testCallbackUUID)
	internalWebshellConnectionMutex.Lock()
	internalWebshellConnections[testCallbackUUID] = Arguments{TargetUUID: testCallbackUUID}
	internalWebshellConnectionMutex.Unlock()

	removed := handleRemoveInternalP2PConnection(structs.RemoveInternalConnectionMessage{
		ConnectionUUID: testTemporaryUUID,
		C2ProfileName:  "webshell",
	})
	if !removed {
		t.Fatal("expected the mapped webshell connection to be removed")
	}
	if webshellConnectionExists(testCallbackUUID) {
		t.Fatal("webshell connection still exists after removal")
	}
	assertRemovalEdge(t, receiveRemovalEdge(t), "webshell")

	if handleRemoveInternalP2PConnection(structs.RemoveInternalConnectionMessage{
		ConnectionUUID: testTemporaryUUID,
		C2ProfileName:  "webshell",
	}) {
		t.Fatal("duplicate removal unexpectedly succeeded")
	}
	assertNoRemovalEdge(t)
}

func TestManualTCPRemovalReportsExactlyOnce(t *testing.T) {
	resetP2PTestState(t)

	connection, peer := net.Pipe()
	defer peer.Close()
	internalTCPConnectionMutex.Lock()
	internalTCPConnections[testCallbackUUID] = &connection
	internalTCPConnectionMutex.Unlock()

	if !handleRemoveInternalP2PConnection(structs.RemoveInternalConnectionMessage{
		ConnectionUUID: testCallbackUUID,
		C2ProfileName:  "tcp",
	}) {
		t.Fatal("expected manual TCP removal to succeed")
	}
	assertRemovalEdge(t, receiveRemovalEdge(t), "tcp")
	assertNoRemovalRequest(t)
	assertNoRemovalEdge(t)
}

func TestTCPReadFailureQueuesCanonicalRemoval(t *testing.T) {
	resetP2PTestState(t)

	connection, peer := net.Pipe()
	internalTCPConnectionMutex.Lock()
	internalTCPConnections[testCallbackUUID] = &connection
	internalTCPConnectionMutex.Unlock()
	addInternalConnectionUUID(testTemporaryUUID, testCallbackUUID)

	go (poseidonTCP{}).readFromInternalTCPConnections(&connection, testTemporaryUUID)
	peer.Close()

	request := receiveRemovalRequest(t)
	if request.ConnectionUUID != testTemporaryUUID || request.C2ProfileName != "tcp" {
		t.Fatalf("unexpected TCP read-failure request: %+v", request)
	}
	if !tcpConnectionExists(testCallbackUUID) {
		t.Fatal("TCP connection was removed before centralized teardown")
	}
	if !handleRemoveInternalP2PConnection(request) {
		t.Fatal("expected centralized TCP read-failure removal to succeed")
	}
	if tcpConnectionExists(testCallbackUUID) {
		t.Fatal("TCP connection still exists after centralized teardown")
	}
	assertRemovalEdge(t, receiveRemovalEdge(t), "tcp")
	assertNoRemovalRequest(t)
	assertNoRemovalEdge(t)
}

func TestTCPWriteFailureQueuesCanonicalRemoval(t *testing.T) {
	resetP2PTestState(t)

	connection, peer := net.Pipe()
	peer.Close()
	internalTCPConnectionMutex.Lock()
	internalTCPConnections[testTemporaryUUID] = &connection
	internalTCPConnectionMutex.Unlock()

	(poseidonTCP{}).ProcessIngressMessageForP2P(&structs.DelegateMessage{
		UUID:          testTemporaryUUID,
		MythicUUID:    testCallbackUUID,
		Message:       "message that cannot be written to the closed TCP connection",
		C2ProfileName: "tcp",
	})

	request := receiveRemovalRequest(t)
	if request.ConnectionUUID != testTemporaryUUID || request.C2ProfileName != "tcp" {
		t.Fatalf("unexpected TCP write-failure request: %+v", request)
	}
	if !tcpConnectionExists(testCallbackUUID) {
		t.Fatal("TCP connection was removed before centralized teardown")
	}
	if !handleRemoveInternalP2PConnection(request) {
		t.Fatal("expected centralized TCP write-failure removal to succeed")
	}
	assertRemovalEdge(t, receiveRemovalEdge(t), "tcp")
	assertNoRemovalRequest(t)
	assertNoRemovalEdge(t)
}

func TestWebshellNetworkFailuresQueueRemoval(t *testing.T) {
	tests := []struct {
		name      string
		transport roundTripFunc
	}{
		{
			name: "request failure",
			transport: func(*http.Request) (*http.Response, error) {
				return nil, errors.New("connection refused")
			},
		},
		{
			name: "response body failure",
			transport: func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       failingReadCloser{},
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetP2PTestState(t)
			client = &http.Client{Transport: test.transport}
			connection := Arguments{
				URL:        "http://example.test/webshell",
				QueryParam: "q",
				TargetUUID: testCallbackUUID,
			}
			internalWebshellConnectionMutex.Lock()
			internalWebshellConnections[testCallbackUUID] = connection
			internalWebshellConnectionMutex.Unlock()

			(webshell{}).ProcessIngressMessageForP2P(&structs.DelegateMessage{
				UUID:          testCallbackUUID,
				Message:       strings.Repeat("a", 51),
				C2ProfileName: "webshell",
			})

			request := receiveRemovalRequest(t)
			if request.ConnectionUUID != testCallbackUUID || request.C2ProfileName != "webshell" {
				t.Fatalf("unexpected webshell failure request: %+v", request)
			}
			if !webshellConnectionExists(testCallbackUUID) {
				t.Fatal("webshell connection was removed before centralized teardown")
			}
			if !handleRemoveInternalP2PConnection(request) {
				t.Fatal("expected centralized webshell removal to succeed")
			}
			assertRemovalEdge(t, receiveRemovalEdge(t), "webshell")
			assertNoRemovalRequest(t)
			assertNoRemovalEdge(t)
		})
	}
}

func TestWebshellProtocolFailuresKeepConnection(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{name: "non-200 response", statusCode: http.StatusNotFound, body: "not found"},
		{name: "malformed XML", statusCode: http.StatusOK, body: "<span"},
		{name: "unexpected response ID", statusCode: http.StatusOK, body: `<span id="other">YQ==</span>`},
		{name: "empty response", statusCode: http.StatusOK, body: `<span id="task_response"></span>`},
		{name: "invalid base64", statusCode: http.StatusOK, body: `<span id="task_response">%%%</span>`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resetP2PTestState(t)
			client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.statusCode,
					Body:       io.NopCloser(strings.NewReader(test.body)),
					Header:     make(http.Header),
				}, nil
			})}
			connection := Arguments{
				URL:        "http://example.test/webshell",
				QueryParam: "q",
				TargetUUID: testCallbackUUID,
			}
			internalWebshellConnectionMutex.Lock()
			internalWebshellConnections[testCallbackUUID] = connection
			internalWebshellConnectionMutex.Unlock()

			(webshell{}).ProcessIngressMessageForP2P(&structs.DelegateMessage{
				UUID:          testCallbackUUID,
				Message:       strings.Repeat("a", 51),
				C2ProfileName: "webshell",
			})

			if !webshellConnectionExists(testCallbackUUID) {
				t.Fatal("webshell protocol failure removed the connection")
			}
			assertNoRemovalRequest(t)
			assertNoRemovalEdge(t)
		})
	}
}
