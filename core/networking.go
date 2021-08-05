// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This code is responsible for communicating over UDP
package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	backoff = 50 * time.Millisecond
)

// Holds an observation captured over UDP
type networkingObservation struct {
	address net.IP
	port    int
	buffer  *[]byte
}

// Holds the state of communication over UDP
type networking struct {
	sourcePort       int
	targetPort       int
	running          bool
	connection       *net.UDPConn
	lastTransmission time.Time
	listeners        sync.Map
	logger           *log.Logger
}

// Creates an instance communicating over UDP
func netInit(targetPort int, logger *log.Logger) *networking {
	n := &networking{0, targetPort, true, nil, time.Now(), sync.Map{}, logger}
	go n.netRun()
	return n
}

// Main routine handling communication over UDP
func (n *networking) netRun() {
	buffer := make([]byte, 1024*1024)
	for n.running {
		n.netOpen()
		for n.running {
			if fail := n.netReceive(buffer); fail != nil {
				n.logger.Printf("ERROR: UDP reception failed (%s), will attempt to reset", fail)
				n.netClose()
				time.Sleep(time.Second)
				break
			}
		}
	}
}

// Opens and binds the socket used by the routine
func (n *networking) netOpen() {
	for n.sourcePort = 10000; n.sourcePort <= 50000; n.sourcePort++ {
		attempt := fmt.Sprintf(":%d", n.sourcePort)
		address, fail := net.ResolveUDPAddr("udp4", attempt)
		if fail != nil {
			n.logger.Printf("DEBUG: UDP failed to resolve %s (%s), moving on", attempt, fail)
			continue
		}
		if connection, fail := net.ListenUDP("udp4", address); fail == nil {
			n.logger.Printf("INFO: UDP listening on %v", address)
			n.connection = connection
			return
		}
		n.logger.Printf("INFO: UDP failed to listen on %v (%s), moving on", address, fail)
	}
	n.logger.Panicf("CRITICAL: UDP cannot find an available port")
}

// Processes the incoming packets from the socket
func (n *networking) netReceive(buffer []byte) error {
	connection := n.connection
	if connection == nil {
		n.logger.Printf("INFO: UDP link is absent, blocking reception for a second")
		time.Sleep(time.Second)
		return nil
	}
	deadline := time.Now().Add(time.Second)
	if fail := connection.SetReadDeadline(deadline); fail != nil {
		n.logger.Printf("DEBUG: UDP did not allow to set a read deadline (%s)", fail)
		return fail
	}
	read, address, fail := connection.ReadFromUDP(buffer)
	if netError, ok := fail.(net.Error); ok && netError.Timeout() {
	} else if fail != nil {
		n.logger.Printf("ERROR: UDP read failed (%s)", fail)
		return fail
	} else if address.Port == n.targetPort {
		n.netNotify(buffer, read, address)
	}
	return nil
}

// Notify of received observation
func (n *networking) netNotify(buffer []byte, read int, address *net.UDPAddr) {
	received := make([]byte, read)
	copy(received, buffer[:read])
	n.logger.Printf("INFO: UDP read %d bytes from %v - %s", read, address, hex.EncodeToString(received))
	n.listeners.Range(func(key, value interface{}) bool {
		if key == nil {
			n.logger.Printf("ERROR: UDP found nil reception channel, ignoring")
			return false
		}
		channel := key.(*chan networkingObservation)
		*channel <- networkingObservation{address.IP, address.Port, &received}
		return true
	})
}

// Sends request out
func (n *networking) netTransmit(destination net.IP, port int, request *[]byte) bool {
	if !n.running {
		n.logger.Printf("ERROR: UDP is not running, dropping outgoing request - %s", hex.EncodeToString(*request))
		return false
	}
	connection := n.connection
	if connection == nil {
		n.logger.Printf("ERROR: UDP link is absent, dropping outgoing request - %s", hex.EncodeToString(*request))
		return false
	}
	n.netBackoff()
	attempt := fmt.Sprintf("%s:%d", destination, port)
	address, fail := net.ResolveUDPAddr("udp4", attempt)
	if fail != nil {
		n.logger.Printf("ERROR: UDP failed to resolve %s (%s), dropping outgoing request - %s", attempt, fail, hex.EncodeToString(*request))
		return false
	}
	written, fail := connection.WriteToUDP(*request, address)
	if fail != nil {
		n.logger.Printf("ERROR: UDP failed to write to %v (%s), abandoning attempt - %s", address, fail, hex.EncodeToString(*request))
		return false
	}
	if written != len(*request) {
		n.logger.Printf("ERROR: UDP wrote only %d bytes to %v, abandoning attempt - %s", written, address, hex.EncodeToString(*request))
		return false
	}
	n.logger.Printf("INFO: UDP wrote %d bytes to %v - %s", written, address, hex.EncodeToString(*request))
	n.lastTransmission = time.Now()
	return true
}

// Backs off (if necessary) before sending a request
func (n *networking) netBackoff() {
	now := time.Now()
	elapsed := now.Sub(n.lastTransmission)
	if elapsed < backoff {
		n.logger.Printf("INFO: UDP backing off")
		time.Sleep(backoff - elapsed)
	}
}

// Closes the socket used by the routine
func (n *networking) netClose() {
	n.logger.Printf("INFO: UDP closing link")
	connection := n.connection
	n.connection = nil
	if connection != nil {
		if fail := connection.Close(); fail != nil {
			n.logger.Printf("DEBUG: UDP failed to close link (%s), ignoring", fail)
		}
	}
}

// Stops this routine
func (n *networking) netTerminate() {
	n.logger.Printf("INFO: UDP terminating link")
	n.running = false
	n.netClose()
}

// Acquire an observation channel
func (n *networking) netAcquireChannel() *chan networkingObservation {
	channel := make(chan networkingObservation, 100)
	n.listeners.Store(&channel, struct{}{})
	return &channel
}

// Release an observation channel
func (n *networking) netReleaseChannel(channel *chan networkingObservation) {
	n.listeners.Delete(channel)
}

// Lists broadcating addresses for all local networks
func netBroadcasts(logger *log.Logger) []net.IP {
	all := make([]net.IP, 0)
	netIterateInterfaces(func(network *net.IPNet) bool {
		ip := network.IP.To4()
		mask := network.Mask
		for i := range ip {
			ip[i] |= ^mask[i]
		}
		all = append(all, ip)
		return false
	}, logger)
	if len(all) == 0 {
		all = append(all, net.IPv4bcast)
	}
	logger.Printf("INFO: UDP broadcast IP addresses - %v", all)
	return all
}

// Looks up the IP address from among own interfaces in the same subnet
func netMatchOwnAddress(other net.IP, logger *log.Logger) net.IP {
	matched := net.IPv4bcast
	netIterateInterfaces(func(network *net.IPNet) bool {
		own := network.IP.To4()
		ownNetwork := own.Mask(network.Mask)
		otherNetwork := other.Mask(network.Mask)
		if ownNetwork.Equal(otherNetwork) {
			matched = own
			return true
		}
		return false
	}, logger)
	logger.Printf("INFO: UDP matched %v own IP address to %v", matched, other)
	return matched
}

func netIterateInterfaces(apply func(network *net.IPNet) bool, logger *log.Logger) {
	interfaces, fail := net.Interfaces()
	if fail != nil {
		logger.Panicf("CRITICAL: Could not iterate network interfaces (%s)", fail)
	}
	for _, entry := range interfaces {
		if entry.Flags&net.FlagUp == 0 || entry.Flags&net.FlagBroadcast == 0 || entry.Flags&net.FlagLoopback != 0 {
			logger.Printf("INFO: UDP interface %v not applicable (must be: up, broadcast, non-loopback), skipping", entry)
			continue
		}
		logger.Printf("INFO: UDP interface %v applicable (must be: up, broadcast, non-loopback)", entry)
		if addresses, fail := entry.Addrs(); fail == nil {
			for _, address := range addresses {
				logger.Printf("INFO: UDP iterating over %v", address)
				switch network := address.(type) {
				case *net.IPNet:
					if network.IP.DefaultMask() != nil {
						logger.Printf("INFO: UDP using %v", network)
						if apply(network) {
							return
						}
					}
				}
			}
		} else {
			logger.Printf("DEBUG: UDP could not query a network interface %v (%s), skipping", entry, fail)
		}
	}
}
