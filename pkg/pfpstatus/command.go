/*
 * Copyright 2023 Red Hat, Inc.
 *
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
 */

package pfpstatus

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	CommandName = "pfpstatus"
)

func IsCommand(argv0 string) bool {
	return filepath.Base(argv0) == CommandName
}

func Execute() int {
	var targets []string
	if len(os.Args[1:]) > 0 {
		targets = append(targets, os.Args[1:]...)
	} else {
		matches, err := filepath.Glob(filepath.Join(BaseDirectory, "*.json"))
		if err != nil {
			// TODO: log
			return 1
		}
		targets = append(targets, matches...)
	}
	fmt.Println("[")
	if len(targets) > 0 {
		data, err := os.ReadFile(targets[0])
		if err != nil {
			// TODO: log
			return 2
		}
		fmt.Println(string(data))
	}
	for _, target := range targets[1:] {
		fmt.Println(",")
		data, err := os.ReadFile(target)
		if err != nil {
			// TODO: log
			return 4
		}
		fmt.Println(string(data))
	}
	fmt.Println("]")
	return 0
}
