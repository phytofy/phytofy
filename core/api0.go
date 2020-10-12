// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code handles the OpenAPI for PHYTOFY RL v0 (DEPRECATED)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type api0 struct {
	logger     *log.Logger
	controller *ctrl0Controller
}

type api0SetLedsArguments struct {
	Serial  schdlSerial `json:"serial"`
	Payload struct {
		Levels schdlLevels `json:"levels"`
	} `json:"payload"`
}

type api0ScheduleAddArguments struct {
	Serial  schdlSerial `json:"serial"`
	Payload struct {
		Levels     schdlLevels `json:"levels"`
		Start      uint32      `json:"start"`
		Stop       uint32      `json:"stop"`
		ScheduleID uint32      `json:"schedule_id"`
	} `json:"payload"`
}

type api0GetSerialsResult struct {
	Serials schdlSerials `json:"serials"`
}

type api0ImportSchedulesArguments struct {
	Schedules []schdlAttached `json:"schedules"`
}

type api0ImportSchedulesResult struct {
	Error string `json:"error,omitempty"`
}

func api0Init(logger *log.Logger) *api0 {
	return &api0{
		logger,
		ctrl0Init(logger),
	}
}

// Handles the "set-leds" command
func (api *api0) api0SetLeds(jsonArguments []byte) ([]byte, error) {
	var arguments api0SetLedsArguments
	if fail := json.Unmarshal(jsonArguments, &arguments); fail != nil {
		return nil, fail
	}
	if fail := schdlCheckLevels(arguments.Payload.Levels); fail != nil {
		return nil, fail
	}
	if !api.controller.ctrl0WaitForSerials(schdlSerials{arguments.Serial}, time.Minute) {
		return nil, fmt.Errorf("Failed to locate the fixture (to set levels), seen - %v", api.controller.ctrl0GetSerials())
	}
	if !api.controller.ctrl0TransmitLedsSetRequest(arguments.Serial, arguments.Payload.Levels) {
		return nil, fmt.Errorf("Failed to communicate with the fixture (to set levels)")
	}
	return nil, nil
}

// Handles the "schedule-add" command
func (api *api0) api0ScheduleAdd(jsonArguments []byte) ([]byte, error) {
	var arguments api0ScheduleAddArguments
	if fail := json.Unmarshal(jsonArguments, &arguments); fail != nil {
		return nil, fail
	}
	if fail := schdlCheckLevels(arguments.Payload.Levels); fail != nil {
		return nil, fail
	}
	if !api.controller.ctrl0WaitForSerials(schdlSerials{arguments.Serial}, time.Minute) {
		return nil, fmt.Errorf("Failed to locate the fixture (to add schedule), seen - %v", api.controller.ctrl0GetSerials())
	}
	schedule := schdlDetached{schdlTiming{arguments.Payload.Start, arguments.Payload.Stop}, arguments.Payload.Levels}
	if !api.controller.ctrl0TransmitScheduleAddRequest(arguments.Serial, schedule, arguments.Payload.ScheduleID) {
		return nil, fmt.Errorf("Failed to communicate with the fixture (to add schedule)")
	}
	return nil, nil
}

// Handles the "schedules-clear" command
func (api *api0) api0SchedulesClear(jsonArguments []byte) ([]byte, error) {
	if !api.controller.ctrl0TransmitScheduleClearRequests() {
		return nil, fmt.Errorf("Failed to communicate with the fixtures (to clear schedules)")
	}
	return nil, nil
}

// Handles the "get-serials" command
func (api *api0) api0GetSerials(jsonArguments []byte) ([]byte, error) {
	result := api0GetSerialsResult{api.controller.ctrl0GetSerials()}
	jsonResult, fail := json.Marshal(&result)
	if fail != nil {
		return nil, fail
	}
	return jsonResult, nil
}

// Handles the "import-schedules" command
func (api *api0) api0ImportSchedules(jsonArguments []byte) ([]byte, error) {
	var arguments api0ImportSchedulesArguments
	var result api0ImportSchedulesResult
	var fail error
	if fail = json.Unmarshal(jsonArguments, &arguments); fail != nil {
		result = api0ImportSchedulesResult{fail.Error()}
	} else if fail = api.controller.ctrl0ImportSchedules(arguments.Schedules); fail != nil {
		result = api0ImportSchedulesResult{fail.Error()}
	}
	jsonResult, critical := json.Marshal(&result)
	if critical != nil {
		return nil, critical
	}
	return jsonResult, fail
}

// Dispatches API function call
func (api *api0) api0Dispatch(name string, jsonArguments []byte) ([]byte, error) {
	switch name {
	case "set-leds":
		return api.api0SetLeds(jsonArguments)
	case "schedule-add":
		return api.api0ScheduleAdd(jsonArguments)
	case "schedules-clear":
		return api.api0SchedulesClear(jsonArguments)
	case "get-serials":
		return api.api0GetSerials(jsonArguments)
	case "import-schedules":
		return api.api0ImportSchedules(jsonArguments)
	}
	return []byte{}, fmt.Errorf("Unknown API function - %s", name)
}

// Launches a web server for PHYTOFY RL v0
func (api *api0) api0Launch(port uint16, includeUI bool) error {
	routes := []webRoute{
		webRoute{"set-leds", http.MethodPost, "/v0/set-leds", api.api0Dispatch},
		webRoute{"schedule-add", http.MethodPost, "/v0/schedule-add", api.api0Dispatch},
		webRoute{"schedules-clear", http.MethodPost, "/v0/schedules-clear", api.api0Dispatch},
		webRoute{"get-serials", http.MethodGet, "/v0/get-serials", api.api0Dispatch},
		webRoute{"get-serials", http.MethodGet, "/api/get-serials", api.api0Dispatch},
		webRoute{"import-schedules", http.MethodPost, "/api/import-schedules", api.api0Dispatch},
	}
	return webLaunch(port, routes, includeUI, api.logger)
}
