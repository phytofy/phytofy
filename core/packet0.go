// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for encoding/decoding packets for PHYTOFY RL v0 (DEPRECATED)
package main

import (
	"encoding/json"
	"log"
	"strconv"
	"time"
)

// A bare-bones request
type pckt0Request struct {
	Destination string
	Source      string
	Port        int
	ServiceType string
	MessageType string
	Payload     interface{}
}

type pckt0TemperatureRequestPayload struct {
	Target string
}

type pckt0CommissioningRequestPayload struct {
	ClockRefTime string
}

type pckt0SchedulingSetRequestPayload struct {
	ScheduleID    uint32
	StartDateTime string
	EndDateTime   string
	ModuleInfo    []pckt0ModuleInfoEntry
}

type pckt0SchedulingDeleteRequestPayload struct {
	ScheduleID uint32
}

type pckt0LedsSetRequestPayload struct {
	ModuleInfo []pckt0ModuleInfoEntry
}

type pckt0LogContentGetRequestPayload struct {
	StartDateTime string
	EndDateTime   string
}

type pckt0LogLevelSetRequestPayload struct {
	LogLevel int
}

type pckt0ModuleInfoEntry struct {
	ModuleID schdlSerial
	PWMInfo  pckt0PWMInfo
}

type pckt0PWMInfo struct {
	UV        uint8
	Blue      uint8
	Green     uint8
	HyperRed  uint8
	FarRed    uint8
	WarmWhite uint8
	EqWhite   uint8
}

// A bare-bones reply
type pckt0Reply struct {
	Destination string
	Source      string
	ServiceType string
	MessageType string
	Payload     json.RawMessage
}

type pckt0CommissioningReplyPayload struct {
	ScheduleIDs []uint32
}

type pckt0ModuleDataReplyPayload []struct {
	ChannelCalibration pckt0Calibration
	ModuleID           schdlSerial
	Version            int
}

type pckt0TemperatureReplyPayload struct {
	RoomTemp int
}

type pckt0Calibration [7][4]float64

// Prepares a request for heartbeat
func pckt0PrepareHeartbeatRequest() *pckt0Request {
	return &pckt0Request{"", "", 0, "Heartbeat", "Request", nil}
}

// Prepares a request for temperature
func pckt0PrepareTemperatureRequest() *pckt0Request {
	request := &pckt0Request{"", "", 0, "Temperature", "Request", nil}
	request.Payload = pckt0TemperatureRequestPayload{"room"}
	return request
}

// Prepares a request for module data
func pckt0PrepareModuleDataRequest() *pckt0Request {
	return &pckt0Request{"", "", 0, "Commissioning", "GetModuleData", nil}
}

// Prepares a request for module data
func pckt0PrepareCommissioningRequest() *pckt0Request {
	request := &pckt0Request{"", "", 0, "Commissioning", "Request", nil}
	request.Payload = pckt0CommissioningRequestPayload{pckt0ConvertTime(time.Now().Unix())}
	return request
}

// Prepares a request for reset
func pckt0PrepareResetRequest() *pckt0Request {
	return &pckt0Request{"", "", 0, "Reset", "Reboot", nil}
}

// Prepares a request for setting a schedule
func pckt0PrepareSchedulingSetRequest(scheduleStart, scheduleStop int64, calibratedLevels []uint8, scheduleID uint32, serials schdlSerials) *pckt0Request {
	request := &pckt0Request{"", "", 0, "Scheduling", "Set", nil}
	payload := pckt0SchedulingSetRequestPayload{
		scheduleID,
		pckt0ConvertTime(scheduleStart),
		pckt0ConvertTime(scheduleStop),
		make([]pckt0ModuleInfoEntry, 0),
	}
	for _, serial := range serials {
		pwmInfo := levelsToPWMInfo(calibratedLevels)
		entry := pckt0ModuleInfoEntry{serial, pwmInfo}
		payload.ModuleInfo = append(payload.ModuleInfo, entry)
	}
	request.Payload = payload
	return request
}

// Prepares a request for deleting a schedule
func pckt0PrepareSchedulingDeleteRequest(scheduleID uint32) *pckt0Request {
	request := &pckt0Request{"", "", 0, "Scheduling", "Delete", nil}
	request.Payload = pckt0SchedulingDeleteRequestPayload{scheduleID}
	return request
}

// Prepares a request for setting LEDs
func pckt0PrepareLedsSetRequest(calibratedLevels []uint8, serials schdlSerials) *pckt0Request {
	request := &pckt0Request{"", "", 0, "LightState", "Set", nil}
	payload := pckt0LedsSetRequestPayload{
		make([]pckt0ModuleInfoEntry, 0),
	}
	for _, serial := range serials {
		pwmInfo := levelsToPWMInfo(calibratedLevels)
		entry := pckt0ModuleInfoEntry{serial, pwmInfo}
		payload.ModuleInfo = append(payload.ModuleInfo, entry)
	}
	request.Payload = payload
	return request
}

// Prepares a request for hardware logs
func pckt0PrepareLogContentRequest(logStart, logStop int64) *pckt0Request {
	request := &pckt0Request{"", "", 0, "Logging", "Request", nil}
	payload := pckt0LogContentGetRequestPayload{
		pckt0ConvertTime(logStart),
		pckt0ConvertTime(logStop),
	}
	request.Payload = payload
	return request
}

// Prepares a request for setting a harware log level
func pckt0PrepareLogLevelSetRequest(level int) *pckt0Request {
	request := &pckt0Request{"", "", 0, "Logging", "Set", nil}
	payload := pckt0LogLevelSetRequestPayload{level}
	request.Payload = payload
	return request
}

// Converts UNIX epoch to Windows time
func pckt0ConvertTime(stamp int64) string {
	return strconv.FormatInt(stamp*10000000+116444736000000000, 10)
}

// Parses a commissioning reply
func pckt0ParseCommissioningReply(reply *pckt0Reply, logger *log.Logger) (*[]uint32, bool) {
	if reply.ServiceType == "Commissioning" && reply.MessageType == "Reply" {
		var actual pckt0CommissioningReplyPayload
		if fail := json.Unmarshal(reply.Payload, &actual); fail != nil {
			logger.Printf("ERROR: Failed to unmarshal commissioning reply payload (%s)", fail)
			return nil, false
		}
		return &(actual.ScheduleIDs), true
	}
	return nil, false
}

// Parses a module data reply
func pckt0ParseModuleDataReply(reply *pckt0Reply, logger *log.Logger) (*map[schdlSerial]pckt0Calibration, bool) {
	if reply.ServiceType == "Commissioning" && reply.MessageType == "ReplyModuleData" {
		var actual pckt0ModuleDataReplyPayload
		if fail := json.Unmarshal(reply.Payload, &actual); fail != nil {
			logger.Printf("ERROR: Failed to unmarshal module data reply payload (%s)", fail)
			return nil, false
		}
		calibrations := make(map[schdlSerial]pckt0Calibration)
		for _, item := range actual {
			calibrations[item.ModuleID] = item.ChannelCalibration
		}
		return &calibrations, true
	}
	return nil, false
}

// Parses a temperature reply
func pckt0ParseTemperatureReply(reply *pckt0Reply, logger *log.Logger) (*int, bool) {
	if reply.ServiceType == "Temperature" && reply.MessageType == "Reply" {
		var actual pckt0TemperatureReplyPayload
		if fail := json.Unmarshal(reply.Payload, &actual); fail != nil {
			logger.Printf("ERROR: Failed to unmarshal temperature reply payload (%s)", fail)
			return nil, false
		}
		temperature := actual.RoomTemp
		return &temperature, true
	}
	return nil, false
}

// Parses a heartbeat reply
func pckt0ParseHeartbeatReply(reply *pckt0Reply) bool {
	return reply.ServiceType == "Heartbeat" && reply.MessageType == "Reply"
}

// Parses a hardware logs reply
func pckt0ParseLogContentReply(reply *pckt0Reply) (*json.RawMessage, bool) {
	if reply.ServiceType == "Logging" && reply.MessageType == "Reply" {
		return &reply.Payload, true
	}
	return nil, false
}

func pckt0EncodeRequest(request *pckt0Request, logger *log.Logger) *[]byte {
	buffer, fail := json.Marshal(*request)
	if fail != nil {
		logger.Printf("ERROR: Failed to encode a request (%s), skipping - %+v", fail, *request)
		return nil
	}
	return &buffer
}

func pckt0DecodeReply(buffer *[]byte, logger *log.Logger) *pckt0Reply {
	var reply pckt0Reply
	if fail := json.Unmarshal(*buffer, &reply); fail != nil {
		logger.Printf("ERROR: Failed to decode a reply (%s), skipping", fail)
		return nil
	}
	return &reply
}

// Converts levels to an apropriate structure
func levelsToPWMInfo(calibratedLevels []uint8) pckt0PWMInfo {
	pwmUv := uint8(0)
	if len(calibratedLevels) > 0 {
		pwmUv = calibratedLevels[0]
	}
	pwmBlue := uint8(0)
	if len(calibratedLevels) > 1 {
		pwmBlue = calibratedLevels[1]
	}
	pwmGreen := uint8(0)
	if len(calibratedLevels) > 2 {
		pwmGreen = calibratedLevels[2]
	}
	pwmHyperRed := uint8(0)
	if len(calibratedLevels) > 3 {
		pwmHyperRed = calibratedLevels[3]
	}
	pwmFarRed := uint8(0)
	if len(calibratedLevels) > 4 {
		pwmFarRed = calibratedLevels[4]
	}
	pwmWarmWhite := uint8(0)
	if len(calibratedLevels) > 5 {
		pwmWarmWhite = calibratedLevels[5]
	}
	pwmEqWhite := uint8(0)
	if len(calibratedLevels) > 6 {
		pwmEqWhite = calibratedLevels[6]
	}
	pwmInfo := pckt0PWMInfo{pwmUv, pwmBlue, pwmGreen, pwmHyperRed, pwmFarRed, pwmWarmWhite, pwmEqWhite}
	return pwmInfo
}
