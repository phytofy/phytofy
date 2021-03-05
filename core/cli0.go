// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code handles the CLI for PHYTOFY RL v0 (DEPRECATED)
package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

func cli0Wrapper(command string, argument string, logger *log.Logger) (string, error) {
	api := api0Init(logger)
	if command == "v0-get-serials" {
		api.controller.ctrl0WaitForAnySerials(ctrl0HeartbeatInterval)
	}
	result, fail := api.api0Dispatch(command[3:], []byte(argument))
	time.Sleep(5 * time.Second) // Wait until the commands are flushed (a consequence of protocol design)
	return string(result), fail
}

func cli0ImportSchedules(command string, argument string, logger *log.Logger) (string, error) {
	schedules, fail := schdlReadSchedulesFromFile(argument, 6)
	if fail != nil {
		return "", fail
	}
	jsonSchedules, fail := json.Marshal(&schedules)
	if fail != nil {
		return "", fail
	}
	api := api0Init(logger)
	result, fail := api.api0ImportSchedules(jsonSchedules)
	time.Sleep(5 * time.Second) // Wait until the commands are flushed (a consequence of protocol design)
	return string(result), fail
}

func cli0Web(includeUI bool) cliFunction {
	return func(command string, argument string, logger *log.Logger) (string, error) {
		api := api0Init(logger)
		port, fail := strconv.ParseUint(argument, 10, 16)
		if fail != nil {
			return "", fail
		}
		return "", api.api0Launch(uint16(port), includeUI)
	}
}

// This function registers the commands & arguments to the CLI for PHYTOFY RL v0
func cli0Commands() []cliCommand {
	return []cliCommand{
		{"v0-set-leds", "JSON", "JSON-formatted input for the command", cli0Wrapper},
		{"v0-schedule-add", "JSON", "JSON-formatted input for the command", cli0Wrapper},
		{"v0-schedules-clear", "JSON", "JSON-formatted input for the command", cli0Wrapper},
		{"v0-get-serials", "JSON", "JSON-formatted input for the command", cli0Wrapper},
		{"v0-import-schedules", "CSV", "CSV file with schedules & recipes", cli0ImportSchedules},
		{"v0-api", "PORT", "TCP port to expose API on", cli0Web(false)},
		{"v0-app", "PORT", "TCP port to expose API & UI on", cli0Web(true)},
	}
}
