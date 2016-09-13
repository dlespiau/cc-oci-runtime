// Copyright (c) 2016 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type Request struct {
	Id      uint            `json:"id"`
	Command string          `json:"command"`
	Data    json.RawMessage `json:"data"`
}

type Response struct {
	Id      uint                   `json:"id"`
	Success bool                   `json:"success"`
	Error   string                 `json:"error,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// XXX: could do with its own package to remove that ugly namespacing
type JsonProtoHandler func([]byte, interface{}) (map[string]interface{}, error)

type JsonProto struct {
	handlers map[string]JsonProtoHandler
}

func NewJsonProto() *JsonProto {
	return &JsonProto{
		handlers: make(map[string]JsonProtoHandler),
	}
}

func (proto *JsonProto) Handle(cmd string, handler JsonProtoHandler) {
	proto.handlers[cmd] = handler
}

type clientCtx struct {
	conn net.Conn

	reader  *bufio.Reader
	encoder *json.Encoder

	userData interface{}
}

func (proto *JsonProto) handleRequest(ctx *clientCtx, req *Request) (*Response, error) {
	handler, ok := proto.handlers[req.Command]
	if !ok {
		return nil, fmt.Errorf("couldn't find command '%s'", req.Command)
	}

	var resp *Response

	if respMap, err := handler(req.Data, ctx.userData); err != nil {
		resp = &Response{
			Id:      req.Id,
			Success: false,
			Error:   err.Error(),
			Data:    respMap,
		}
	} else if respMap != nil {
		resp = &Response{
			Id:      req.Id,
			Success: true,
			Data:    respMap,
		}
	} else {
		resp = &Response{
			Id:      req.Id,
			Success: true,
		}
	}

	return resp, nil
}

func (proto *JsonProto) Serve(conn net.Conn, userData interface{}) {
	ctx := &clientCtx{
		conn:     conn,
		userData: userData,
		reader:   bufio.NewReader(conn),
		encoder:  json.NewEncoder(conn),
	}

	for {
		// Read one line.
		line, err := ctx.reader.ReadBytes('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "read error: %s", err)
			return
		}

		// Parse the request.
		req := Request{}
		if err = json.Unmarshal(line, &req); err != nil {
			fmt.Fprintf(os.Stderr, "malformed request, ignoring: '%v\n'", line)
			continue
		}

		// Execute the corresponding handler
		resp, err := proto.handleRequest(ctx, &req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could handle request: %s\n", err)
		}

		// The command didn't generate any response, next!
		if resp == nil {
			continue
		}

		// Send the response back to the client.
		if err = ctx.encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "couldn't encode response: %s\n", err)
		}
	}
}
