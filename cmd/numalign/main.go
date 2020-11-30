/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2020 Red Hat, Inc.
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/openshift-kni/debug-tools/internal/pkg/numalign"
	"github.com/openshift-kni/debug-tools/internal/pkg/vfs"
)

type config struct {
	sleepHours string
	procfsRoot string
	sysfsRoot  string
	debug      bool
}

func (cfg *config) SetFlags() {
	flag.StringVarP(&cfg.sleepHours, "sleep-hours", "S", "", "sleep hours once done.")
	flag.StringVarP(&cfg.procfsRoot, "procfs", "p", "/proc", "procfs root.")
	flag.StringVarP(&cfg.sysfsRoot, "sysfs", "s", "/sys", "sysfs root.")
	flag.BoolVarP(&cfg.debug, "debug", "D", false, "enable debug mode.")
}

func (cfg *config) GetProcFSRoot() string {
	return cfg.procfsRoot
}

func (cfg *config) GetSysFSRoot() string {
	return cfg.sysfsRoot
}

func (cfg *config) IsDebugEnabled() bool {
	if _, ok := os.LookupEnv("NUMALIGN_DEBUG"); ok {
		return true
	}
	return cfg.debug
}

func (cfg *config) GetSleepTime() time.Duration {
	var sleepTime time.Duration
	if val, ok := os.LookupEnv("NUMALIGN_SLEEP_HOURS"); ok {
		cfg.sleepHours = val
	}
	if cfg.sleepHours == "" {
		return 0
	}
	hours, err := strconv.Atoi(cfg.sleepHours)
	if err != nil {
		log.Fatalf("%v", err)
	}
	if hours > 0 {
		sleepTime = time.Duration(hours) * time.Hour
	}
	return sleepTime
}

func (cfg config) String() string {
	return fmt.Sprintf("sleep=%v procfs=%q sysfs=%q debug=%v", cfg.sleepHours, cfg.procfsRoot, cfg.sysfsRoot, cfg.debug)
}

func main() {
	cfg := &config{}
	cfg.SetFlags()
	flag.Parse()

	if !cfg.IsDebugEnabled() {
		log.SetOutput(ioutil.Discard)
	} else {
		log.Printf("SYS: %s", cfg)
	}

	sleepTime := cfg.GetSleepTime()

	numaRes, err := numalign.NewResources(vfs.LinuxFS{}, cfg.GetProcFSRoot(), cfg.GetSysFSRoot(), os.Environ(), flag.Args())
	if err != nil {
		log.Fatalf("%v", err)
	}

	res := numaRes.CheckAlignment()
	fmt.Printf("%s", res.JSON())

	time.Sleep(sleepTime)
	if !res.Aligned {
		os.Exit(-1)
	}
}
