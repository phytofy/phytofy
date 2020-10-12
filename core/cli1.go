// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code handles the CLI for PHYTOFY RL v1
package main

import (
	"encoding/json"
	"log"
	"strconv"
)

func cli1Wrapper(command string, argument string, logger *log.Logger) (string, error) {
	api := api1Init(logger, false)
	if command == "v1-get-serials" {
		api.controller.discoverer.dscvr1WaitForAnySerials(dscvr1DiscoveryInterval)
	}
	result, fail := api.api1Dispatch(command[3:], []byte(argument))
	return string(result), fail
}

func cli1ImportSchedules(command string, argument string, logger *log.Logger) (string, error) {
	schedules, fail := schdlReadSchedulesFromFile(argument, 6)
	if fail != nil {
		return "", fail
	}
	jsonSchedules, fail := json.Marshal(&schedules)
	if fail != nil {
		return "", fail
	}
	api := api1Init(logger, false)
	result, fail := api.api1ImportSchedules(jsonSchedules)
	return string(result), fail
}

func cli1Web(includeUI bool) cliFunction {
	return func(command string, argument string, logger *log.Logger) (string, error) {
		api := api1Init(logger, includeUI)
		port, fail := strconv.ParseUint(argument, 10, 16)
		if fail != nil {
			return "", fail
		}
		return "", api.api1Launch(uint16(port), includeUI)
	}
}

// This function registers the commands & arguments to the CLI for PHYTOFY RL v1
func cli1Commands() []cliCommand {
	return []cliCommand{
		cliCommand{"v1-set-module-calibration", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-module-calibration", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-serial-number", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-serial-number", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-short-address", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-short-address", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-group-id", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-group-id", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-fixture-info", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-fixture-info", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-time-reference", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-time-reference", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-leds-pwm", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-leds-irradiance", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-leds", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-schedule-pwm", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-schedule-irradiance", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-schedule", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-schedule-count", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-scheduling-state", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-delete-schedule", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-delete-all-schedules", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-stop-scheduling", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-resume-scheduling", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-set-illuminance-configuration", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-illuminance-configuration", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-module-temperature", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-get-serials", "JSON", "JSON-formatted input for the command", cli1Wrapper},
		cliCommand{"v1-import-schedules", "CSV", "CSV file with schedules & recipes", cli1ImportSchedules},
		cliCommand{"v1-api", "PORT", "TCP port to expose API on", cli1Web(false)},
		cliCommand{"v1-app", "PORT", "TCP port to expose API & UI on", cli1Web(true)},
	}
}
