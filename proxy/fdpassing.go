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
	"errors"
	"fmt"
	"net"
	"syscall"
)

// FD passing support.
//
// fds are passed through the out of band data of AF_UNIX sockets. Because it's
// not possible to send out of band data without actual data to write, we write
// a dummy byte, 'F', to the socket when passing the fd.

var fileTag []byte = []byte{'F'}

func WriteFd(c *net.UnixConn, fd int) error {
	rights := syscall.UnixRights(fd)
	_, _, err := c.WriteMsgUnix(fileTag, rights, nil)
	return err
}

func ReadFd(c *net.UnixConn) (int, error) {
	oob := make([]byte, 32)
	buf := make([]byte, 1)

	// Retrieve out of band data
	n, oobn, _, _, err := c.ReadMsgUnix(buf, oob)
	if oobn == 0 {
		return -1, errors.New("no out of band data read")
	}
	if n != 1 && buf[0] != 'F' {
		return -1, errors.New("couldn't read fd passing tag")
	}

	// Parse the fd out of the out of band data
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return -1, err
	}
	if len(scms) != 1 {
		return -1, fmt.Errorf("unexpected number of control messages (%d)", len(scms))
	}
	scm := scms[0]
	fds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		return -1, err
	}
	if len(fds) != 1 {
		return -1, fmt.Errorf("unexpected number of fds (%d)", len(fds))
	}
	return fds[0], nil
}
