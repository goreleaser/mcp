package main

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func TestFrameFilter(t *testing.T) {
	ping := `{"jsonrpc":"2.0","id":1,"method":"ping"}`
	tests := []struct {
		name    string
		in      string
		forward string
		code    float64
	}{
		{name: "request", in: ping, forward: ping},
		{name: "notification", in: `{"jsonrpc":"2.0","method":"ping"}`, forward: `{"jsonrpc":"2.0","method":"ping"}`},
		{name: "response", in: `{"jsonrpc":"2.0","id":1,"result":{}}`, forward: `{"jsonrpc":"2.0","id":1,"result":{}}`},
		{name: "batch", in: "[" + ping + "]", forward: "[" + ping + "]"},
		{name: "unknown method", in: `{"jsonrpc":"2.0","id":1,"method":"nope"}`, forward: `{"jsonrpc":"2.0","id":1,"method":"nope"}`},
		{name: "blank lines", in: "\n  \n" + ping, forward: ping},
		{name: "syntax error", in: `{"jsonrpc": "2.0", this is not json`, code: jsonrpc.CodeParseError},
		{name: "truncated", in: `{"jsonrpc":"2.0"`, code: jsonrpc.CodeParseError},
		{name: "number", in: `5`, code: jsonrpc.CodeInvalidRequest},
		{name: "null", in: `null`, code: jsonrpc.CodeInvalidRequest},
		{name: "empty batch", in: `[]`, code: jsonrpc.CodeInvalidRequest},
		{name: "batch with bad member", in: `[` + ping + `,5]`, code: jsonrpc.CodeInvalidRequest},
		{name: "no method", in: `{"jsonrpc":"2.0"}`, code: jsonrpc.CodeInvalidRequest},
		{name: "wrong version", in: `{"jsonrpc":"1.0","id":1,"method":"ping"}`, code: jsonrpc.CodeInvalidRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var errs strings.Builder
			filter := newFrameFilter(io.NopCloser(strings.NewReader(tt.in+"\n")), &errs)
			forwarded, err := io.ReadAll(filter)
			require.NoError(t, err)
			require.NoError(t, filter.Close())

			if tt.forward != "" {
				require.Equal(t, tt.forward+"\n", string(forwarded))
				require.Empty(t, errs.String())
				return
			}

			require.Empty(t, string(forwarded))
			var reply map[string]any
			require.NoError(t, json.Unmarshal([]byte(errs.String()), &reply))
			require.Equal(t, "2.0", reply["jsonrpc"])
			require.Nil(t, reply["id"])
			require.Equal(t, tt.code, reply["error"].(map[string]any)["code"])
		})
	}
}

// TestSessionSurvivesBadFrames guards against a client ending the session with
// a single frame the SDK cannot decode.
func TestSessionSurvivesBadFrames(t *testing.T) {
	inRead, inWrite := io.Pipe()
	outRead, outWrite := io.Pipe()
	out := &syncWriter{w: outWrite}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v1"}, nil)
	session, err := server.Connect(t.Context(), &mcp.IOTransport{
		Reader: newFrameFilter(inRead, out),
		Writer: out,
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = session.Close() })

	replies := make(chan map[string]any)
	go func() {
		defer close(replies)
		scanner := bufio.NewScanner(outRead)
		for scanner.Scan() {
			var reply map[string]any
			if err := json.Unmarshal(scanner.Bytes(), &reply); err != nil {
				return
			}
			replies <- reply
		}
	}()

	send := func(frame string) {
		_, err := io.WriteString(inWrite, frame+"\n")
		require.NoError(t, err)
	}
	next := func() map[string]any {
		select {
		case reply := <-replies:
			return reply
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for a reply")
			return nil
		}
	}

	send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"v1"}}}`)
	require.Equal(t, float64(1), next()["id"])
	send(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)

	for _, frame := range []string{
		`{"jsonrpc": "2.0", this is not json`,
		`5`,
		`null`,
		`[]`,
		`{"jsonrpc":"2.0"}`,
		`{"jsonrpc":"1.0","id":9,"method":"ping"}`,
	} {
		send(frame)
		reply := next()
		require.Nil(t, reply["id"], frame)
		require.NotNil(t, reply["error"], frame)
	}

	send(`{"jsonrpc":"2.0","id":2,"method":"ping"}`)
	reply := next()
	require.Equal(t, float64(2), reply["id"])
	require.NotNil(t, reply["result"])
}

func TestErrorFrameEndsWithNewline(t *testing.T) {
	for _, frame := range [][]byte{parseErrorFrame, invalidRequestFrame} {
		require.Equal(t, byte('\n'), frame[len(frame)-1])
		require.True(t, json.Valid(frame))
	}
}
