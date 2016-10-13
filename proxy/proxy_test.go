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
	"net"
	"sync"
	"testing"

	"github.com/01org/cc-oci-runtime/proxy/api"
	"github.com/01org/cc-oci-runtime/tests/mock"

	hyper "github.com/hyperhq/runv/hyperstart/api/json"
	"github.com/stretchr/testify/assert"
)

type testRig struct {
	t  *testing.T
	wg sync.WaitGroup

	// hyperstart mocking
	Hyperstart      *mock.Hyperstart
	ctlPath, ioPath string

	// proxy (the object we'll test)
	Proxy     *proxy
	protocol  *Protocol
	proxyConn net.Conn // socket used by proxy to communicate with Client

	// client
	Client *api.Client
}

func newTestRig(t *testing.T, proto *Protocol) *testRig {
	return &testRig{
		t:        t,
		protocol: proto,
	}
}

func (rig *testRig) Start() {
	// client <-> proxy connection
	clientConn, proxyConn, err := Socketpair()
	assert.Nil(rig.t, err)
	rig.Client = api.NewClient(clientConn)

	// Start hyperstart go routine
	rig.Hyperstart = mock.NewHyperstart(rig.t)
	rig.Hyperstart.Start()

	// Start proxy main go routine
	rig.Proxy = NewProxy()
	rig.proxyConn = proxyConn
	rig.wg.Add(1)
	go func() {
		rig.Proxy.serveNewClient(rig.protocol, proxyConn)
		rig.wg.Done()
	}()
}

func (rig *testRig) Stop() {
	rig.Client.Close()
	rig.proxyConn.Close()
	rig.Hyperstart.Stop()
	rig.wg.Wait()
}

const testContainerId = "0987654321"

func TestHello(t *testing.T) {
	proto := NewProtocol()
	proto.Handle("hello", helloHandler)

	rig := newTestRig(t, proto)
	rig.Start()

	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()
	err := rig.Client.Hello(testContainerId, ctlSocketPath, ioSocketPath)
	assert.Nil(t, err)

	// Hello should register a new vm object
	proxy := rig.Proxy
	proxy.Lock()
	vm := proxy.vms[testContainerId]
	proxy.Unlock()

	assert.NotNil(t, vm)
	assert.Equal(t, testContainerId, vm.containerId)

	// A new Hello message with the same containerId should error out
	err = rig.Client.Hello(testContainerId, "fooCtl", "fooIo")
	assert.NotNil(t, err)

	// This test shouldn't send anything to hyperstart
	msgs := rig.Hyperstart.GetLastMessages()
	assert.Equal(t, 0, len(msgs))

	rig.Stop()
}

func TestBye(t *testing.T) {
	proto := NewProtocol()
	proto.Handle("hello", helloHandler)
	proto.Handle("bye", byeHandler)

	rig := newTestRig(t, proto)
	rig.Start()

	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()
	err := rig.Client.Hello(testContainerId, ctlSocketPath, ioSocketPath)
	assert.Nil(t, err)

	err = rig.Client.Bye()
	assert.Nil(t, err)

	// Bye should unregister the vm object
	proxy := rig.Proxy
	proxy.Lock()
	vm := proxy.vms[testContainerId]
	proxy.Unlock()
	assert.Nil(t, vm)

	// Bye while not attached should return an error
	err = rig.Client.Bye()
	assert.NotNil(t, err)

	// This test shouldn't send anything to hyperstart
	msgs := rig.Hyperstart.GetLastMessages()
	assert.Equal(t, 0, len(msgs))

	rig.Stop()
}

func TestAttach(t *testing.T) {
	proto := NewProtocol()
	proto.Handle("hello", helloHandler)
	proto.Handle("attach", attachHandler)
	proto.Handle("bye", byeHandler)

	rig := newTestRig(t, proto)
	rig.Start()

	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()
	err := rig.Client.Hello(testContainerId, ctlSocketPath, ioSocketPath)
	assert.Nil(t, err)

	// Attaching to an unknown VM should return an error
	err = rig.Client.Attach("foo")
	assert.NotNil(t, err)

	// Attaching to an existing VM should work. To test we are effectively
	// attached, we issue a bye that would error out if not attatched.
	err = rig.Client.Attach(testContainerId)
	assert.Nil(t, err)
	err = rig.Client.Bye()
	assert.Nil(t, err)

	// This test shouldn't send anything with hyperstart
	msgs := rig.Hyperstart.GetLastMessages()
	assert.Equal(t, 0, len(msgs))

	rig.Stop()
}

func TestHyperPing(t *testing.T) {
	proto := NewProtocol()
	proto.Handle("hello", helloHandler)
	proto.Handle("hyper", hyperHandler)

	rig := newTestRig(t, proto)
	rig.Start()

	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()
	err := rig.Client.Hello(testContainerId, ctlSocketPath, ioSocketPath)
	assert.Nil(t, err)

	// Send ping and verify we have indeed received the message on the
	// hyperstart side. Ping is somewhat interesting because it's a case of
	// an hyper message without data.
	err = rig.Client.Hyper("ping", nil)
	assert.Nil(t, err)

	msgs := rig.Hyperstart.GetLastMessages()
	assert.Equal(t, 1, len(msgs))

	msg := msgs[0]
	assert.Equal(t, hyper.INIT_PING, int(msg.Code))
	assert.Equal(t, 0, len(msg.Message))

	rig.Stop()
}

func TestHyperStartpod(t *testing.T) {
	proto := NewProtocol()
	proto.Handle("hello", helloHandler)
	proto.Handle("hyper", hyperHandler)

	rig := newTestRig(t, proto)
	rig.Start()

	ctlSocketPath, ioSocketPath := rig.Hyperstart.GetSocketPaths()
	err := rig.Client.Hello(testContainerId, ctlSocketPath, ioSocketPath)
	assert.Nil(t, err)

	// Send startopd and verify we have indeed received the message on the
	// hyperstart side. startpod is interesting because it's a case of an
	// hyper message with JSON data.
	startpod := hyper.Pod{
		Hostname: "testhostname",
		ShareDir: "rootfs",
	}
	err = rig.Client.Hyper("startpod", &startpod)
	assert.Nil(t, err)

	msgs := rig.Hyperstart.GetLastMessages()
	assert.Equal(t, 1, len(msgs))

	msg := msgs[0]
	assert.Equal(t, hyper.INIT_STARTPOD, int(msg.Code))
	received := hyper.Pod{}
	err = json.Unmarshal(msg.Message, &received)
	assert.Nil(t, err)
	assert.Equal(t, startpod.Hostname, received.Hostname)
	assert.Equal(t, startpod.ShareDir, received.ShareDir)

	rig.Stop()
}
