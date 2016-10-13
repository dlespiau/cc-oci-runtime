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

package api

import (
	"encoding/json"
	"errors"
	"net"
)

type Client struct {
	conn    *net.UnixConn
	encoder *json.Encoder
	decoder *json.Decoder
}

// Creates a new client object to communicate with the proxy using the
// connection conn
func NewClient(conn *net.UnixConn) *Client {
	return &Client{
		conn:    conn,
		encoder: json.NewEncoder(conn),
		decoder: json.NewDecoder(conn),
	}
}

func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) sendPayload(id string, payload interface{}) (*Response, error) {
	var err error

	req := Request{}
	req.Id = id
	if payload != nil {
		if req.Data, err = json.Marshal(payload); err != nil {
			return nil, err
		}
	}

	if err := client.encoder.Encode(&req); err != nil {
		return nil, err
	}

	resp := Response{}
	if err := client.decoder.Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func errorFromResponse(resp *Response) error {
	// We should always have an error with the response, but better safe
	// than sorry.
	if resp.Success == false {
		if resp.Error != "" {
			return errors.New(resp.Error)
		}

		return errors.New("unknown error")
	}

	return nil
}

func (client *Client) Hello(containerId, ctlSerial, ioSerial string) error {
	hello := Hello{
		ContainerId: containerId,
		CtlSerial:   ctlSerial,
		IoSerial:    ioSerial,
	}

	resp, err := client.sendPayload("hello", &hello)
	if err != nil {
		return err
	}

	return errorFromResponse(resp)
}

func (client *Client) Attach(containerId string) error {
	hello := Attach{
		ContainerId: containerId,
	}

	resp, err := client.sendPayload("attach", &hello)
	if err != nil {
		return err
	}

	return errorFromResponse(resp)
}

func (client *Client) AllocateIo(nStreams int) (uint64, error) {
	allocate := AllocateIo{
		NStreams: nStreams,
	}

	resp, err := client.sendPayload("allocateIO", &allocate)
	if err != nil {
		return 0, err
	}

	err = errorFromResponse(resp)
	if err != nil {
		return 0, err
	}

	ioBase, ok := resp.Data["ioBase"]
	if !ok {
		return 0, errors.New("no ioBase in response")
	}

	return ioBase.(uint64), nil
}

func (client *Client) Hyper(hyperName string, hyperMessage interface{}) error {
	var data []byte

	if hyperMessage != nil {
		var err error

		data, err = json.Marshal(hyperMessage)
		if err != nil {
			return err
		}
	}

	hyper := Hyper{
		HyperName: hyperName,
		Data:      data,
	}

	resp, err := client.sendPayload("hyper", &hyper)
	if err != nil {
		return err
	}

	return errorFromResponse(resp)
}

func (client *Client) Bye() error {
	resp, err := client.sendPayload("bye", nil)
	if err != nil {
		return err
	}

	return errorFromResponse(resp)
}
