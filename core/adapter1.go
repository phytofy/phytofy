// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for communication with serial port adapters
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type dptr1Identifier string

type dptr1Adapter struct {
	logger    *log.Logger
	address   net.IP
	port      int
	adapterID dptr1Identifier
	handle    *net.TCPConn
	sequence  uint32
	inbox     sync.Map
	outbox    chan []byte
	lut       map[schdlSerial]pckt1ShortAddress
	lutLock   *sync.Mutex
	lastSeen  time.Time
}

const (
	dptr1CommandTimeout   = 10 * time.Second
	dptr1ReconnectTimeout = 2 * dscvr1DiscoveryInterval
)

// Generates an adapter identifier from IP address and port
func dptr1Identify(address net.IP, port int) dptr1Identifier {
	return dptr1Identifier(fmt.Sprintf("%s:%d", address.String(), port))
}

// The code handling the adapter communication
func dptr1Init(logger *log.Logger, address net.IP, port int) *dptr1Adapter {
	return &dptr1Adapter{
		logger,
		address,
		port,
		dptr1Identify(address, port),
		nil,
		rand.Uint32(),
		sync.Map{},
		make(chan []byte, 100),
		make(map[schdlSerial]pckt1ShortAddress),
		&sync.Mutex{},
		time.Now(),
	}
}

func (adapter *dptr1Adapter) dptr1Activate(conditioning bool) {
	go adapter.dptr1Conduit()
	go adapter.dptr1Probe()
	if conditioning {
		go adapter.dptr1Conditioner()
	}
	go adapter.dptr1Connector()
}

func (adapter *dptr1Adapter) dptr1Connector() {
	var octets bytes.Buffer
	for time.Now().Before(adapter.lastSeen.Add(dptr1ReconnectTimeout)) {
		handle := adapter.handle
		if handle != nil {
			adapter.dptr1Close(handle)
		}
		handle, fail := adapter.dptr1Open()
		if fail != nil {
			time.Sleep(time.Second)
			adapter.logger.Printf("INFO: [%s] Retrying adapter connection", adapter.adapterID)
			continue
		}
		adapter.handle = handle
		for time.Now().Before(adapter.lastSeen.Add(dptr1ReconnectTimeout)) {
			if fail := adapter.dptr1Process(handle, &octets); fail != nil {
				adapter.logger.Printf("ERROR: [%s] Failed to process octets coming (%s)", adapter.adapterID, fail)
				adapter.dptr1Close(handle)
				time.Sleep(time.Second)
				adapter.logger.Printf("INFO: [%s] Retrying connection", adapter.adapterID)
				break
			}
		}
	}
}

// Upens the socket used by the thread
func (adapter *dptr1Adapter) dptr1Open() (*net.TCPConn, error) {
	address, fail := net.ResolveTCPAddr("tcp4", string(adapter.adapterID))
	if fail != nil {
		adapter.logger.Printf("ERROR: [%s] Failed to resolve the address (%s)", adapter.adapterID, fail)
		return nil, fail
	}
	handle, fail := net.DialTCP("tcp4", nil, address)
	if fail != nil {
		adapter.logger.Printf("ERROR: [%s] Failed to connect (%s)", adapter.adapterID, fail)
		return nil, fail
	}
	if fail := handle.SetNoDelay(true); fail != nil {
		adapter.logger.Printf("DEBUG: [%s] Failed to set no-delay (%s)", adapter.adapterID, fail)
	}
	if fail := handle.SetKeepAlivePeriod(time.Second); fail != nil {
		adapter.logger.Printf("DEBUG: [%s] Failed to set keep-alive period (%s)", adapter.adapterID, fail)
	}
	if fail := handle.SetKeepAlive(true); fail != nil {
		adapter.logger.Printf("DEBUG: [%s] Failed to set keep-alive (%s)", adapter.adapterID, fail)
	}
	return handle, nil
}

// Transmits a command to the adapter
func (adapter *dptr1Adapter) dptr1Transmit(packet pckt1Packet) error {
	octets, fail := pckt1Encode(packet)
	if fail != nil {
		return fmt.Errorf("Failed to encode a packet (%s)", fail)
	}
	if len(adapter.outbox) != 0 {
		time.Sleep(10 * time.Millisecond)
	}
	adapter.outbox <- octets
	return nil
}

// Issues command and collects reply/replies
func (adapter *dptr1Adapter) dptr1Exchange(command pckt1Packet, timeout time.Duration) ([]pckt1Packet, error) {
	// Open transaction
	inbox := make(chan pckt1Packet, 256)
	transaction := command.Header.SequenceNumber
	adapter.inbox.Store(transaction, inbox)
	// Send the command
	adapter.logger.Printf("INFO: [%s] <- %s", adapter.adapterID, pckt1ToString(command))
	if fail := adapter.dptr1Transmit(command); fail != nil {
		return nil, fail
	}
	// Check if it is necessary to wait for replies
	replies := make([]pckt1Packet, 0)
	if pckt1IsReplying(command.Header.FunctionCode) {
		// Wait for replies
		if command.Header.ShortAddress == pckt1ShortAddressBroadcast {
			// Expecting more replies so wait full timeout
			time.Sleep(timeout)
		} else {
			// Expecting <=1 reply, wait for it no longer than timeout
			select {
			case reply := <-inbox:
				replies = append(replies, reply)
			case <-time.After(timeout):
			}
		}
		// Collect all (remaining) replies
		for len(inbox) > 0 {
			reply := <-inbox
			replies = append(replies, reply)
			// Log replies
			adapter.logger.Printf("INFO: [%s] -> %s", adapter.adapterID, pckt1ToString(reply))
		}
	} else {
		time.Sleep(timeout)
	}
	// Close transaction
	adapter.inbox.Delete(transaction)
	// Return replies
	return replies, nil
}

// Assembles a command and runs the packet exchange
func (adapter *dptr1Adapter) dptr1AssembleAndExchange(shortAddress pckt1ShortAddress, functionCode pckt1FunctionCode, payload pckt1Payload) ([]pckt1Packet, error) {
	switch functionCode {
	case pckt1FunctionCodeSetShortAddress, pckt1FunctionCodeGetShortAddress:
		shortAddress = pckt1ShortAddressBroadcast
	}
	var client [4]byte
	copy(client[:], netMatchOwnAddress(adapter.address, adapter.logger).To4())
	packet := pckt1Packet{
		pckt1Header{
			client,
			atomic.AddUint32(&adapter.sequence, 1),
			shortAddress,
			functionCode,
		},
		payload,
	}
	replies, fail := adapter.dptr1Exchange(packet, dptr1CommandTimeout)
	if fail != nil {
		return nil, fail
	}
	if fail := dptr1CheckReplies(functionCode, replies); fail == nil {
		switch functionCode {
		case pckt1FunctionCodeSetSerialNumber:
			specificPayload := payload.(*pckt1CommandPayloadSetSerialNumber)
			adapter.dptr1ReassociateSingle(specificPayload.Serial, shortAddress)
		case pckt1FunctionCodeSetShortAddress:
			specificPayload := payload.(*pckt1CommandPayloadSetShortAddress)
			adapter.dptr1ReassociateSingle(specificPayload.Serial, specificPayload.ShortAddress)
		}
	} else {
		adapter.logger.Printf("ERROR: [%s] Failure reported in received replies (%s)", adapter.adapterID, fail)
	}
	return replies, nil
}

// Processes the incoming packets from the socket
func (adapter *dptr1Adapter) dptr1Process(handle *net.TCPConn, octets *bytes.Buffer) error {
	if fail := handle.SetReadDeadline(time.Now().Add(time.Second)); fail != nil {
		adapter.logger.Printf("DEBUG: [%s] Failed to set read deadline (%s)", adapter.adapterID, fail)
	}
	octet := []byte{0}
	read, fail := handle.Read(octet)
	if fail != nil && !os.IsTimeout(fail) {
		adapter.logger.Printf("DEBUG: [%s] Failed to read (%s)", adapter.adapterID, fail)
		return fail
	}
	if read != 0 {
		octets.Write(octet)
		adapter.lastSeen = time.Now()
		replies := pckt1Parse(octets, string(adapter.adapterID), adapter.logger)
		for _, reply := range replies {
			queue, present := adapter.inbox.Load(reply.Header.SequenceNumber)
			if present {
				queue.(chan pckt1Packet) <- reply
			} else {
				adapter.logger.Printf("DEBUG: [%s] Dropped orphaned reply - %+v", adapter.adapterID, reply)
			}
		}
	}
	return nil
}

// Closes the socket used by the thread
func (adapter *dptr1Adapter) dptr1Close(handle *net.TCPConn) {
	if fail := handle.Close(); fail != nil {
		adapter.logger.Printf("DEBUG: [%s] Failed to close connection (%s)", adapter.adapterID, fail)
	}
}

// Collects the list of serial numbers seen recently
func (adapter *dptr1Adapter) dptr1ListSeenSerials() schdlSerials {
	serials := make(schdlSerials, 0)
	adapter.lutLock.Lock()
	for serial := range adapter.lut {
		serials = append(serials, serial)
	}
	adapter.lutLock.Unlock()
	return serials
}

// Checks if given serials are among the list of serial numbers seen recently
func (adapter *dptr1Adapter) dptr1CheckSeenSerials(serials schdlSerials) bool {
	seenSet := make(map[schdlSerial]struct{})
	seenList := adapter.dptr1ListSeenSerials()
	for _, serial := range seenList {
		seenSet[serial] = struct{}{}
	}
	for _, serial := range serials {
		if _, exists := seenSet[serial]; !exists {
			return false
		}
	}
	return true
}

// Looks up the address assigned to the given serial number
func (adapter *dptr1Adapter) dptr1LookUp(serial schdlSerial) pckt1ShortAddress {
	if serial == 0 {
		return pckt1ShortAddressBroadcast
	}
	adapter.lutLock.Lock()
	shortAddress, present := adapter.lut[serial]
	adapter.lutLock.Unlock()
	if !present {
		return pckt1ShortAddressUnassigned
	}
	return shortAddress
}

// Reassociates address with serial
func (adapter *dptr1Adapter) dptr1ReassociateSingle(serial schdlSerial, shortAddress pckt1ShortAddress) {
	adapter.lutLock.Lock()
	for other := range adapter.lut {
		if adapter.lut[other] == shortAddress {
			delete(adapter.lut, other)
		}
	}
	adapter.lut[serial] = shortAddress
	adapter.lutLock.Unlock()
}

// Reassociates addresses with serials
func (adapter *dptr1Adapter) dptr1ReassociateAll(lut map[schdlSerial]pckt1ShortAddress) {
	adapter.lutLock.Lock()
	adapter.lut = lut
	adapter.lutLock.Unlock()
}

// Checks if all replies succeeded
func dptr1CheckReplies(functionCode pckt1FunctionCode, replies []pckt1Packet) error {
	switch functionCode {
	case pckt1FunctionCodeSetModuleCalibration, pckt1FunctionCodeSetSerialNumber, pckt1FunctionCodeSetShortAddress, pckt1FunctionCodeSetGroupID, pckt1FunctionCodeSetFixtureInfo, pckt1FunctionCodeSetTimeReference, pckt1FunctionCodeSetSchedule, pckt1FunctionCodeDeleteSchedule, pckt1FunctionCodeDeleteAllSchedules, pckt1FunctionCodeStopScheduling, pckt1FunctionCodeResumeScheduling, pckt1FunctionCodeSetIlluminanceConfiguration, pckt1FunctionCodeResetForFirmwareUpdate:
		result := ""
		for _, reply := range replies {
			switch reply.Payload.(type) {
			case *pckt1ReplyPayloadGenericNOK:
				if len(result) != 0 {
					result += "; "
				}
				result += fmt.Sprintf("Short address %d replied with error code %d", reply.Header.ShortAddress, reply.Payload.(*pckt1ReplyPayloadGenericNOK).ErrorCode)
			}
		}
		if len(result) != 0 {
			return fmt.Errorf("NACK - %s", result)
		}
	case pckt1FunctionCodeToggleCalibration:
		result := ""
		for _, reply := range replies {
			if !reply.Payload.(*pckt1ReplyPayloadToggleCalibration).Ack {
				if len(result) != 0 {
					result += "; "
				}
				result += fmt.Sprintf("Short address %d", reply.Header.ShortAddress)
			}
		}
		if len(result) != 0 {
			return fmt.Errorf("NACK - %s", result)
		}
	case pckt1FunctionCodeSetLEDs, pckt1FunctionCodeConfirmResetForFirmwareUpdate:
		if len(replies) != 0 {
			return fmt.Errorf("Invalid reply function code - %d", functionCode)
		}
	default:
		if len(replies) == 0 {
			return fmt.Errorf("No replies")
		}
	}
	return nil
}

// Used by the adapter object to send periodically a search request
func (adapter *dptr1Adapter) dptr1Probe() {
	for time.Now().Before(adapter.lastSeen.Add(dptr1ReconnectTimeout)) {
		replies, fail := adapter.dptr1AssembleAndExchange(
			pckt1ShortAddressBroadcast, pckt1FunctionCodeGetSerialNumber, &pckt1CommandPayloadGetSerialNumber{true})
		if fail != nil {
			adapter.logger.Printf("ERROR: [%s] Could not communicate (%s)", adapter.adapterID, fail)
			time.Sleep(time.Second)
			continue
		}
		lut := dptr1ProbeCollectEach(replies)
		unassigned := dptr1ProbeCollectUnassigned(replies)
		duplicated := dptr1ProbeCollectDuplicated(replies)
		unused := dptr1ProbeCollectUnused(replies)
		unassigned = append(unassigned, duplicated...)
		for _, serial := range unassigned {
			if len(unused) == 0 {
				adapter.logger.Printf("ERROR: [%s] Could not find available address", adapter.adapterID)
				break
			}
			available := unused[0]
			replies, fail := adapter.dptr1AssembleAndExchange(
				pckt1ShortAddressBroadcast, pckt1FunctionCodeSetShortAddress, &pckt1CommandPayloadSetShortAddress{serial, available})
			if fail != nil {
				adapter.logger.Printf("ERROR: [%s] Could not communicate (%s)", adapter.adapterID, fail)
			} else if fail := dptr1CheckReplies(pckt1FunctionCodeSetShortAddress, replies); fail != nil {
				adapter.logger.Printf("ERROR: [%s] Could not assign available address to %d (%s)", adapter.adapterID, serial, fail)
			} else {
				lut[serial] = available
				unused = unused[1:]
			}
		}
		adapter.dptr1ReassociateAll(lut)
		time.Sleep(8 * time.Second)
	}
}

// Collects addresses assigned to each serial number
func dptr1ProbeCollectEach(replies []pckt1Packet) map[schdlSerial]pckt1ShortAddress {
	each := make(map[schdlSerial]pckt1ShortAddress)
	for _, reply := range replies {
		shortAddress := reply.Header.ShortAddress
		serial := reply.Payload.(*pckt1ReplyPayloadGetSerialNumber).Serial
		each[serial] = shortAddress
	}
	return each
}

// Collects all serial numbers without assigned address
func dptr1ProbeCollectUnassigned(replies []pckt1Packet) []schdlSerial {
	unassigned := make([]schdlSerial, 0)
	for _, reply := range replies {
		shortAddress := reply.Header.ShortAddress
		if shortAddress == pckt1ShortAddressUnassigned {
			unassigned = append(unassigned, reply.Payload.(*pckt1ReplyPayloadGetSerialNumber).Serial)
		}
	}
	return unassigned
}

// Collects all serial numbers with misassigned addresses
func dptr1ProbeCollectDuplicated(replies []pckt1Packet) []schdlSerial {
	duplicated := make([]schdlSerial, 0)
	seen := make(map[pckt1ShortAddress]schdlSerial)
	for _, reply := range replies {
		shortAddress := reply.Header.ShortAddress
		serial := reply.Payload.(*pckt1ReplyPayloadGetSerialNumber).Serial
		if shortAddress != pckt1ShortAddressUnassigned {
			if sort.Search(len(seen), func(i int) bool { _, present := seen[shortAddress]; return present }) == -1 {
				seen[shortAddress] = serial
			} else {
				if other, present := seen[shortAddress]; present && other != serial {
					duplicated = append(duplicated, other)
					duplicated = append(duplicated, serial)
				}
			}
		}
	}
	return duplicated
}

// Collects all unassigned addresses
func dptr1ProbeCollectUnused(replies []pckt1Packet) []pckt1ShortAddress {
	used := make(map[pckt1ShortAddress]struct{})
	for _, reply := range replies {
		used[reply.Header.ShortAddress] = struct{}{}
	}
	unused := make([]pckt1ShortAddress, 0)
	for shortAddress := pckt1ShortAddress(pckt1ShortAddressBegin); shortAddress <= pckt1ShortAddressEnd; shortAddress++ {
		_, present := used[shortAddress]
		if !present {
			unused = append(unused, shortAddress)
		}
	}
	return unused
}

// Decouples the threads sending packets from the socket
func (adapter *dptr1Adapter) dptr1Conduit() {
	for time.Now().Before(adapter.lastSeen.Add(dptr1ReconnectTimeout)) {
		handle := adapter.handle
		if handle == nil {
			time.Sleep(2 * time.Second)
			continue
		}
		select {
		case octets := <-adapter.outbox:
			if fail := handle.SetWriteDeadline(time.Now().Add(time.Second)); fail != nil {
				adapter.logger.Printf("DEBUG: [%s] Failed to set write deadline (%s)", adapter.adapterID, fail)
			}
			written, fail := handle.Write(octets)
			if fail != nil {
				adapter.logger.Printf("ERROR: [%s] Failed to transmit a packet (%s), dropping - %s", adapter.adapterID, fail, hex.EncodeToString(octets))
				time.Sleep(time.Second)
			} else if written < len(octets) {
				adapter.logger.Printf("ERROR: [%s] Failed to fully transmit a packet (%s), dropping - %s", adapter.adapterID, fail, hex.EncodeToString(octets))
				time.Sleep(time.Second)
			} else {
				adapter.logger.Printf("INFO: [%s] <- %s", adapter.adapterID, hex.EncodeToString(octets))
				time.Sleep(10 * time.Millisecond)
			}
		case <-time.After(time.Second):
		}
	}
}

// Conditions fixtures
func (adapter *dptr1Adapter) dptr1Conditioner() {
	for time.Now().Before(adapter.lastSeen.Add(dptr1ReconnectTimeout)) {
		for _, serial := range adapter.dptr1ListSeenSerials() {
			adapter.logger.Printf("INFO: [%s] conditioning %d", adapter.adapterID, serial)
			shortAddress := adapter.dptr1LookUp(serial)
			if shortAddress == pckt1ShortAddressUnassigned {
				continue
			}
			adapter.dptr1ConditionerSync(shortAddress)
			adapter.dptr1ConditionerToggle(shortAddress)
			adapter.dptr1ConditionerScale(shortAddress)
		}
		adapter.logger.Printf("INFO: [%s] conditioned", adapter.adapterID)
		time.Sleep(10 * time.Minute)
	}
}

// Syncs time
func (adapter *dptr1Adapter) dptr1ConditionerSync(shortAddress pckt1ShortAddress) {
	now := uint32(time.Now().Unix())
	payload := &pckt1CommandPayloadSetTimeReference{now}
	replies, fail := adapter.dptr1AssembleAndExchange(shortAddress, pckt1FunctionCodeSetTimeReference, payload)
	if fail := dptr1CheckResult(pckt1FunctionCodeSetTimeReference, replies, fail); fail != nil {
		adapter.logger.Printf("ERROR: [%s] Failure to sync time from %d (%s)", adapter.adapterID, shortAddress, fail)
	}
}

// Toggles scheduling state
func (adapter *dptr1Adapter) dptr1ConditionerToggle(shortAddress pckt1ShortAddress) {
	replies, fail := adapter.dptr1AssembleAndExchange(shortAddress, pckt1FunctionCodeGetScheduleCount, nil)
	if fail := dptr1CheckResult(pckt1FunctionCodeGetScheduleCount, replies, fail); fail != nil {
		adapter.logger.Printf("ERROR: [%s] Failure when counting schedules from %d (%s)", adapter.adapterID, shortAddress, fail)
	} else {
		count := replies[0].Payload.(*pckt1ReplyPayloadGetScheduleCount).ScheduleCount
		var functionCode pckt1FunctionCode
		if count != 0 {
			functionCode = pckt1FunctionCodeResumeScheduling
		} else {
			functionCode = pckt1FunctionCodeStopScheduling
		}
		if fail := dptr1CheckResult(functionCode, replies, fail); fail != nil {
			adapter.logger.Printf("ERROR: [%s] Failure to toggle scheduling from %d (%s)", adapter.adapterID, shortAddress, fail)
			return
		}
	}
}

// Scales illuminance
func (adapter *dptr1Adapter) dptr1ConditionerScale(shortAddress pckt1ShortAddress) {
	replies0, fail0 := adapter.dptr1AssembleAndExchange(shortAddress, pckt1FunctionCodeGetModuleCalibration, &pckt1CommandPayloadGetModuleCalibration{0})
	if fail0 := dptr1CheckResult(pckt1FunctionCodeGetModuleCalibration, replies0, fail0); fail0 != nil {
		adapter.logger.Printf("ERROR: [%s] Failure when fetching module 0 calibration from %d (%s)", adapter.adapterID, shortAddress, fail0)
		return
	}
	replies1, fail1 := adapter.dptr1AssembleAndExchange(shortAddress, pckt1FunctionCodeGetModuleCalibration, &pckt1CommandPayloadGetModuleCalibration{1})
	if fail1 := dptr1CheckResult(pckt1FunctionCodeGetModuleCalibration, replies1, fail1); fail1 != nil {
		adapter.logger.Printf("ERROR: [%s] Failure when fetching module 1 calibration from %d (%s)", adapter.adapterID, shortAddress, fail1)
		return
	}
	calibration0 := replies0[0].Payload.(*pckt1ReplyPayloadGetModuleCalibration).Calibration
	calibration1 := replies1[0].Payload.(*pckt1ReplyPayloadGetModuleCalibration).Calibration
	configuration := dptr1IlluminanceConfiguration(calibration0, calibration1)
	replies, fail := adapter.dptr1AssembleAndExchange(
		shortAddress, pckt1FunctionCodeSetIlluminanceConfiguration, &pckt1CommandPayloadSetIlluminanceConfiguration{configuration})
	if fail := dptr1CheckResult(pckt1FunctionCodeSetIlluminanceConfiguration, replies, fail); fail != nil {
		adapter.logger.Printf("ERROR: [%s] Failure when setting illuminance configuration from %d (%s)", adapter.adapterID, shortAddress, fail)
	}
}

// Calculates illuminance configuration coefficients
func dptr1IlluminanceConfiguration(calibration0 pckt1Calibration, calibration1 pckt1Calibration) [6]float32 {
	configuration := [6]float32{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}
	for index := range configuration {
		minM, _ := dptr1CalibrationMinAvg(calibration0, calibration1, index)
		configuration[index] = float32(dptr1SingleIlluminanceConfiguration(minM))
	}
	return configuration
}

// Returns min and avg of calibration values
func dptr1CalibrationMinAvg(calibration0 pckt1Calibration, calibration1 pckt1Calibration, index int) (float64, float64) {
	m0 := float64(calibration0[index].CoefficientM)
	m1 := float64(calibration1[index].CoefficientM)
	minM := math.Min(m0, m1)
	avgM := (m0 + m1) / 2.0
	return minM, avgM
}

// Calculates single illuminance configuration coefficient
func dptr1SingleIlluminanceConfiguration(minimum float64) float64 {
	return 100.0 / minimum
}

// Checks the replies and cosolidates into single error in case of any failure
func dptr1CheckResult(functionCode pckt1FunctionCode, replies []pckt1Packet, fail error) error {
	if fail == nil {
		fail = dptr1CheckReplies(functionCode, replies)
	}
	if fail != nil {
		return fmt.Errorf("Failed result (%s)", fail)
	}
	return nil
}
