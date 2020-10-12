// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for encoding/decoding packets
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
)

type pckt1FunctionCode uint8

const (
	pckt1FunctionCodeSetModuleCalibration          = pckt1FunctionCode(0)
	pckt1FunctionCodeGetModuleCalibration          = pckt1FunctionCode(1)
	pckt1FunctionCodeSetSerialNumber               = pckt1FunctionCode(2)
	pckt1FunctionCodeGetSerialNumber               = pckt1FunctionCode(3)
	pckt1FunctionCodeSetShortAddress               = pckt1FunctionCode(4)
	pckt1FunctionCodeGetShortAddress               = pckt1FunctionCode(5)
	pckt1FunctionCodeSetGroupID                    = pckt1FunctionCode(6)
	pckt1FunctionCodeGetGroupID                    = pckt1FunctionCode(7)
	pckt1FunctionCodeSetFixtureInfo                = pckt1FunctionCode(8)
	pckt1FunctionCodeGetFixtureInfo                = pckt1FunctionCode(9)
	pckt1FunctionCodeSetTimeReference              = pckt1FunctionCode(10)
	pckt1FunctionCodeGetTimeReference              = pckt1FunctionCode(11)
	pckt1FunctionCodeSetLEDs                       = pckt1FunctionCode(12)
	pckt1FunctionCodeGetLEDs                       = pckt1FunctionCode(13)
	pckt1FunctionCodeSetSchedule                   = pckt1FunctionCode(14)
	pckt1FunctionCodeGetSchedule                   = pckt1FunctionCode(15)
	pckt1FunctionCodeGetScheduleCount              = pckt1FunctionCode(16)
	pckt1FunctionCodeGetSchedulingState            = pckt1FunctionCode(17)
	pckt1FunctionCodeDeleteSchedule                = pckt1FunctionCode(18)
	pckt1FunctionCodeDeleteAllSchedules            = pckt1FunctionCode(19)
	pckt1FunctionCodeStopScheduling                = pckt1FunctionCode(20)
	pckt1FunctionCodeResumeScheduling              = pckt1FunctionCode(21)
	pckt1FunctionCodeSetIlluminanceConfiguration   = pckt1FunctionCode(24)
	pckt1FunctionCodeGetIlluminanceConfiguration   = pckt1FunctionCode(25)
	pckt1FunctionCodeGetModuleTemperature          = pckt1FunctionCode(26)
	pckt1FunctionCodeToggleCalibration             = pckt1FunctionCode(27)
	pckt1FunctionCodeResetForFirmwareUpdate        = pckt1FunctionCode(200)
	pckt1FunctionCodeConfirmResetForFirmwareUpdate = pckt1FunctionCode(201)
)

var pckt1KnownCodes = []pckt1FunctionCode{
	pckt1FunctionCodeSetModuleCalibration,
	pckt1FunctionCodeGetModuleCalibration,
	pckt1FunctionCodeSetSerialNumber,
	pckt1FunctionCodeGetSerialNumber,
	pckt1FunctionCodeSetShortAddress,
	pckt1FunctionCodeGetShortAddress,
	pckt1FunctionCodeSetGroupID,
	pckt1FunctionCodeGetGroupID,
	pckt1FunctionCodeSetFixtureInfo,
	pckt1FunctionCodeGetFixtureInfo,
	pckt1FunctionCodeSetTimeReference,
	pckt1FunctionCodeGetTimeReference,
	pckt1FunctionCodeSetLEDs,
	pckt1FunctionCodeGetLEDs,
	pckt1FunctionCodeSetSchedule,
	pckt1FunctionCodeGetSchedule,
	pckt1FunctionCodeGetScheduleCount,
	pckt1FunctionCodeGetSchedulingState,
	pckt1FunctionCodeDeleteSchedule,
	pckt1FunctionCodeDeleteAllSchedules,
	pckt1FunctionCodeStopScheduling,
	pckt1FunctionCodeResumeScheduling,
	pckt1FunctionCodeSetIlluminanceConfiguration,
	pckt1FunctionCodeGetIlluminanceConfiguration,
	pckt1FunctionCodeGetModuleTemperature,
	pckt1FunctionCodeToggleCalibration,
	pckt1FunctionCodeResetForFirmwareUpdate,
	pckt1FunctionCodeConfirmResetForFirmwareUpdate,
}

type pckt1ShortAddress uint8

const (
	pckt1ShortAddressBroadcast     = pckt1ShortAddress(0)
	pckt1ShortAddressBegin         = pckt1ShortAddress(1)
	pckt1ShortAddressEnd           = pckt1ShortAddress(247)
	pckt1ShortAddressReservedBegin = pckt1ShortAddress(248)
	pckt1ShortAddressReservedEnd   = pckt1ShortAddress(254)
	pckt1ShortAddressUnassigned    = pckt1ShortAddress(255)
)

const (
	pckt1SchedulerStopped         = 0
	pckt1SchedulerRunningNothing  = 1
	pckt1SchedulerRunningSchedule = 2
)

const (
	pckt1ScheduleSearchByID    = 0
	pckt1ScheduleSearchByIndex = 1
)

const (
	pckt1LEDsModule0Mask     = uint8(0x1)
	pckt1LEDsModule0Enabled  = uint8(0x1)
	pckt1LEDsModule0Disabled = uint8(0x0)
	pckt1LEDsModule1Mask     = uint8(0x2)
	pckt1LEDsModule1Enabled  = uint8(0x2)
	pckt1LEDsModule1Disabled = uint8(0x0)
	pckt1UseMask             = uint8(0x4)
	pckt1UseIrradiance       = uint8(0x4)
	pckt1UsePWM              = uint8(0x0)
)

type pckt1Calibration [6]struct {
	CoefficientA float32 `json:"coefficient_a"`
	CoefficientB float32 `json:"coefficient_b"`
	CoefficientM float32 `json:"coefficient_m"`
}

type pckt1Packet struct {
	Header  pckt1Header  `json:"header"`
	Payload pckt1Payload `json:"payload"`
}

type pckt1Header struct {
	ClientIPv4     [4]byte           `json:"client_ipv4"`
	SequenceNumber uint32            `json:"sequence_number"`
	ShortAddress   pckt1ShortAddress `json:"short_address"`
	FunctionCode   pckt1FunctionCode `json:"function_code"`
}

type pckt1Payload interface{}

type pckt1CommandPayloadSetModuleCalibration struct {
	ModuleID    uint8            `json:"module_id"`
	Calibration pckt1Calibration `json:"calibration"`
}

type pckt1CommandPayloadGetModuleCalibration struct {
	ModuleID uint8 `json:"module_id"`
}

type pckt1CommandPayloadSetSerialNumber struct {
	Serial schdlSerial `json:"serial"`
}

type pckt1CommandPayloadGetSerialNumber struct {
	RandomBackOff bool `json:"random_backoff"`
}

type pckt1CommandPayloadSetShortAddress struct {
	Serial       schdlSerial       `json:"serial"`
	ShortAddress pckt1ShortAddress `json:"short_address"`
}

type pckt1CommandPayloadGetShortAddress struct {
	Serial schdlSerial `json:"serial"`
}

type pckt1CommandPayloadSetGroupID struct {
	GroupID uint32 `json:"group_id"`
}

type pckt1CommandPayloadSetFixtureInfo struct {
	FWVersion uint32 `json:"fw_version"`
	HWVersion uint32 `json:"hw_version"`
}

type pckt1CommandPayloadSetTimeReference struct {
	LinuxEpoch uint32 `json:"linux_epoch"`
}

type pckt1CommandPayloadSetLEDsPWM struct {
	Config uint8     `json:"config"`
	Levels [6]uint32 `json:"levels"`
}

type pckt1CommandPayloadSetLEDsIrradiance struct {
	Config uint8      `json:"config"`
	Levels [6]float32 `json:"levels"`
}

type pckt1CommandPayloadGetLEDs struct {
	Config uint8 `json:"config"`
}

type pckt1CommandPayloadSetSchedulePreamble struct {
	ScheduleID uint32 `json:"schedule_id"`
	Start      uint32 `json:"start"`
	Stop       uint32 `json:"stop"`
	Config     uint8  `json:"config"`
}

type pckt1CommandPayloadSetSchedulePWM struct {
	pckt1CommandPayloadSetSchedulePreamble
	Levels [6]uint32 `json:"levels"`
}

type pckt1CommandPayloadSetScheduleIrradiance struct {
	pckt1CommandPayloadSetSchedulePreamble
	Levels [6]float32 `json:"levels"`
}

type pckt1CommandPayloadGetSchedule struct {
	ScheduleKey     uint32 `json:"schedule_key"`
	ScheduleKeyType uint8  `json:"schedule_key_type"`
}

type pckt1CommandPayloadDeleteSchedule struct {
	ScheduleID uint32 `json:"schedule_id"`
}

type pckt1CommandPayloadSetIlluminanceConfiguration struct {
	Configuration [6]float32 `json:"configuration"`
}

type pckt1CommandPayloadToggleCalibration struct {
	CalibrationEnabled bool `json:"calibration_enabled"`
}

type pckt1ReplyPayloadPreamble struct {
	Ack bool `json:"ack"`
}

type pckt1ReplyPayloadGenericOK struct {
	pckt1ReplyPayloadPreamble
}

type pckt1ReplyPayloadGenericNOK struct {
	pckt1ReplyPayloadPreamble
	ErrorCode uint8 `json:"error_code"`
}

type pckt1ReplyPayloadGetModuleCalibration struct {
	ModuleID    uint8            `json:"module_id"`
	Calibration pckt1Calibration `json:"calibration"`
}

type pckt1ReplyPayloadGetSerialNumber struct {
	Serial schdlSerial `json:"serial"`
}

type pckt1ReplyPayloadGetShortAddress struct {
	ShortAddress pckt1ShortAddress `json:"short_address"`
	Serial       schdlSerial       `json:"serial"`
}

type pckt1ReplyPayloadGetGroupID struct {
	GroupID uint32 `json:"group_id"`
}

type pckt1ReplyPayloadGetFixtureInfo struct {
	FWVersion uint32     `json:"fw_version"`
	HWVersion uint32     `json:"hw_version"`
	Max       [6]float32 `json:"max"`
}

type pckt1ReplyPayloadGetTimeReference struct {
	LinuxEpoch uint32 `json:"linux_epoch"`
}

type pckt1ReplyPayloadGetLEDsPreamble struct {
	Config uint8 `json:"config"`
}

type pckt1ReplyPayloadGetLEDsPWM struct {
	pckt1ReplyPayloadGetLEDsPreamble
	Levels [6]uint32 `json:"levels"`
}

type pckt1ReplyPayloadGetLEDsIrradiance struct {
	pckt1ReplyPayloadGetLEDsPreamble
	Levels [6]float32 `json:"levels"`
}

type pckt1ReplyPayloadGetSchedulePreamble struct {
	ScheduleID uint32 `json:"schedule_id"`
	Start      uint32 `json:"start"`
	Stop       uint32 `json:"stop"`
	Config     uint8  `json:"config"`
}

type pckt1ReplyPayloadGetSchedulePWM struct {
	pckt1ReplyPayloadGetSchedulePreamble
	Levels [6]uint32 `json:"levels"`
}

type pckt1ReplyPayloadGetScheduleIrradiance struct {
	pckt1ReplyPayloadGetSchedulePreamble
	Levels [6]float32 `json:"levels"`
}

type pckt1ReplyPayloadGetScheduleCount struct {
	ScheduleCount uint32 `json:"schedule_count"`
}

type pckt1ReplyPayloadGetSchedulingState struct {
	SchedulingState uint8  `json:"scheduling_state"`
	ScheduleID      uint32 `json:"schedule_id"`
}

type pckt1ReplyPayloadGetIlluminanceConfiguration struct {
	Configuration [6]float32 `json:"configuration"`
}

type pckt1ReplyPayloadGetModuleTemperature struct {
	Temperatures0 [6]float32 `json:"temperatures_0"`
	Temperatures1 [6]float32 `json:"temperatures_1"`
}

type pckt1ReplyPayloadToggleCalibration struct {
	Ack bool `json:"ack"`
}

const (
	pckt1HeaderSize = 10
	pckt1CRC16Size  = 2
)

var (
	// Table of CRC values for high–order byte
	pckt1CRCLUTHi = []uint8{
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x00, 0xC1, 0x81, 0x40, 0x01, 0xC0, 0x80, 0x41,
		0x01, 0xC0, 0x80, 0x41, 0x00, 0xC1, 0x81, 0x40,
	}

	// Table of CRC values for low–order byte
	pckt1CRCLUTLo = []uint8{
		0x00, 0xC0, 0xC1, 0x01, 0xC3, 0x03, 0x02, 0xC2,
		0xC6, 0x06, 0x07, 0xC7, 0x05, 0xC5, 0xC4, 0x04,
		0xCC, 0x0C, 0x0D, 0xCD, 0x0F, 0xCF, 0xCE, 0x0E,
		0x0A, 0xCA, 0xCB, 0x0B, 0xC9, 0x09, 0x08, 0xC8,
		0xD8, 0x18, 0x19, 0xD9, 0x1B, 0xDB, 0xDA, 0x1A,
		0x1E, 0xDE, 0xDF, 0x1F, 0xDD, 0x1D, 0x1C, 0xDC,
		0x14, 0xD4, 0xD5, 0x15, 0xD7, 0x17, 0x16, 0xD6,
		0xD2, 0x12, 0x13, 0xD3, 0x11, 0xD1, 0xD0, 0x10,
		0xF0, 0x30, 0x31, 0xF1, 0x33, 0xF3, 0xF2, 0x32,
		0x36, 0xF6, 0xF7, 0x37, 0xF5, 0x35, 0x34, 0xF4,
		0x3C, 0xFC, 0xFD, 0x3D, 0xFF, 0x3F, 0x3E, 0xFE,
		0xFA, 0x3A, 0x3B, 0xFB, 0x39, 0xF9, 0xF8, 0x38,
		0x28, 0xE8, 0xE9, 0x29, 0xEB, 0x2B, 0x2A, 0xEA,
		0xEE, 0x2E, 0x2F, 0xEF, 0x2D, 0xED, 0xEC, 0x2C,
		0xE4, 0x24, 0x25, 0xE5, 0x27, 0xE7, 0xE6, 0x26,
		0x22, 0xE2, 0xE3, 0x23, 0xE1, 0x21, 0x20, 0xE0,
		0xA0, 0x60, 0x61, 0xA1, 0x63, 0xA3, 0xA2, 0x62,
		0x66, 0xA6, 0xA7, 0x67, 0xA5, 0x65, 0x64, 0xA4,
		0x6C, 0xAC, 0xAD, 0x6D, 0xAF, 0x6F, 0x6E, 0xAE,
		0xAA, 0x6A, 0x6B, 0xAB, 0x69, 0xA9, 0xA8, 0x68,
		0x78, 0xB8, 0xB9, 0x79, 0xBB, 0x7B, 0x7A, 0xBA,
		0xBE, 0x7E, 0x7F, 0xBF, 0x7D, 0xBD, 0xBC, 0x7C,
		0xB4, 0x74, 0x75, 0xB5, 0x77, 0xB7, 0xB6, 0x76,
		0x72, 0xB2, 0xB3, 0x73, 0xB1, 0x71, 0x70, 0xB0,
		0x50, 0x90, 0x91, 0x51, 0x93, 0x53, 0x52, 0x92,
		0x96, 0x56, 0x57, 0x97, 0x55, 0x95, 0x94, 0x54,
		0x9C, 0x5C, 0x5D, 0x9D, 0x5F, 0x9F, 0x9E, 0x5E,
		0x5A, 0x9A, 0x9B, 0x5B, 0x99, 0x59, 0x58, 0x98,
		0x88, 0x48, 0x49, 0x89, 0x4B, 0x8B, 0x8A, 0x4A,
		0x4E, 0x8E, 0x8F, 0x4F, 0x8D, 0x4D, 0x4C, 0x8C,
		0x44, 0x84, 0x85, 0x45, 0x87, 0x47, 0x46, 0x86,
		0x82, 0x42, 0x43, 0x83, 0x41, 0x81, 0x80, 0x40,
	}
)

// Calculates CRC 16
func pckt1CRC16(octets []byte) uint16 {
	crcHi := uint8(0xFF)           // high byte of CRC initialized
	crcLo := uint8(0xFF)           // low byte of CRC initialized
	for _, octet := range octets { // pass through octets
		index := int(crcHi ^ octet) // calculate the CRC
		crcHi = crcLo ^ pckt1CRCLUTHi[index]
		crcLo = pckt1CRCLUTLo[index]
	}
	return uint16(crcHi)<<8 | uint16(crcLo)
}

// Decodes header from binary
func pckt1DecodeHeader(octets []byte) (*pckt1Header, error) {
	var header pckt1Header
	if fail := binary.Read(bytes.NewBuffer(octets), binary.LittleEndian, &header); fail != nil {
		return nil, fail
	}
	return &header, nil
}

// Encodes header to binary
func pckt1EncodeHeader(header pckt1Header) ([]byte, error) {
	var octets bytes.Buffer
	if fail := binary.Write(&octets, binary.LittleEndian, header); fail != nil {
		return nil, fail
	}
	return octets.Bytes(), nil
}

// Decodes payload from binary
func pckt1DecodePayload(octets []byte, header pckt1Header) (pckt1Payload, error) {
	payload, _, fail := pckt1PrepareReplyPayload(octets, header)
	if fail != nil {
		return nil, fail
	}
	if fail := binary.Read(bytes.NewBuffer(octets), binary.LittleEndian, payload); fail != nil {
		return nil, fail
	}
	return payload, nil
}

// Encodes payload to binary
func pckt1EncodePayload(payload pckt1Payload) ([]byte, error) {
	var octets bytes.Buffer
	if payload != nil {
		if fail := binary.Write(&octets, binary.LittleEndian, payload); fail != nil {
			return nil, fail
		}
	}
	return octets.Bytes(), nil
}

// Encodes the packet into binary
func pckt1Encode(packet pckt1Packet) ([]byte, error) {
	var octets bytes.Buffer
	encodedHeader, fail := pckt1EncodeHeader(packet.Header)
	if fail != nil {
		return nil, fmt.Errorf("Failed to encode header (%s)", fail)
	}
	if writtenHeader, fail := octets.Write(encodedHeader); fail != nil {
		return nil, fmt.Errorf("Failed to write header (%s)", fail)
	} else if writtenHeader != len(encodedHeader) {
		return nil, fmt.Errorf("Failed to write full header - %+v", packet.Header)
	}
	encodedPayload, fail := pckt1EncodePayload(packet.Payload)
	if fail != nil {
		return nil, fmt.Errorf("Failed to encode payload (%s)", fail)
	}
	if writtenPayload, fail := octets.Write(encodedPayload); fail != nil {
		return nil, fmt.Errorf("Failed to write payload (%s)", fail)
	} else if writtenPayload != len(encodedPayload) {
		return nil, fmt.Errorf("Failed to write full payload - %+v", packet.Payload)
	}
	crc16 := pckt1CRC16(octets.Bytes())
	if fail := binary.Write(&octets, binary.LittleEndian, crc16); fail != nil {
		return nil, fmt.Errorf("Failed to write CRC (%s)", fail)
	}
	return octets.Bytes(), nil
}

// Decodes a packet from binary
func pckt1Decode(octets []byte) (*pckt1Packet, error) {
	header, fail := pckt1DecodeHeader(octets)
	if fail != nil {
		return nil, fmt.Errorf("Failed to decode header (%s)", fail)
	}
	payload, fail := pckt1DecodePayload(octets[pckt1HeaderSize:], *header)
	if fail != nil {
		return nil, fmt.Errorf("Failed to decode payload (%s)", fail)
	}
	packet := pckt1Packet{*header, payload}
	return &packet, nil
}

// Decodes CRC from binary at the given offset
func pckt1DecodeCRC16(octets []byte) (uint16, error) {
	var crc16 uint16
	if fail := binary.Read(bytes.NewBuffer(octets), binary.LittleEndian, &crc16); fail != nil {
		return 0, fmt.Errorf("Failed to read CRC (%s)", fail)
	}
	return crc16, nil
}

func pckt1Skip(buffer *bytes.Buffer, skip int, logger *log.Logger) {
	dropped := make([]byte, skip)
	read, fail := buffer.Read(dropped)
	if fail != nil {
		logger.Panicf("CRITICAL: Failed to skip %d bytes while parsing (%s)", len(dropped), fail)
		return
	}
	if read < len(dropped) {
		logger.Panicf("CRITICAL: Failed to skip %d (of %d) bytes while parsing", len(dropped)-read, len(dropped))
	}
}

// Parses all available packets
func pckt1Parse(buffer *bytes.Buffer, identifier string, logger *log.Logger) []pckt1Packet {
	packets := make([]pckt1Packet, 0)
	for {
		octets := buffer.Bytes()
		if len(octets) < pckt1HeaderSize+1 {
			break
		}
		header, fail := pckt1DecodeHeader(octets)
		if fail != nil {
			logger.Printf("ERROR: [%s] Failed to decode header (%s); Skipping %02x", identifier, fail, octets[0])
			pckt1Skip(buffer, 1, logger)
			continue
		}
		code := header.FunctionCode
		if !pckt1KnownCode(code) {
			logger.Printf("WARNING: [%s] Bad function code (%d); Skipping %02x", identifier, code, octets[0])
			pckt1Skip(buffer, 1, logger)
			continue
		}
		if len(octets) < pckt1HeaderSize+pckt1LookupPayloadSizeUntilVariantDifferentiator(code) {
			break
		}
		prepared, payloadSize, fail := pckt1PrepareReplyPayload(octets[pckt1HeaderSize:], *header)
		if fail != nil {
			logger.Printf("WARNING: [%s] Bad variant (%s); Skipping %02x", identifier, fail, octets[0])
			pckt1Skip(buffer, 1, logger)
			continue
		}
		if prepared == nil {
			logger.Printf("WARNING: [%s] Bad variant; Skipping %02x", identifier, octets[0])
			pckt1Skip(buffer, 1, logger)
			continue
		}
		size := pckt1HeaderSize + payloadSize + pckt1CRC16Size
		if len(octets) < size {
			break
		}
		crc16, fail := pckt1DecodeCRC16(octets[size-pckt1CRC16Size:])
		if fail != nil {
			logger.Printf("ERROR: [%s] Failed to decode CRC (%s); Skipping %02x", identifier, fail, octets[0])
			pckt1Skip(buffer, 1, logger)
			continue
		}
		if pckt1CRC16(octets[:size-pckt1CRC16Size]) == crc16 {
			packet, fail := pckt1Decode(octets)
			if fail != nil {
				logger.Printf("ERROR: [%s] Failed to decode packet (%s); Skipping %02x", identifier, fail, octets[0])
				pckt1Skip(buffer, 1, logger)
				continue
			}
			packets = append(packets, *packet)
			logger.Printf("INFO: [%s] -> %s", identifier, hex.EncodeToString(octets[:size]))
			pckt1Skip(buffer, size, logger)
		} else {
			logger.Printf("WARNING: [%s] Bad checksum; Skipping %02x", identifier, octets[0])
			pckt1Skip(buffer, 1, logger)
		}
	}
	return packets
}

func pckt1PrepareGenericReplyPayload(octets []byte) (pckt1Payload, int) {
	if len(octets) > 0 {
		if octets[0] == 1 {
			return new(pckt1ReplyPayloadGenericOK), binary.Size(pckt1ReplyPayloadGenericOK{})
		} else if octets[0] == 0 && len(octets) > 1 {
			return new(pckt1ReplyPayloadGenericNOK), binary.Size(pckt1ReplyPayloadGenericNOK{})
		}
	}
	return nil, 0
}

func pckt1PrepareReplyPayload(octets []byte, header pckt1Header) (pckt1Payload, int, error) {
	payload := pckt1Payload(nil)
	size := 0
	switch header.FunctionCode {
	case pckt1FunctionCodeSetModuleCalibration:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetModuleCalibration:
		payload = new(pckt1ReplyPayloadGetModuleCalibration)
		size = binary.Size(pckt1ReplyPayloadGetModuleCalibration{})
	case pckt1FunctionCodeSetSerialNumber:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetSerialNumber:
		payload = new(pckt1ReplyPayloadGetSerialNumber)
		size = binary.Size(pckt1ReplyPayloadGetSerialNumber{})
	case pckt1FunctionCodeSetShortAddress:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetShortAddress:
		payload = new(pckt1ReplyPayloadGetShortAddress)
		size = binary.Size(pckt1ReplyPayloadGetShortAddress{})
	case pckt1FunctionCodeSetGroupID:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetGroupID:
		payload = new(pckt1ReplyPayloadGetGroupID)
		size = binary.Size(pckt1ReplyPayloadGetGroupID{})
	case pckt1FunctionCodeSetFixtureInfo:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetFixtureInfo:
		payload = new(pckt1ReplyPayloadGetFixtureInfo)
		size = binary.Size(pckt1ReplyPayloadGetFixtureInfo{})
	case pckt1FunctionCodeSetTimeReference:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetTimeReference:
		payload = new(pckt1ReplyPayloadGetTimeReference)
		size = binary.Size(pckt1ReplyPayloadGetTimeReference{})
	case pckt1FunctionCodeSetLEDs:
		return nil, 0, fmt.Errorf("Missing payload format for function code - %d", pckt1FunctionCodeSetLEDs)
	case pckt1FunctionCodeGetLEDs:
		if len(octets) > 0 {
			switch octets[0] & pckt1UseMask {
			case pckt1UsePWM:
				payload = new(pckt1ReplyPayloadGetLEDsPWM)
				size = binary.Size(pckt1ReplyPayloadGetLEDsPWM{})
			case pckt1UseIrradiance:
				payload = new(pckt1ReplyPayloadGetLEDsIrradiance)
				size = binary.Size(pckt1ReplyPayloadGetLEDsIrradiance{})
			}
		}
	case pckt1FunctionCodeSetSchedule:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetSchedule:
		if len(octets) > 12 {
			switch octets[12] & pckt1UseMask {
			case pckt1UsePWM:
				payload = new(pckt1ReplyPayloadGetSchedulePWM)
				size = binary.Size(pckt1ReplyPayloadGetSchedulePWM{})
			case pckt1UseIrradiance:
				payload = new(pckt1ReplyPayloadGetScheduleIrradiance)
				size = binary.Size(pckt1ReplyPayloadGetScheduleIrradiance{})
			}
		}
	case pckt1FunctionCodeGetScheduleCount:
		payload = new(pckt1ReplyPayloadGetScheduleCount)
		size = binary.Size(pckt1ReplyPayloadGetScheduleCount{})
	case pckt1FunctionCodeGetSchedulingState:
		payload = new(pckt1ReplyPayloadGetSchedulingState)
		size = binary.Size(pckt1ReplyPayloadGetSchedulingState{})
	case pckt1FunctionCodeDeleteSchedule:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeDeleteAllSchedules:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeStopScheduling:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeResumeScheduling:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeSetIlluminanceConfiguration:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeGetIlluminanceConfiguration:
		payload = new(pckt1ReplyPayloadGetIlluminanceConfiguration)
		size = binary.Size(pckt1ReplyPayloadGetIlluminanceConfiguration{})
	case pckt1FunctionCodeGetModuleTemperature:
		payload = new(pckt1ReplyPayloadGetModuleTemperature)
		size = binary.Size(pckt1ReplyPayloadGetModuleTemperature{})
	case pckt1FunctionCodeToggleCalibration:
		payload = new(pckt1ReplyPayloadToggleCalibration)
		size = binary.Size(pckt1ReplyPayloadToggleCalibration{})
	case pckt1FunctionCodeResetForFirmwareUpdate:
		payload, size = pckt1PrepareGenericReplyPayload(octets)
	case pckt1FunctionCodeConfirmResetForFirmwareUpdate:
		return nil, 0, fmt.Errorf("Missing payload format for function code - %d", pckt1FunctionCodeConfirmResetForFirmwareUpdate)
	default:
		return nil, 0, fmt.Errorf("Unknown payload format for function code - %d", header.FunctionCode)
	}
	if payload == nil {
		return nil, 0, fmt.Errorf("Unknown payload format variant for function code - %d", header.FunctionCode)
	}
	return payload, size, nil
}

// Checks if given function code is implemented
func pckt1KnownCode(code pckt1FunctionCode) bool {
	for _, known := range pckt1KnownCodes {
		if code == known {
			return true
		}
	}
	return false
}

// Looks up the offset of the variant diffrentiator in the payload
func pckt1LookupPayloadSizeUntilVariantDifferentiator(code pckt1FunctionCode) int {
	switch code {
	case pckt1FunctionCodeGetLEDs:
		return 1
	case pckt1FunctionCodeGetSchedule:
		return 13
	default:
		return 0
	}
}

// Tells if a command gets a reply
func pckt1IsReplying(code pckt1FunctionCode) bool {
	switch code {
	case pckt1FunctionCodeSetLEDs, pckt1FunctionCodeConfirmResetForFirmwareUpdate:
		return false
	}
	return true
}

// Convert packet to string
func pckt1ToString(packet pckt1Packet) string {
	header := fmt.Sprintf("%+v", packet.Header)
	payload := "<unknown>"
	if packet.Payload == nil {
		payload = "<nil>"
	} else {
		switch packet.Payload.(type) {
		case *pckt1CommandPayloadSetModuleCalibration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetModuleCalibration)))

		case *pckt1CommandPayloadGetModuleCalibration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadGetModuleCalibration)))
		case *pckt1CommandPayloadSetSerialNumber:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetSerialNumber)))
		case *pckt1CommandPayloadGetSerialNumber:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadGetSerialNumber)))
		case *pckt1CommandPayloadSetShortAddress:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetShortAddress)))
		case *pckt1CommandPayloadGetShortAddress:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadGetShortAddress)))
		case *pckt1CommandPayloadSetGroupID:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetGroupID)))
		case *pckt1CommandPayloadSetFixtureInfo:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetFixtureInfo)))
		case *pckt1CommandPayloadSetTimeReference:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetTimeReference)))
		case *pckt1CommandPayloadSetLEDsPWM:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetLEDsPWM)))
		case *pckt1CommandPayloadSetLEDsIrradiance:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetLEDsIrradiance)))
		case *pckt1CommandPayloadGetLEDs:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadGetLEDs)))
		case *pckt1CommandPayloadSetSchedulePWM:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetSchedulePWM)))
		case *pckt1CommandPayloadSetScheduleIrradiance:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetScheduleIrradiance)))
		case *pckt1CommandPayloadGetSchedule:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadGetSchedule)))
		case *pckt1CommandPayloadDeleteSchedule:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadDeleteSchedule)))
		case *pckt1CommandPayloadSetIlluminanceConfiguration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadSetIlluminanceConfiguration)))
		case *pckt1CommandPayloadToggleCalibration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1CommandPayloadToggleCalibration)))
		case *pckt1ReplyPayloadGenericOK:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGenericOK)))
		case *pckt1ReplyPayloadGenericNOK:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGenericNOK)))
		case *pckt1ReplyPayloadGetModuleCalibration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetModuleCalibration)))
		case *pckt1ReplyPayloadGetSerialNumber:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetSerialNumber)))
		case *pckt1ReplyPayloadGetShortAddress:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetShortAddress)))
		case *pckt1ReplyPayloadGetGroupID:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetGroupID)))
		case *pckt1ReplyPayloadGetFixtureInfo:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetFixtureInfo)))
		case *pckt1ReplyPayloadGetTimeReference:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetTimeReference)))
		case *pckt1ReplyPayloadGetLEDsPWM:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetLEDsPWM)))
		case *pckt1ReplyPayloadGetLEDsIrradiance:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetLEDsIrradiance)))
		case *pckt1ReplyPayloadGetSchedulePWM:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetSchedulePWM)))
		case *pckt1ReplyPayloadGetScheduleIrradiance:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetScheduleIrradiance)))
		case *pckt1ReplyPayloadGetScheduleCount:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetScheduleCount)))
		case *pckt1ReplyPayloadGetSchedulingState:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetSchedulingState)))
		case *pckt1ReplyPayloadGetIlluminanceConfiguration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetIlluminanceConfiguration)))
		case *pckt1ReplyPayloadGetModuleTemperature:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadGetModuleTemperature)))
		case *pckt1ReplyPayloadToggleCalibration:
			payload = fmt.Sprintf("%+v", *(packet.Payload.(*pckt1ReplyPayloadToggleCalibration)))
		}
	}
	return fmt.Sprintf("{%s %s}", header, payload)
}
