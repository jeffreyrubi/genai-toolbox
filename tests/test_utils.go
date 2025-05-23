// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tests

import (
	"net"
	"reflect"
	"unsafe"

	"github.com/googleapis/genai-toolbox/internal/server"
)

// GetServerListenerAddr uses reflection to safely get the listener address from a Server.
// This is a test utility function only and should not be used in production code.
func GetServerListenerAddr(srv *server.Server) net.Addr {
	// Use reflection to access the unexported listener field
	val := reflect.ValueOf(srv).Elem()
	listenerField := val.FieldByName("listener")
	
	// Check if the field exists and is not nil
	if !listenerField.IsValid() || listenerField.IsNil() {
		return nil
	}
	
	// Get the listener value using unsafe pointer
	listenerPtr := unsafe.Pointer(listenerField.UnsafeAddr())
	listener := *(*net.Listener)(listenerPtr)
	
	if listener == nil {
		return nil
	}
	
	return listener.Addr()
}
