// Copyright 2016 Square Inc.
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

package main

import "fmt"
import (
	"time"

	klog "github.com/square/keywhiz-fs/log"

	"net/url"

	"path/filepath"

	"github.com/square/go-sq-metrics"
)

type syncerEntry struct {
	Client
	ClientConfig
}

// A Syncer manages a collection of clients, handling downloads and writing out updated secrets.
// Construct one using the NewSyncer and AddClient functions
type Syncer struct {
	clients          map[string]syncerEntry
	defaultOwnership Ownership
}

// NewSyncer instantiates the main stateful object in Keysync.
func NewSyncer(configs map[string]ClientConfig, serverURL *url.URL, caFile *string, defaultUser, defaultGroup string, debug bool, metricsHandle *sqmetrics.SquareMetrics) Syncer {
	syncer := Syncer{clients: map[string]syncerEntry{}}
	for name, config := range configs {
		fmt.Printf("Client %s: %v\n", name, config)
		klogConfig := klog.Config{
			Debug:      debug,
			Syslog:     false,
			Mountpoint: name,
		}
		client := NewClient(config.Cert, config.Key, *caFile, serverURL, time.Minute, klogConfig, metricsHandle)
		user := config.User
		group := config.Group
		if user == "" {
			user = defaultUser
		}
		if group == "" {
			group = defaultGroup
		}
		NewOwnership(user, group)
		syncer.clients[name] = syncerEntry{Client: client, ClientConfig: config}
	}
	return syncer
}

// RunNow runs the syncer once, for all clients, without sleeps.
func (s *Syncer) RunNow() error {
	// TODO: Ensure tmpfs is set up properly so we don't accidentally write to disks.
	for name, entry := range s.clients {
		fmt.Printf("Updating %s", name)
		client := entry.Client
		secrets, ok := client.SecretList()
		if !ok {
			//SecretList logged the error, continue on
			continue
		}
		for _, secretMetadata := range secrets {
			// TODO: Optimizations to avoid needlessly fetching secrets
			secret, err := client.Secret(secretMetadata.Name)
			if err != nil {
				// client.Secret logged the error, continue on
				continue
			}
			// We split out the filename to prevent a maliciously-named secret from
			// writing outside of the intended secrets directory.
			_, filename := filepath.Split(secret.Name)
			name := filepath.Join(entry.Mountpoint, filename)
			atomicWrite(name, secret, s.defaultOwnership)
		}
	}
	return nil
}
