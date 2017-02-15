// Copyright 2017 Square Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package keysync

import (
	"crypto/tls"
	"io/ioutil"
)

// fixture fully reads test data from a file in the fixtures/ subdirectory.
func fixture(file string) (content []byte) {
	content, err := ioutil.ReadFile("fixtures/" + file)
	panicOnError(err)
	return
}

// Load the file with cert & private key into a tls.Config
func testCerts(file string) (config *tls.Config) {
	config = new(tls.Config)
	cert, err := tls.LoadX509KeyPair(file, file)
	panicOnError(err)

	config.Certificates = []tls.Certificate{cert}

	return config
}

// Helper function to panic on error
func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
