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
type ProtocolHandler func([]byte, interface{}) (map[string]interface{}, error)

type Protocol struct {
	handlers map[string]ProtocolHandler
}

func NewProtocol() *Protocol {
	return &Protocol{
		handlers: make(map[string]ProtocolHandler),
	}
}

func (proto *Protocol) Handle(cmd string, handler ProtocolHandler) {
	proto.handlers[cmd] = handler
}

type clientCtx struct {
	conn net.Conn

	decoder *json.Decoder
	encoder *json.Encoder

	userData interface{}
}

func (proto *Protocol) handleRequest(ctx *clientCtx, req *Request) (*Response, error) {
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
	} else {
		resp = &Response{
			Id:      req.Id,
			Success: true,
			Data:    respMap,
		}
	}

	return resp, nil
}

func (proto *Protocol) Serve(conn net.Conn, userData interface{}) {
	ctx := &clientCtx{
		conn:     conn,
		userData: userData,
		decoder:  json.NewDecoder(conn),
		encoder:  json.NewEncoder(conn),
	}

	for {
		// Parse a request.
		req := Request{}
		err := ctx.decoder.Decode(&req)
		if err == io.EOF {
			return
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "couldn't decode request: %s", err)
			return
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
