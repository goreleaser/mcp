package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/jsonrpc"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//nolint:gochecknoglobals
var (
	parseErrorFrame     = errorFrame(jsonrpc.CodeParseError, "Parse error")
	invalidRequestFrame = errorFrame(jsonrpc.CodeInvalidRequest, "Invalid Request")
)

// errorFrame builds a JSON-RPC error response. Its identifier is null, as the
// frame it answers could not be decoded.
func errorFrame(code int, message string) []byte {
	return fmt.Appendf(nil, `{"jsonrpc":"2.0","id":null,"error":{"code":%d,"message":%q}}`+"\n", code, message)
}

// stdioTransport returns a transport over stdin and stdout that answers
// undecodable frames with a JSON-RPC error instead of ending the session.
//
// The SDK stops reading at the first frame it cannot decode, so one garbled
// message from a client ends the session.
// See https://github.com/modelcontextprotocol/go-sdk/issues/1209.
func stdioTransport() *mcp.IOTransport {
	out := &syncWriter{w: os.Stdout}
	return &mcp.IOTransport{
		Reader: newFrameFilter(os.Stdin, out),
		Writer: out,
	}
}

// frameFilter forwards newline-delimited JSON-RPC frames, replacing the ones
// that cannot be decoded with an error frame written to errs.
type frameFilter struct {
	src     io.ReadCloser
	in      *bufio.Reader
	errs    io.Writer
	pending []byte
}

func newFrameFilter(src io.ReadCloser, errs io.Writer) *frameFilter {
	return &frameFilter{src: src, in: bufio.NewReader(src), errs: errs}
}

func (f *frameFilter) Read(p []byte) (int, error) {
	for len(f.pending) == 0 {
		line, err := f.in.ReadBytes('\n')
		if frame := bytes.TrimSpace(line); len(frame) > 0 {
			if reply := checkFrame(frame); reply != nil {
				if _, werr := f.errs.Write(reply); werr != nil {
					return 0, werr
				}
			} else {
				f.pending = line
			}
		}
		if err != nil {
			if len(f.pending) == 0 {
				return 0, err
			}
			break
		}
	}
	n := copy(p, f.pending)
	f.pending = f.pending[n:]
	return n, nil
}

func (f *frameFilter) Close() error { return f.src.Close() }

// checkFrame returns the error frame to reply with, or nil if the frame can be
// decoded.
func checkFrame(frame []byte) []byte {
	if !json.Valid(frame) {
		return parseErrorFrame
	}
	msgs := []json.RawMessage{frame}
	if frame[0] == '[' {
		// Decode into a fresh slice: unmarshaling into msgs would reuse, and
		// thus overwrite, the storage of frame.
		var batch []json.RawMessage
		if err := json.Unmarshal(frame, &batch); err != nil || len(batch) == 0 {
			return invalidRequestFrame
		}
		msgs = batch
	}
	for _, msg := range msgs {
		if _, err := jsonrpc.DecodeMessage(msg); err != nil {
			return invalidRequestFrame
		}
	}
	return nil
}

// syncWriter serializes writes, so error frames cannot interleave with the
// frames the SDK writes.
type syncWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *syncWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

func (*syncWriter) Close() error { return nil }
