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

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	mb ".."
)

var (
	socket  string
	message = "Hello world"
)

func init() {
	flag.StringVar(&socket, "socket", "", "Path of MultiBinder socket")
	flag.StringVar(&message, "message", "Hello world", "Homepage message")
}

func main() {
	flag.Parse()
	client, err := mb.NewMultiBinderClient(socket)
	if err != nil {
		log.Fatalf("NewMultiBinderClient failed: %#v", err)
	}

	addresses := []mb.BindAddress{
		mb.BindAddress{Address: "0.0.0.0", Port: 8083},
		mb.BindAddress{Address: "0.0.0.0", Port: 8084},
	}

	for _, addr := range addresses {
		fd, err := client.Bind(addr)
		if err != nil {
			log.Fatalf("Bind failed: %#v", err)
		}

		f := os.NewFile(uintptr(fd), fmt.Sprintf("socket-fd-%d", fd))
		l, err := net.FileListener(f)
		if err != nil {
			log.Fatalf("Failed to listen on fd %d", fd)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
			// The "/" pattern matches everything, so we need to check
			// that we're at the root here.
			if req.URL.Path != "/" {
				http.NotFound(w, req)
				return
			}
			fmt.Fprintf(w, "%s (%d)\n", message, fd)
		})

		s := &http.Server{}
		s.Handler = mux
		go s.Serve(l)
		fmt.Printf("Listening for port %d on fd %d\n", addr.Port, fd)
	}

	time.Sleep(time.Hour)
}
