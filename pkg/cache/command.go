// Copyright (c) Facebook, Inc. and its affiliates. All Rights Reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"log"
	"os/exec"
	"time"
)

// Command encapsulates a command to run.
type Command struct {
	Command string   // Name of command (relative or absolute)
	Args    []string // Slice of arguments for command
	Timeout int      // Number of seconds before process times out
}

// RunCommand forks a process to run Command returns its output with err (if
// any). If the process times out it will be killed.
func RunCommand(command Command) ([]byte, error) {
	absPath, err := exec.LookPath(command.Command)
	if err != nil {
		log.Printf("didn't find %s executable", command.Command)
		return nil, err
	}

	// We only use context for the timeout and kill process functionality...
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(command.Timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, absPath, command.Args...)
	output, err := cmd.Output()

	// The error returned by cmd.Output() will be OS specific based on what
	// happens when a process is killed.
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("Command %s timed out\n", command.Command)
		return []byte(nil), ctx.Err()
	}

	// If there's no context error, we know the command completed (or errored).
	if err != nil {
		log.Printf("Command %s returned non-zero, err %v, output %v\n", command.Command, err, output)
		return []byte(nil), err
	}

	return output, nil
}
