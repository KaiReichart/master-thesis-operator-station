package gps

import (
	"bytes"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kaireichart/master-thesis-operator-station/events"
)

var (
	currentGPS        *Position
	gpsMutex          = &sync.Mutex{}
	wsClients         = make(map[*websocket.Conn]bool)
	wsClientsMux      = &sync.Mutex{}
	targetIP          = "192.168.178.194"
	targetIPMutex     = &sync.Mutex{}
	isSendingToTarget = false
	sendingMutex      = &sync.Mutex{}

	// Currock Hill coordinates
	currockHillLat = 54.9275
	currockHillLon = -1.8342
	maxDistanceNM  = 9.0
	maxDistanceMux = &sync.Mutex{}
)

func Init() {
	go startUDPListener()
}

func startUDPListener() {
	// Create UDP listener on port 49002
	addr := net.UDPAddr{
		Port: 49002,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Error listening for UDP: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Listening for fs2ff broadcasts on port 49002...")

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading UDP: %v", err)
			continue
		}

		// Need at least a 5-byte header plus data
		if n < 6 {
			continue
		}

		// Check for XGPS header
		if bytes.Equal(buffer[0:4], []byte("XGPS")) {
			// Debug log the raw packet
			log.Printf("Received XGPS packet, length: %d", n)
			log.Printf("Raw packet data: %x", buffer[5:n])
			log.Printf("Raw packet string: %s", string(buffer[5:n]))

			// Parse GPS data
			gpsData, err := parseXGPSPacket(buffer[5:n])
			if err != nil {
				log.Printf("Error parsing GPS data: %v", err)
				continue
			}

			// Convert to our GPSPosition type and update
			position := Position{
				Latitude:  float64(gpsData.Latitude),
				Longitude: float64(gpsData.Longitude),
				Altitude:  float64(gpsData.AltitudeMSL * 0.3048), // Convert feet to meters
				Timestamp: time.Now(),
			}

			// Update current GPS position
			gpsMutex.Lock()
			currentGPS = &position
			gpsMutex.Unlock()

			// Calculate distance to Currock Hill
			distance := calculateDistanceNM(
				position.Latitude,
				position.Longitude,
				currockHillLat,
				currockHillLon,
			)

			// Check if we should send based on distance
			shouldSend := distance <= maxDistanceNM

			// Update sending state if needed
			sendingMutex.Lock()
			if isSendingToTarget != shouldSend {
				isSendingToTarget = shouldSend
				// Create and record the event
				event := events.Event{
					Type:      "sending_toggled",
					Program:   "GPS",
					Timestamp: time.Now(),
				}
				events.LogEvent(event)
			}
			sendingMutex.Unlock()

			// Forward the packet to target IP if enabled and set
			if shouldSend {
				targetIPMutex.Lock()
				if targetIP != "" {
					targetAddr := &net.UDPAddr{
						Port: 49002,
						IP:   net.ParseIP(targetIP),
					}
					targetConn, err := net.DialUDP("udp", nil, targetAddr)
					if err != nil {
						log.Printf("Error creating target connection: %v", err)
					} else {
						_, err := targetConn.Write(buffer[:n])
						if err != nil {
							log.Printf("Error sending UDP packet to target: %v", err)
						}
						targetConn.Close()
					}
				}
				targetIPMutex.Unlock()
			}

			// Broadcast to all WebSocket clients
			wsClientsMux.Lock()
			for client := range wsClients {
				err := client.WriteJSON(position)
				if err != nil {
					log.Printf("Error sending GPS data to client: %v", err)
					client.Close()
					delete(wsClients, client)
				}
			}
			wsClientsMux.Unlock()

			// Log the position update
			log.Printf("Position: Lat=%.6f, Lon=%.6f, Alt=%.1fm, Hdg=%.1f°, GS=%.1fkts, Distance to Currock Hill=%.1fnm",
				position.Latitude,
				position.Longitude,
				position.Altitude,
				gpsData.TrueHeading,
				gpsData.GroundSpeed,
				distance)
		}
	}
}

// GetCurrentPosition returns the current GPS position
func GetCurrentPosition() *Position {
	gpsMutex.Lock()
	defer gpsMutex.Unlock()
	return currentGPS
}

// GetTargetIP returns the current target IP
func GetTargetIP() string {
	targetIPMutex.Lock()
	defer targetIPMutex.Unlock()
	return targetIP
}

// GetDistanceThreshold returns the current distance threshold
func GetDistanceThreshold() float64 {
	maxDistanceMux.Lock()
	defer maxDistanceMux.Unlock()
	return maxDistanceNM
}

// IsSendingToTarget returns whether GPS data is being sent to target
func IsSendingToTarget() bool {
	sendingMutex.Lock()
	defer sendingMutex.Unlock()
	return isSendingToTarget
}
