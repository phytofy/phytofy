// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code handles the control of PHYTOFY RL v0 (DEPRECATED)
package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	ctrl0PhytofyPort           = 6000
	ctrl0HeartbeatInterval     = 10 * time.Second
	ctrl0CommissioningInterval = time.Minute
)

// Controls the PHYTOFY RL v0 fixtures
type ctrl0Controller struct {
	logger     *log.Logger
	networking *networking
	observer   *chan networkingObservation
	adapters   sync.Map
	modules    sync.Map
}

// Creates an instance of PHYTOFY RL v0 controller
func ctrl0Init(logger *log.Logger) *ctrl0Controller {
	networking := netInit(ctrl0PhytofyPort, logger)
	observer := networking.netAcquireChannel()
	controller := &ctrl0Controller{logger, networking, observer, sync.Map{}, sync.Map{}}
	go controller.ctrl0Process()
	go controller.ctrl0CommissioningRoutine()
	go controller.ctrl0HeartbeatRoutine()
	go controller.ctrl0ThermometerRoutine()
	go controller.ctrl0LoggerRoutine()
	go controller.ctrl0ForgettingRoutine()
	return controller
}

// Processes the incoming replies from the adapter
func (controller *ctrl0Controller) ctrl0Process() {
	for controller.networking.running {
		observation := <-*controller.observer
		adapterID := observation.address.String()
		controller.logger.Printf("INFO: [%s] -> %s", adapterID, string(*observation.buffer))
		if reply := pckt0DecodeReply(observation.buffer, controller.logger); reply != nil {
			if scheduleIDs, matched := pckt0ParseCommissioningReply(reply, controller.logger); matched {
				controller.ctrl0ProcessCommissioningReply(adapterID, scheduleIDs)
			} else if calibrations, matched := pckt0ParseModuleDataReply(reply, controller.logger); matched {
				controller.ctrl0ProcessModuleDataReply(adapterID, calibrations)
			} else if pckt0ParseHeartbeatReply(reply) {
				controller.logger.Printf("INFO: Received a heartbeat from %s", adapterID)
			} else if temperature, matched := pckt0ParseTemperatureReply(reply, controller.logger); matched {
				controller.logger.Printf("INFO: Temperature at %s - %d", adapterID, *temperature)
			} else if logs, matched := pckt0ParseLogContentReply(reply); matched {
				controller.logger.Printf("INFO: Hardware logs of %s - %s", adapterID, hex.EncodeToString(*logs))
			}
			controller.ctrl0UpdateLastSeen(adapterID)
		}
	}
}

func (controller *ctrl0Controller) ctrl0ProcessCommissioningReply(adapterID string, scheduleIDs *[]uint32) {
	controller.adapters.Store(adapterID, &dptr0Adapter{adapterID, *scheduleIDs, time.Now()})
	request := pckt0PrepareModuleDataRequest()
	if !controller.ctrl0Transmit(adapterID, request) {
		controller.logger.Printf("ERROR: Failed to transmit module data request to %s", adapterID)
	}
}

func (controller *ctrl0Controller) ctrl0ProcessModuleDataReply(adapterID string, calibrations *map[schdlSerial]pckt0Calibration) {
	for serial := range *calibrations {
		controller.modules.Store(serial, &dptr0Module{serial, adapterID, (*calibrations)[serial], time.Now()})
	}
}

// Wraps request into a complete message and sends it out
func (controller *ctrl0Controller) ctrl0Transmit(destination string, request *pckt0Request) bool {
	ipDestination := net.ParseIP(destination)
	if ipDestination == nil {
		controller.logger.Printf("ERROR: Failed to parse destination address %s, skipping transmission - %+v", destination, request)
		return false
	}
	request.Source = netMatchOwnAddress(ipDestination, controller.logger).String()
	request.Port = controller.networking.sourcePort
	request.Destination = destination
	buffer := pckt0EncodeRequest(request, controller.logger)
	if buffer == nil {
		return false
	}
	if !controller.networking.netTransmit(ipDestination, ctrl0PhytofyPort, buffer) {
		return false
	}
	controller.logger.Printf("INFO: [%s] <- %s", destination, string(*buffer))
	return true
}

// Broadcasts a request
func (controller *ctrl0Controller) ctrl0Broadcast(request *pckt0Request) bool {
	result := true
	for _, broadcast := range netBroadcasts(controller.logger) {
		result = result && controller.ctrl0Transmit(broadcast.String(), request)
	}
	return result
}

// Transmits the "set-leds" request to each relevant adapter
func (controller *ctrl0Controller) ctrl0TransmitLedsSetRequest(serial schdlSerial, levels schdlLevels) bool {
	module, present := controller.modules.Load(serial)
	if !present {
		return false
	}
	calibration := module.(*dptr0Module).calibration
	pwms := ctrl0LevelsIntoPwms(levels, calibration)
	request := pckt0PrepareLedsSetRequest(pwms, schdlSerials{serial})
	adapterID := module.(*dptr0Module).adapterID
	return controller.ctrl0Transmit(adapterID, request)
}

// Transmits the "schedule-add" request to each relevant adapter
func (controller *ctrl0Controller) ctrl0TransmitScheduleAddRequest(serial schdlSerial, schedule schdlDetached, scheduleID uint32) bool {
	module, present := controller.modules.Load(serial)
	if !present {
		return false
	}
	calibration := module.(*dptr0Module).calibration
	pwms := ctrl0LevelsIntoPwms(schedule.Levels, calibration)
	request := pckt0PrepareSchedulingSetRequest(int64(schedule.Start), int64(schedule.Stop), pwms, scheduleID, schdlSerials{serial})
	adapterID := module.(*dptr0Module).adapterID
	return controller.ctrl0Transmit(adapterID, request)
}

// Transmits the "schedule-clear" request to each relevant adapter
func (controller *ctrl0Controller) ctrl0TransmitScheduleClearRequests() bool {
	result := true
	controller.adapters.Range(func(key, value interface{}) bool {
		adapterID := key.(string)
		for _, scheduleID := range value.(*dptr0Adapter).scheduleIDs {
			request := pckt0PrepareSchedulingDeleteRequest(scheduleID)
			result = result && controller.ctrl0Transmit(adapterID, request)
		}
		return true
	})
	return result
}

// Waits for any serials to be present
func (controller *ctrl0Controller) ctrl0WaitForAnySerials(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		seen := controller.ctrl0GetSerials()
		if len(seen) != 0 {
			return true
		}
	}
	return false
}

// Waits for serials to be present
func (controller *ctrl0Controller) ctrl0WaitForSerials(serials schdlSerials, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	missing := true
	for time.Now().Before(deadline) {
		missing = false
		for _, serial := range serials {
			module, present := controller.modules.Load(serial)
			if !present {
				missing = true
				break
			}
			if time.Now().Sub(module.(*dptr0Module).lastSeen) > time.Minute {
				missing = true
				break
			}
			adapterID := module.(*dptr0Module).adapterID
			adapter, present := controller.adapters.Load(adapterID)
			if !present {
				missing = true
				break
			}
			if time.Now().Sub(adapter.(*dptr0Adapter).lastSeen) > time.Minute {
				missing = true
				break
			}
		}
		if missing {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	return !missing
}

// Separates the serials by the adapter they are behind
func (controller *ctrl0Controller) ctrl0SplitByAdapter(serials []uint32) map[string][]uint32 {
	separated := make(map[string][]uint32)
	for _, serial := range serials {
		module, present := controller.modules.Load(serial)
		if !present {
			controller.logger.Printf("ERROR: Could not find adapter for serial %d", serial)
			return nil
		}
		adapterID := module.(*dptr0Module).adapterID
		if group, present := separated[adapterID]; !present {
			separated[adapterID] = []uint32{serial}
		} else {
			group = append(group, serial)
		}
	}
	return separated
}

// Makes note of the last time and adapter sent back a reply
func (controller *ctrl0Controller) ctrl0UpdateLastSeen(adapterID string) {
	if adapter, present := controller.adapters.Load(adapterID); present {
		adapter.(*dptr0Adapter).lastSeen = time.Now()
	} else {
		controller.adapters.Store(adapterID, &dptr0Adapter{adapterID, []uint32{}, time.Now()})
	}
}

// Import schedules
func (controller *ctrl0Controller) ctrl0ImportSchedules(schedules []schdlAttached) error {
	aggregated, fail := schdlAggregateSchedules(schedules, true)
	if fail != nil {
		controller.logger.Printf("ERROR: Failed to aggregate schedules (%s)", fail)
		return fail
	}
	serials := make(schdlSerials, 0)
	for serial := range aggregated {
		serials = append(serials, serial)
	}
	if !controller.ctrl0WaitForSerials(serials, time.Minute) {
		fail := fmt.Errorf("Failed to locate all fixtures, seen - %v", controller.ctrl0GetSerials())
		controller.logger.Printf("ERROR: %s", fail)
		return fail
	}
	if !controller.ctrl0TransmitScheduleClearRequests() {
		fail := fmt.Errorf("Failed to transmit schedule clear requests")
		controller.logger.Printf("ERROR: %s", fail)
		return fail
	}
	for serial := range aggregated {
		schedules := aggregated[serial]
		for scheduleID, schedule := range schedules {
			if !controller.ctrl0TransmitScheduleAddRequest(serial, schedule, uint32(scheduleID)) {
				fail := fmt.Errorf("Failed to transmit schedule add request to %d - %+v", serial, schedule)
				controller.logger.Printf("ERROR: %s", fail)
				return fail
			}
		}
	}
	return nil
}

// Handles the "get-serials" command
func (controller *ctrl0Controller) ctrl0GetSerials() schdlSerials {
	serials := make(schdlSerials, 0)
	controller.modules.Range(func(key, value interface{}) bool {
		serial := key.(schdlSerial)
		serials = append(serials, serial)
		return true
	})
	sort.Slice(serials, func(i, j int) bool { return serials[i] < serials[j] })
	return serials
}

// Use channels' levels as PMW%
func ctrl0LevelsIntoPwms(levels schdlLevels, calibration pckt0Calibration) []uint8 {
	pwms := []uint8{0, 0, 0, 0, 0, 0, 0}
	for i, level := range levels {
		pwms[i] = uint8(level)
	}
	return pwms
}

// Routine periodically triggering the commissioning request
func (controller *ctrl0Controller) ctrl0CommissioningRoutine() {
	for controller.networking.running {
		request := pckt0PrepareCommissioningRequest()
		if !controller.ctrl0Broadcast(request) {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(ctrl0CommissioningInterval)
	}
}

// Routine periodically triggering the heartbeat request
func (controller *ctrl0Controller) ctrl0HeartbeatRoutine() {
	for controller.networking.running {
		request := pckt0PrepareHeartbeatRequest()
		if !controller.ctrl0Broadcast(request) {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(ctrl0HeartbeatInterval)
	}
}

// Routine periodically triggering the temperature request
func (controller *ctrl0Controller) ctrl0ThermometerRoutine() {
	for controller.networking.running {
		request := pckt0PrepareTemperatureRequest()
		if !controller.ctrl0Broadcast(request) {
			time.Sleep(time.Second)
			continue
		}
		time.Sleep(ctrl0HeartbeatInterval)
	}
}

// Routine periodically fetching the hardware logs
func (controller *ctrl0Controller) ctrl0LoggerRoutine() {
	then := int64(0)
	for controller.networking.running {
		now := time.Now().Unix()
		request := pckt0PrepareLogContentRequest(then, now)
		if !controller.ctrl0Broadcast(request) {
			time.Sleep(time.Second)
			continue
		}
		then = now
		time.Sleep(time.Hour)
	}
}

// Routine used by the discoverer to forget old adapters (it removes adapters which did not make contact for too long)
func (controller *ctrl0Controller) ctrl0ForgettingRoutine() {
	for controller.networking.running {
		now := time.Now()
		controller.modules.Range(func(key, value interface{}) bool {
			if now.Sub(value.(*dptr0Module).lastSeen) > time.Minute {
				controller.modules.Delete(key)
			}
			return true
		})
		controller.adapters.Range(func(key, value interface{}) bool {
			if now.Sub(value.(*dptr0Adapter).lastSeen) > time.Minute {
				controller.adapters.Delete(key)
			}
			return true
		})
		time.Sleep(time.Minute)
	}
}
