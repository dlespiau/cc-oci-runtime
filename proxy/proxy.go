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
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
)

// Main struct holding the proxy state
type proxy struct {
	// Protect concurrent accesses from separate client goroutines to this
	// structure fields
	sync.Mutex

	// client socket
	listener net.Listener

	// vms are hashed by their containerID
	vms map[string]*vm
}

// Represents a client, either a cc-oci-runtime or cc-shim process having
// opened a socket to the proxy
type client struct {
	proxy *proxy
	vm    *vm

	conn net.Conn
}

// "hello"
type helloCmd struct {
	ContainerId string `json:containerId`
	CtlSerial   string `json:ctlSerial`
	IoSerial    string `json:ioSerial`
}

func helloHandler(data []byte, userData interface{}) (map[string]interface{}, error) {
	client := userData.(*client)
	hello := helloCmd{}

	if err := json.Unmarshal(data, &hello); err != nil {
		return nil, err
	}

	if hello.ContainerId == "" || hello.CtlSerial == "" || hello.IoSerial == "" {
		return nil, errors.New("malformed hello command")
	}

	proxy := client.proxy
	proxy.Lock()
	if _, ok := proxy.vms[hello.ContainerId]; ok {

		proxy.Unlock()
		return nil, errors.New("container already registered")
	}
	vm := NewVM(hello.ContainerId, hello.CtlSerial, hello.IoSerial)
	client.vm = vm
	proxy.vms[hello.ContainerId] = vm
	proxy.Unlock()

	if err := vm.Connect(); err != nil {
		proxy.Lock()
		delete(proxy.vms, hello.ContainerId)
		proxy.Unlock()
		return nil, err
	}

	return nil, nil
}

func NewProxy() *proxy {
	return &proxy{
		vms: make(map[string]*vm),
	}
}

func (proxy *proxy) Init() error {
	fds := listenFds()

	if len(fds) != 1 {
		return fmt.Errorf("couldn't find activated socket (%d)", len(fds))
	}

	fd := fds[0]
	l, err := net.FileListener(fd)
	if err != nil {
		return fmt.Errorf("couldn't listen on socket: %s", err.Error())
	}

	proxy.listener = l

	return nil
}

func (proxy *proxy) Serve() {

	// Define the client (runtime/shim) <-> proxy protocol
	proto := NewJsonProto()
	proto.Handle("hello", helloHandler)

	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "couldn't accept connection:", err)
			continue
		}

		newClient := &client{
			proxy: proxy,
			conn:  conn,
		}

		go proto.Serve(conn, newClient)
	}
}

func main() {
	proxy := NewProxy()
	if err := proxy.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "init:", err.Error())
		os.Exit(1)
	}
	proxy.Serve()
}
