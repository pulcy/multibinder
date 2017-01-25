//    Copyright 2017 Ewout Prangsma
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package multibinder

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"syscall"
	"time"
)

type BindAddress struct {
	Address string `json:"address"` // IP to listen on (e.g. '0.0.0.0')
	Port    int    `json:"port"`    // Port to listen on (e.g. 8080)
}

type MultiBinderClient struct {
	socket string
	id     string
}

// NewMultiBinderClient creates a new client for multibinder on the given socket path.
func NewMultiBinderClient(socket string) (*MultiBinderClient, error) {
	if socket == "" {
		return nil, fmt.Errorf("Socket not set")
	}
	// Create random ID
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return &MultiBinderClient{
		socket: socket,
		id:     hex.EncodeToString(b),
	}, nil
}

type bindRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	ID      string        `json:"id"`
	Params  []BindAddress `json:"params"`
}

type bindResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Bind requests a binding for the given address.
// The result is a file descriptor that can be listened on.
func (c *MultiBinderClient) Bind(address BindAddress) (int, error) {
	req := bindRequest{
		JSONRPC: "2.0",
		Method:  "bind",
		ID:      c.id,
		Params:  []BindAddress{address},
	}
	data, _ := json.Marshal(req)

	sk, err := net.DialTimeout("unix", c.socket, time.Second*10)
	if err != nil {
		return 0, err
	}
	defer sk.Close()

	unixSk := sk.(*net.UnixConn)
	if _, err := unixSk.Write(data); err != nil {
		return 0, err
	}

	buf := make([]byte, 256)
	oob := make([]byte, 256)
	bufn, oobn, _, _, err := unixSk.ReadMsgUnix(buf, oob)
	if err != nil {
		return 0, err
	}

	var resp bindResponse
	if err := json.Unmarshal(buf[:bufn], &resp); err != nil {
		return 0, err
	}
	if resp.Error.Message != "" {
		return 0, fmt.Errorf(resp.Error.Message)
	}

	scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return 0, fmt.Errorf("ParseSocketControlMessage: %v", err)
	}
	if len(scms) != 1 {
		return 0, fmt.Errorf("expected 1 SocketControlMessage; got scms = %#v", scms)
	}
	scm := scms[0]
	gotFds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		return 0, fmt.Errorf("syscall.ParseUnixRights: %v", err)
	}
	if len(gotFds) != 1 {
		return 0, fmt.Errorf("wanted 1 fd; got %d", len(gotFds))
	}

	return gotFds[0], nil
}
