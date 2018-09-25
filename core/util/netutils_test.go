/*
Copyright 2016 Medcl (m AT medcl.net)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"github.com/stretchr/testify/assert"
	"net"
	"strconv"
	"testing"
)

func TestTestPort(t *testing.T) {
	port := 42122
	res := TestPort(port)
	assert.Equal(t, true, res)

	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(port))

	res = TestPort(port)
	assert.Equal(t, false, res)
	ln.Close()
}

func TestGetAvailablePort(t *testing.T) {
	port := 42123
	res := TestPort(port)
	assert.Equal(t, true, res)

	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(port))

	p1 := GetAvailablePort("", port)
	assert.Equal(t, 42124, p1)
	ln.Close()
}

func TestAutoGetAddress(t *testing.T) {
	port := 42123
	res := TestPort(port)
	assert.Equal(t, true, res)

	ln, _ := net.Listen("tcp", ":"+strconv.Itoa(port))

	var p1 string

	p1 = AutoGetAddress(":42123")
	assert.Equal(t, ":42124", p1)
	ln.Close()

	ln, _ = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))

	p1 = AutoGetAddress("127.0.0.1:42123")
	assert.Equal(t, "127.0.0.1:42124", p1)
	ln.Close()

}

func TestGetValidAddress(t *testing.T) {
	addr := ":8001"
	addr = GetValidAddress(addr)
	assert.Equal(t, "127.0.0.1:8001", addr)
}
