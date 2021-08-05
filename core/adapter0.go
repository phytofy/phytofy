// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for communication with SBC adapters for PHYTOFY RL v0 (DEPRECATED)
package main

import "time"

// Holds the information about an SBC adapter
type dptr0Adapter struct {
	adapterID   string
	scheduleIDs []uint32
	lastSeen    time.Time
}

// Holds the information about a fixture module
type dptr0Module struct {
	serial      schdlSerial
	adapterID   string
	calibration pckt0Calibration
	lastSeen    time.Time
}
