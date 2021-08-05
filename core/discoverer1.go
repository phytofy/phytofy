// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for discovery of MOXA NPort adapters
package main

import (
	"encoding/hex"
	"log"
	"reflect"
	"sync"
	"time"
)

const (
	dscvr1MoxaDiscoveryPort     = 4800
	dscvr1MoxaCommunicationPort = 4001
	dscvr1MoxaModelVariantMask  = 0x0F
	dscvr1FieldCode             = 0
	dscvr1FieldLength           = 3
	dscvr1FieldModel            = 12
	dscvr1FieldModelVariant     = 13
	dscvr1FieldMAC              = 14
	dscvr1FieldIP               = 20
	dscvr1CodeDiscoveryReply    = 0x81
	dscvr1LengthDiscoveryReply  = 24
	dscvr1DiscoveryInterval     = 10 * time.Second
)

var (
	dscvr1MoxaOIU          = []byte{0x00, 0x90, 0xE8}
	dscvr1DiscoveryRequest = []byte{0x01, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00}
)

type dscvr1Discoverer struct {
	logger       *log.Logger
	networking   *networking
	observer     *chan networkingObservation
	adapters     sync.Map
	conditioning bool
}

// The main thread handling the adapter discovery
func dscvr1Init(logger *log.Logger, conditioning bool) *dscvr1Discoverer {
	networking := netInit(dscvr1MoxaDiscoveryPort, logger)
	observer := networking.netAcquireChannel()
	discoverer := &dscvr1Discoverer{logger, networking, observer, sync.Map{}, conditioning}
	go discoverer.dscvr1Process()
	go discoverer.dscvr1ProbeRoutine()
	go discoverer.dscvr1ForgettingRoutine()
	return discoverer
}

// Processes the incoming packets from the network
func (discoverer *dscvr1Discoverer) dscvr1Process() {
	for discoverer.networking.running {
		observation := <-*discoverer.observer
		discoverer.logger.Printf("INFO: [%s:%d] -> %s", observation.address.String(), observation.port, hex.EncodeToString(*observation.buffer))
		if observation.port == dscvr1MoxaDiscoveryPort {
			if !discoverer.dscvr1Check(observation.buffer) {
				continue
			}
			for i := 0; i < dscvr1Count(observation.buffer); i++ {
				port := dscvr1MoxaCommunicationPort + i
				identifier := dptr1Identify(observation.address, port)
				adapter := dptr1Init(discoverer.logger, observation.address, port)
				if _, loaded := discoverer.adapters.LoadOrStore(identifier, adapter); !loaded {
					adapter.dptr1Activate(discoverer.conditioning)
				}
			}
		}
	}
}

// Checks the incoming packets from the socket
func (discoverer *dscvr1Discoverer) dscvr1Check(buffer *[]byte) bool {
	length := len(*buffer)
	if length != dscvr1LengthDiscoveryReply {
		discoverer.logger.Printf("ERROR: Invalid length of a discovery reply")
		return false
	}
	if (*buffer)[dscvr1FieldLength] != dscvr1LengthDiscoveryReply {
		discoverer.logger.Printf("ERROR: Invalid length field in a discovery reply")
		return false
	}
	oiuBegin := dscvr1FieldMAC
	oiuEnd := dscvr1FieldMAC + len(dscvr1MoxaOIU)
	reportedOIU := (*buffer)[oiuBegin:oiuEnd]
	if !reflect.DeepEqual(reportedOIU, dscvr1MoxaOIU) {
		discoverer.logger.Printf("ERROR: Invalid OIU in a commissioning reply")
		return false
	}
	return true
}

// Counts the number of serial port adapters
func dscvr1Count(buffer *[]byte) int {
	variant := (*buffer)[dscvr1FieldModelVariant] & dscvr1MoxaModelVariantMask
	switch variant {
	case 1:
		return 1
	case 2:
		return 2
	case 4:
		return 4
	case 7:
		return 8
	case 8:
		return 16
	default:
		return 0
	}
}

// Waits for any fixtures to be present
func (discoverer *dscvr1Discoverer) dscvr1WaitForAnySerials(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		found := false
		discoverer.adapters.Range(func(key, value interface{}) bool {
			found = len(value.(*dptr1Adapter).dptr1ListSeenSerials()) != 0
			return !found
		})
		if found {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

// Waits for fixtures with given serial numbers to be present
func (discoverer *dscvr1Discoverer) dscvr1WaitForSerials(serials schdlSerials, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		seenSetAll := make(map[schdlSerial]struct{})
		discoverer.adapters.Range(func(key, value interface{}) bool {
			seenList := value.(*dptr1Adapter).dptr1ListSeenSerials()
			for _, serial := range seenList {
				seenSetAll[serial] = struct{}{}
			}
			return true
		})
		seenAll := true
		for _, serial := range serials {
			if _, exists := seenSetAll[serial]; !exists {
				seenAll = false
				break
			}
		}
		if seenAll {
			return true
		}
		time.Sleep(time.Second)
	}
	return false
}

// Waits for fixtures to be present
func (discoverer *dscvr1Discoverer) dscvr1WaitForSerial(serial schdlSerial, timeout time.Duration) bool {
	if serial == 0 {
		return discoverer.dscvr1WaitForAnySerials(timeout)
	}
	return discoverer.dscvr1WaitForSerials(schdlSerials{serial}, timeout)
}

// Looks up the adapter where fixture with given serial is attached to
func (discoverer *dscvr1Discoverer) dscvr1LookUp(serial schdlSerial) []*dptr1Adapter {
	singleSerial := schdlSerials{serial}
	adapters := make([]*dptr1Adapter, 0)
	discoverer.adapters.Range(func(key, value interface{}) bool {
		if serial == 0 || value.(*dptr1Adapter).dptr1CheckSeenSerials(singleSerial) {
			adapters = append(adapters, value.(*dptr1Adapter))
		}
		return serial == 0 || !value.(*dptr1Adapter).dptr1CheckSeenSerials(singleSerial)
	})
	return adapters
}

// Used by the discoverer to send periodically a discovery request
func (discoverer *dscvr1Discoverer) dscvr1ProbeRoutine() {
	hexedDiscoveryRequest := hex.EncodeToString(dscvr1DiscoveryRequest)
	tick := time.Tick(dscvr1DiscoveryInterval)
	for discoverer.networking.running {
		select {
		case <-tick:
			broadcasts := netBroadcasts(discoverer.logger)
			for _, broadcast := range broadcasts {
				if !discoverer.networking.netTransmit(broadcast, dscvr1MoxaDiscoveryPort, &dscvr1DiscoveryRequest) {
					discoverer.logger.Printf("ERROR: Failed to broadcast the discovery request to %v", broadcast)
				} else {
					discoverer.logger.Printf("INFO: [%v:%d] <- %s", broadcast, dscvr1MoxaDiscoveryPort, hexedDiscoveryRequest)
				}
			}
		}
	}
}

// Routine used by the discoverer to forget old adapters
func (discoverer *dscvr1Discoverer) dscvr1ForgettingRoutine() {
	tick := time.Tick(dscvr1DiscoveryInterval)
	for discoverer.networking.running {
		select {
		case <-tick:
			now := time.Now()
			// Removes adapters which did not make contact for too long
			discoverer.adapters.Range(func(key, value interface{}) bool {
				if now.Sub(value.(*dptr1Adapter).lastSeen) > 5*dscvr1DiscoveryInterval {
					discoverer.adapters.Delete(key)
				}
				return true
			})
		}
	}
}
