// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"runtime"
	"testing"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/vmware/vic/lib/config/executor"
	"github.com/vmware/vic/lib/system"
	"github.com/vmware/vic/lib/tether"
	"github.com/vmware/vic/pkg/dio"
	"github.com/vmware/vic/pkg/trace"
	"github.com/vmware/vic/pkg/vsphere/extraconfig"
)

// Copied from lib/tether
// because there's no easy way to use test code from other packages and a separate
// package causes cyclic dependencies.
// Some modifications to deal with the change of package and attach usage

var Mocked Mocker

type Mocker struct {
	Base tether.BaseOperations

	// allow tests to tell when the tether has finished setup
	Started chan bool
	// allow tests to tell when the tether has finished
	Cleaned chan bool

	// debug output gets logged here
	LogBuffer bytes.Buffer

	// session output gets logged here
	SessionLogBuffer bytes.Buffer

	// the hostname of the system
	Hostname string
	// the ip configuration for name index networks
	IPs map[string]net.IP
	// filesystem mounts, indexed by disk label
	Mounts map[string]string

	WindowCol uint32
	WindowRow uint32
	Signal    ssh.Signal
}

// Start implements the extension method
func (t *Mocker) Start() error {
	return nil
}

// Stop implements the extension method
func (t *Mocker) Stop() error {
	return nil
}

// Reload implements the extension method
func (t *Mocker) Reload(config *tether.ExecutorConfig) error {
	// the tether has definitely finished it's startup by the time we hit this
	close(t.Started)
	return nil
}

func (t *Mocker) Setup(_ tether.Config) error {
	return nil
}

func (t *Mocker) Cleanup() error {
	close(t.Cleaned)
	return nil
}

func (t *Mocker) Log() (io.Writer, error) {
	return &t.LogBuffer, nil
}

func (t *Mocker) SessionLog(session *tether.SessionConfig) (dio.DynamicMultiWriter, dio.DynamicMultiWriter, error) {
	return dio.MultiWriter(&t.SessionLogBuffer), dio.MultiWriter(&t.SessionLogBuffer), nil
}

func (t *Mocker) HandleSessionExit(config *tether.ExecutorConfig, session *tether.SessionConfig) func() {
	// check for executor behaviour
	return func() {
		if session.ID == config.ID {
			tthr.Stop()
		}
	}
}

func (t *Mocker) ProcessEnv(env []string) []string {
	return t.Base.ProcessEnv(env)
}

// SetHostname sets both the kernel hostname and /etc/hostname to the specified string
func (t *Mocker) SetHostname(hostname string, aliases ...string) error {
	defer trace.End(trace.Begin("mocking hostname to " + hostname))

	// TODO: we could mock at a much finer granularity, only extracting the syscall
	// that would exercise the file modification paths, however it's much less generalizable
	t.Hostname = hostname
	return nil
}

func (t *Mocker) SetupFirewall() error {
	return nil
}

// Apply takes the network endpoint configuration and applies it to the system
func (t *Mocker) Apply(endpoint *tether.NetworkEndpoint) error {
	defer trace.End(trace.Begin("mocking endpoint configuration for " + endpoint.Network.Name))
	t.IPs[endpoint.Network.Name] = endpoint.Assigned.IP

	return nil
}

// MountLabel performs a mount with the source treated as a disk label
// This assumes that /dev/disk/by-label is being populated, probably by udev
func (t *Mocker) MountLabel(ctx context.Context, label, target string) error {
	defer trace.End(trace.Begin(fmt.Sprintf("mocking mounting %s on %s", label, target)))

	if t.Mounts == nil {
		t.Mounts = make(map[string]string)
	}

	t.Mounts[label] = target
	return nil
}

// MountTarget performs a mount with the source treated as an nfs target
func (t *Mocker) MountTarget(ctx context.Context, source url.URL, target string, mountOptions string) error {
	defer trace.End(trace.Begin(fmt.Sprintf("mocking mounting %s on %s", source.String(), target)))

	if t.Mounts == nil {
		t.Mounts = make(map[string]string)
	}

	t.Mounts[source.String()] = target
	return nil
}

// Fork triggers vmfork and handles the necessary pre/post OS level operations
func (t *Mocker) Fork() error {
	defer trace.End(trace.Begin("mocking fork"))
	return errors.New("Fork test not implemented")
}

// TestMain simply so we have control of debugging level and somewhere to call package wide test setup
func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)

	// replace the Sys variable with a mock
	tether.Sys = system.System{
		Hosts:      &tether.MockHosts{},
		ResolvConf: &tether.MockResolvConf{},
		Syscall:    &tether.MockSyscall{},
		Root:       os.TempDir(),
	}

	retCode := m.Run()

	// call with result of m.Run()
	os.Exit(retCode)
}

func StartAttachTether(t *testing.T, cfg *executor.ExecutorConfig, mocker *Mocker) (tether.Tether, extraconfig.DataSource, net.Conn) {
	store := extraconfig.New()
	sink := store.Put
	src := store.Get
	extraconfig.Encode(sink, cfg)
	log.Debugf("Test configuration: %#v", sink)

	tthr = tether.New(src, sink, mocker)
	tthr.Register("mocker", mocker)
	tthr.Register("Attach", server)

	// run the tether to service the attach
	go func() {
		erR := tthr.Start()
		if erR != nil {
			t.Error(erR)
		}
	}()

	// create client on the mock pipe
	conn, err := mockBackChannel(context.Background())
	if err != nil && (err != io.EOF || server.(*testAttachServer).enabled) {
		// we accept the case where the error is end-of-file and the attach server is disabled because that's
		// expected when the tether is shut down.
		t.Error(err)
	}

	return tthr, src, conn
}

func tetherTestSetup(t *testing.T) (string, *Mocker) {
	pc, _, _, _ := runtime.Caller(2)
	name := runtime.FuncForPC(pc).Name()

	log.Infof("Started test setup for %s", name)

	// use the mock ops - fresh one each time as tests might apply different mocked calls
	mocker := Mocker{
		Started: make(chan bool, 0),
		Cleaned: make(chan bool, 0),
	}

	return name, &mocker
}

func tetherTestTeardown(t *testing.T, mocker *Mocker) string {
	<-mocker.Cleaned

	// cleanup
	os.RemoveAll(pathPrefix)
	log.SetOutput(os.Stdout)

	pc, _, _, _ := runtime.Caller(2)
	name := runtime.FuncForPC(pc).Name()

	log.Infof("Finished test teardown for %s", name)

	return name
}
