/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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

package main

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/rjeczalik/notify"
	"k8s.io/client-go/util/flowcontrol"
)

type haproxy struct {
	cfg   string
	sha   string
	mutex *sync.Mutex

	reloadRateLimiter flowcontrol.RateLimiter
}

func (h *haproxy) reload() error {
	h.reloadRateLimiter.Accept()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	c, err := checksum(h.cfg)
	if err != nil {
		return fmt.Errorf("error getting sha of %v: %v", h.cfg, err)
	}

	if strings.Compare(h.sha, c) == 0 {
		// no changes, skip reload
		return nil
	}

	h.sha = c

	output, err := runCommand("sh", []string{"-c", "/haproxy_reload"})
	if err != nil {
		return err
	}

	log.Printf("reloading %v\n", string(output))
	return nil
}

func main() {
	log.Println("triggering haproxy reload to start the process")
	h := haproxy{"/etc/haproxy/haproxy.cfg", "", &sync.Mutex{}, flowcontrol.NewTokenBucketRateLimiter(0.1, 1)}
	err := h.reload()
	if err != nil {
		log.Fatalf("unexpected error starting haproxy: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(h.cfg, c, notify.Write); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	for {
		select {
		case <-quit:
			log.Println("terminating...")
			os.Exit(0)
			break
		case _ = <-c:
			err := h.reload()
			if err != nil {
				log.Printf("unexpected error reloading haproxy: %v", err)
			}
		}
	}
}

// runCommand executes a command returning the stdout and
// stderr output.
func runCommand(command string, args []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []byte{}, fmt.Errorf("command error %q: %v\n%s", strings.Join(cmd.Args, " "), err, string(output))
	}

	return output, nil
}

// checksum returns the sha1 of a file or an
// error if something happened reading the file
func checksum(filename string) (string, error) {
	c, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", sha1.Sum(c)), nil
}
