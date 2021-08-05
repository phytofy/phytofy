// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code handles the OpenAPI for PHYTOFY RL v1
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type api1 struct {
	logger     *log.Logger
	controller *ctrl1Controller
}

type api1GenericArguments struct {
	Serial  schdlSerial     `json:"serial"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type api1GenericResult struct {
	Replies []pckt1Packet `json:"replies"`
	Error   string        `json:"error,omitempty"`
}

type api1GetSerialsResult struct {
	Serials schdlSerials `json:"serials"`
}

type api1ImportSchedulesArguments struct {
	Schedules []schdlAttached `json:"schedules"`
}

type api1ImportSchedulesResult struct {
	Error string `json:"error,omitempty"`
}

func api1Init(logger *log.Logger, conditioning bool) *api1 {
	return &api1{
		logger,
		ctrl1Init(logger, conditioning),
	}
}

// Handles the "get-serials" command
func (api *api1) api1GetSerials(jsonArguments []byte) ([]byte, error) {
	result := api1GetSerialsResult{api.controller.ctrl1GetSerials()}
	jsonResult, fail := json.Marshal(&result)
	if fail != nil {
		return nil, fail
	}
	return jsonResult, nil
}

// Handles the "import-schedules" command
func (api *api1) api1ImportSchedules(jsonArguments []byte) ([]byte, error) {
	var arguments api1ImportSchedulesArguments
	var result api1ImportSchedulesResult
	var fail error
	if fail = json.Unmarshal(jsonArguments, &arguments); fail != nil {
		result = api1ImportSchedulesResult{fail.Error()}
	} else if fail = api.controller.ctrl1ImportSchedules(arguments.Schedules); fail != nil {
		result = api1ImportSchedulesResult{fail.Error()}
	}
	jsonResult, critical := json.Marshal(&result)
	if critical != nil {
		return nil, critical
	}
	return jsonResult, fail
}

// Dispatches API function call
func (api *api1) api1Dispatch(name string, jsonArguments []byte) ([]byte, error) {
	switch name {
	case "set-module-calibration", "get-module-calibration", "set-serial-number", "get-serial-number", "set-short-address", "get-short-address", "set-group-id", "get-group-id", "set-fixture-info", "get-fixture-info", "set-time-reference", "get-time-reference", "set-leds-pwm", "set-leds-irradiance", "get-leds", "set-schedule-pwm", "set-schedule-irradiance", "get-schedule", "get-schedule-count", "get-scheduling-state", "delete-schedule", "delete-all-schedules", "stop-scheduling", "resume-scheduling", "set-illuminance-configuration", "get-illuminance-configuration", "get-module-temperature", "toggle-calibration", "reset-for-firmware-update", "confirm-reset-for-firmware-update":
		serial, functionCode, payload, fail := ctrl1ParseGenericArguments(name, jsonArguments)
		if fail != nil {
			return []byte{}, fail
		}
		replies, fail := api.controller.ctrl1Dispatch(serial, functionCode, payload)
		errorMessage := ""
		if fail != nil {
			errorMessage = fail.Error()
		}
		result := api1GenericResult{replies, errorMessage}
		jsonResult, critical := json.Marshal(&result)
		if critical != nil {
			return []byte{}, critical
		}
		return jsonResult, fail
	case "get-serials":
		return api.api1GetSerials(jsonArguments)
	case "import-schedules":
		return api.api1ImportSchedules(jsonArguments)
	}
	return []byte{}, fmt.Errorf("Unknown API function - %s", name)
}

// Launches a web server for PHYTOFY RL v1
func (api *api1) api1Launch(port uint16, includeUI bool) error {
	routes := []webRoute{
		{"set-module-calibration", http.MethodPost, "/v1/set-module-calibration", api.api1Dispatch},
		{"get-module-calibration", http.MethodPost, "/v1/get-module-calibration", api.api1Dispatch},
		{"set-serial-number", http.MethodPost, "/v1/set-serial-number", api.api1Dispatch},
		{"get-serial-number", http.MethodPost, "/v1/get-serial-number", api.api1Dispatch},
		{"set-short-address", http.MethodPost, "/v1/set-short-address", api.api1Dispatch},
		{"get-short-address", http.MethodPost, "/v1/get-short-address", api.api1Dispatch},
		{"set-group-id", http.MethodPost, "/v1/set-group-id", api.api1Dispatch},
		{"get-group-id", http.MethodPost, "/v1/get-group-id", api.api1Dispatch},
		{"set-fixture-info", http.MethodPost, "/v1/set-fixture-info", api.api1Dispatch},
		{"get-fixture-info", http.MethodPost, "/v1/get-fixture-info", api.api1Dispatch},
		{"set-time-reference", http.MethodPost, "/v1/set-time-reference", api.api1Dispatch},
		{"get-time-reference", http.MethodPost, "/v1/get-time-reference", api.api1Dispatch},
		{"set-leds-pwm", http.MethodPost, "/v1/set-leds-pwm", api.api1Dispatch},
		{"set-leds-irradiance", http.MethodPost, "/v1/set-leds-irradiance", api.api1Dispatch},
		{"get-leds", http.MethodPost, "/v1/get-leds", api.api1Dispatch},
		{"set-schedule-pwm", http.MethodPost, "/v1/set-schedule-pwm", api.api1Dispatch},
		{"set-schedule-irradiance", http.MethodPost, "/v1/set-schedule-irradiance", api.api1Dispatch},
		{"get-schedule", http.MethodPost, "/v1/get-schedule", api.api1Dispatch},
		{"get-schedule-count", http.MethodPost, "/v1/get-schedule-count", api.api1Dispatch},
		{"get-scheduling-state", http.MethodPost, "/v1/get-scheduling-state", api.api1Dispatch},
		{"delete-schedule", http.MethodPost, "/v1/delete-schedule", api.api1Dispatch},
		{"delete-all-schedules", http.MethodPost, "/v1/delete-all-schedules", api.api1Dispatch},
		{"stop-scheduling", http.MethodPost, "/v1/stop-scheduling", api.api1Dispatch},
		{"resume-scheduling", http.MethodPost, "/v1/resume-scheduling", api.api1Dispatch},
		{"set-illuminance-configuration", http.MethodPost, "/v1/set-illuminance-configuration", api.api1Dispatch},
		{"get-illuminance-configuration", http.MethodPost, "/v1/get-illuminance-configuration", api.api1Dispatch},
		{"get-module-temperature", http.MethodPost, "/v1/get-module-temperature", api.api1Dispatch},
		{"toggle-calibration", http.MethodPost, "/v1/toggle-calibration", api.api1Dispatch},
		{"reset-for-firmware-update", http.MethodPost, "/v1/reset-for-firmware-update", api.api1Dispatch},
		{"confirm-reset-for-firmware-update", http.MethodPost, "/v1/confirm-reset-for-firmware-update", api.api1Dispatch},
		{"get-serials", http.MethodGet, "/v1/get-serials", api.api1Dispatch},
		{"get-serials", http.MethodGet, "/api/get-serials", api.api1Dispatch},
		{"import-schedules", http.MethodPost, "/api/import-schedules", api.api1Dispatch},
	}
	return webLaunch(port, routes, includeUI, api.logger)
}
