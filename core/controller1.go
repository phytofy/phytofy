// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for controlling an abstract interface to PHYTOFY RL v1
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"
)

var ctrl1NameToFunctionCode map[string]pckt1FunctionCode = map[string]pckt1FunctionCode{
	"set-module-calibration":            pckt1FunctionCodeSetModuleCalibration,
	"get-module-calibration":            pckt1FunctionCodeGetModuleCalibration,
	"set-serial-number":                 pckt1FunctionCodeSetSerialNumber,
	"get-serial-number":                 pckt1FunctionCodeGetSerialNumber,
	"set-short-address":                 pckt1FunctionCodeSetShortAddress,
	"get-short-address":                 pckt1FunctionCodeGetShortAddress,
	"set-group-id":                      pckt1FunctionCodeSetGroupID,
	"get-group-id":                      pckt1FunctionCodeGetGroupID,
	"set-fixture-info":                  pckt1FunctionCodeSetFixtureInfo,
	"get-fixture-info":                  pckt1FunctionCodeGetFixtureInfo,
	"set-time-reference":                pckt1FunctionCodeSetTimeReference,
	"get-time-reference":                pckt1FunctionCodeGetTimeReference,
	"set-leds-pwm":                      pckt1FunctionCodeSetLEDs,
	"set-leds-irradiance":               pckt1FunctionCodeSetLEDs,
	"get-leds":                          pckt1FunctionCodeGetLEDs,
	"set-schedule-pwm":                  pckt1FunctionCodeSetSchedule,
	"set-schedule-irradiance":           pckt1FunctionCodeSetSchedule,
	"get-schedule":                      pckt1FunctionCodeGetSchedule,
	"get-schedule-count":                pckt1FunctionCodeGetScheduleCount,
	"get-scheduling-state":              pckt1FunctionCodeGetSchedulingState,
	"delete-schedule":                   pckt1FunctionCodeDeleteSchedule,
	"delete-all-schedules":              pckt1FunctionCodeDeleteAllSchedules,
	"stop-scheduling":                   pckt1FunctionCodeStopScheduling,
	"resume-scheduling":                 pckt1FunctionCodeResumeScheduling,
	"set-illuminance-configuration":     pckt1FunctionCodeSetIlluminanceConfiguration,
	"get-illuminance-configuration":     pckt1FunctionCodeGetIlluminanceConfiguration,
	"get-module-temperature":            pckt1FunctionCodeGetModuleTemperature,
	"toggle-calibration":                pckt1FunctionCodeToggleCalibration,
	"reset-for-firmware-update":         pckt1FunctionCodeResetForFirmwareUpdate,
	"confirm-reset-for-firmware-update": pckt1FunctionCodeConfirmResetForFirmwareUpdate,
}

// Controls the PHYTOFY RL v1 fixtures
type ctrl1Controller struct {
	logger     *log.Logger
	discoverer *dscvr1Discoverer
}

// Creates an instance of PHYTOFY RL v0 controller
func ctrl1Init(logger *log.Logger, conditioning bool) *ctrl1Controller {
	discoverer := dscvr1Init(logger, conditioning)
	controller := &ctrl1Controller{logger, discoverer}
	return controller
}

// Parse arguments
func ctrl1ParseGenericArguments(name string, jsonArguments []byte) (schdlSerial, pckt1FunctionCode, pckt1Payload, error) {
	functionCode, present := ctrl1NameToFunctionCode[name]
	if !present {
		return 0, 0xFF, nil, fmt.Errorf("Unknown command - %s", name)
	}
	var arguments api1GenericArguments
	if fail := json.Unmarshal(jsonArguments, &arguments); fail != nil {
		return 0, 0xFF, nil, fmt.Errorf("Failed to parse arguments (%s) - %s", fail, string(jsonArguments))
	}
	var payload pckt1Payload
	switch name {
	case "set-module-calibration":
		payload = new(pckt1CommandPayloadSetModuleCalibration)
	case "get-module-calibration":
		payload = new(pckt1CommandPayloadGetModuleCalibration)
	case "set-serial-number":
		payload = new(pckt1CommandPayloadSetSerialNumber)
	case "get-serial-number":
		payload = new(pckt1CommandPayloadGetSerialNumber)
	case "set-short-address":
		payload = new(pckt1CommandPayloadSetShortAddress)
	case "get-short-address":
		payload = new(pckt1CommandPayloadGetShortAddress)
	case "set-group-id":
		payload = new(pckt1CommandPayloadSetGroupID)
	case "get-group-id":
		payload = nil
	case "set-fixture-info":
		payload = new(pckt1CommandPayloadSetFixtureInfo)
	case "get-fixture-info":
		payload = nil
	case "set-time-reference":
		payload = new(pckt1CommandPayloadSetTimeReference)
	case "get-time-reference":
		payload = nil
	case "set-leds-pwm":
		payload = new(pckt1CommandPayloadSetLEDsPWM)
	case "set-leds-irradiance":
		payload = new(pckt1CommandPayloadSetLEDsIrradiance)
	case "get-leds":
		payload = new(pckt1CommandPayloadGetLEDs)
	case "set-schedule-pwm":
		payload = new(pckt1CommandPayloadSetSchedulePWM)
	case "set-schedule-irradiance":
		payload = new(pckt1CommandPayloadSetScheduleIrradiance)
	case "get-schedule":
		payload = new(pckt1CommandPayloadGetSchedule)
	case "get-schedule-count":
		payload = nil
	case "get-scheduling-state":
		payload = nil
	case "delete-schedule":
		payload = new(pckt1CommandPayloadDeleteSchedule)
	case "delete-all-schedules":
		payload = nil
	case "stop-scheduling":
		payload = nil
	case "resume-scheduling":
		payload = nil
	case "set-illuminance-configuration":
		payload = new(pckt1CommandPayloadSetIlluminanceConfiguration)
	case "get-illuminance-configuration":
		payload = nil
	case "get-module-temperature":
		payload = nil
	default:
		return 0, 0xFF, nil, fmt.Errorf("Unknown API call %s", name)
	}
	if payload != nil {
		if fail := json.Unmarshal(arguments.Payload, payload); fail != nil {
			return 0, 0xFF, nil, fmt.Errorf("Failed to parse payload (%s) - %s", fail, string(arguments.Payload))
		}
	}
	return arguments.Serial, functionCode, payload, nil
}

// Lists all seen serials
func (controller *ctrl1Controller) ctrl1GetSerials() schdlSerials {
	serialsSet := make(map[schdlSerial]struct{})
	controller.discoverer.adapters.Range(func(key, value interface{}) bool {
		for _, serial := range value.(*dptr1Adapter).dptr1ListSeenSerials() {
			serialsSet[serial] = struct{}{}
		}
		return true
	})
	serials := make(schdlSerials, 0)
	for serial := range serialsSet {
		serials = append(serials, serial)
	}
	sort.Slice(serials, func(i, j int) bool { return serials[i] < serials[j] })
	return serials
}

// Dispatches a call to adapter(s)
func (controller *ctrl1Controller) ctrl1Dispatch(serial schdlSerial, functionCode pckt1FunctionCode, payload pckt1Payload) ([]pckt1Packet, error) {
	if !controller.discoverer.dscvr1WaitForSerial(serial, time.Minute) {
		return nil, fmt.Errorf("Timed out waiting for device with serial number %d", serial)
	}
	adapters := controller.discoverer.dscvr1LookUp(serial)
	result := make([]pckt1Packet, 0)
	for _, adapter := range adapters {
		shortAddress := adapter.dptr1LookUp(serial)
		if shortAddress != pckt1ShortAddressUnassigned {
			replies, fail := adapter.dptr1AssembleAndExchange(shortAddress, functionCode, payload)
			if fail != nil {
				return nil, fmt.Errorf("Failed to communicate with device with serial number %d (%s)", serial, fail)
			}
			result = append(result, replies...)
		} else {
			return nil, fmt.Errorf("Could not look up device with serial number %d", serial)
		}
	}
	fail := dptr1CheckReplies(functionCode, result)
	return result, fail
}

// Import schedules
func (controller *ctrl1Controller) ctrl1ImportSchedules(schedules []schdlAttached) error {
	aggregated, fail := schdlAggregateSchedules(schedules, false)
	if fail != nil {
		controller.logger.Printf("ERROR: Failed to aggregate schedules (%s)", fail)
		return fail
	}
	serials := make(schdlSerials, 0)
	for serial := range aggregated {
		serials = append(serials, serial)
	}
	if !controller.discoverer.dscvr1WaitForSerials(serials, time.Minute) {
		fail := fmt.Errorf("Failed to locate all fixtures")
		controller.logger.Printf("ERROR: %s", fail)
		return fail
	}
	for serial := range aggregated {
		repliesDelete, failDelete := controller.ctrl1Dispatch(serial, pckt1FunctionCodeDeleteAllSchedules, nil)
		if fail := dptr1CheckResult(pckt1FunctionCodeDeleteAllSchedules, repliesDelete, failDelete); fail != nil {
			fail := fmt.Errorf("Failed to delete schedule for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
		repliesSync, failSync := controller.ctrl1Dispatch(serial, pckt1FunctionCodeSetTimeReference, &pckt1CommandPayloadSetTimeReference{uint32(time.Now().Unix())})
		if fail := dptr1CheckResult(pckt1FunctionCodeSetTimeReference, repliesSync, failSync); fail != nil {
			fail := fmt.Errorf("Failed to sync time for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
		for scheduleID, schedule := range aggregated[serial] {
			config := pckt1UsePWM | pckt1LEDsModule0Enabled | pckt1LEDsModule1Enabled
			var levels [6]uint32
			for i := 0; i < 6; i++ {
				levels[i] = uint32(schedule.Levels[i])
			}
			payload := &pckt1CommandPayloadSetSchedulePWM{pckt1CommandPayloadSetSchedulePreamble{uint32(scheduleID), schedule.Start, schedule.Stop, config}, levels}
			repliesSet, failSet := controller.ctrl1Dispatch(serial, pckt1FunctionCodeSetSchedule, payload)
			if fail := dptr1CheckResult(pckt1FunctionCodeSetSchedule, repliesSet, failSet); fail != nil {
				fail := fmt.Errorf("Failed to set schedule %d for device with serial number %d (%s)", scheduleID, serial, fail)
				controller.logger.Printf("ERROR: %s", fail)
				return fail
			}
		}
		repliesResume, failResume := controller.ctrl1Dispatch(serial, pckt1FunctionCodeResumeScheduling, nil)
		if fail := dptr1CheckResult(pckt1FunctionCodeResumeScheduling, repliesResume, failResume); fail != nil {
			fail := fmt.Errorf("Failed to resume scheduling for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
		repliesCalibration0, failCalibration0 := controller.ctrl1Dispatch(serial, pckt1FunctionCodeGetModuleCalibration, &pckt1CommandPayloadGetModuleCalibration{0})
		if fail := dptr1CheckResult(pckt1FunctionCodeGetModuleCalibration, repliesCalibration0, failCalibration0); fail != nil {
			fail := fmt.Errorf("Failed to fetch module 0 calibration for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
		repliesCalibration1, failCalibration1 := controller.ctrl1Dispatch(serial, pckt1FunctionCodeGetModuleCalibration, &pckt1CommandPayloadGetModuleCalibration{1})
		if fail := dptr1CheckResult(pckt1FunctionCodeGetModuleCalibration, repliesCalibration1, failCalibration1); fail != nil {
			fail := fmt.Errorf("Failed to fetch module 1 calibration for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
		calibration0 := repliesCalibration0[0].Payload.(*pckt1ReplyPayloadGetModuleCalibration).Calibration
		calibration1 := repliesCalibration1[0].Payload.(*pckt1ReplyPayloadGetModuleCalibration).Calibration
		configuration := dptr1IlluminanceConfiguration(calibration0, calibration1)
		repliesIlluminance, failIlluminance := controller.ctrl1Dispatch(serial, pckt1FunctionCodeSetIlluminanceConfiguration, &pckt1CommandPayloadSetIlluminanceConfiguration{configuration})
		if fail := dptr1CheckResult(pckt1FunctionCodeSetIlluminanceConfiguration, repliesIlluminance, failIlluminance); fail != nil {
			fail := fmt.Errorf("Failed to set illuminance configuration for device with serial number %d (%s)", serial, fail)
			controller.logger.Printf("ERROR: %s", fail)
			return fail
		}
	}
	return nil
}
