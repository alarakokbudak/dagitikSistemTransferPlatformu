package discovery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TrackerRegisterReq struct {
	UserID  string `json:"userId"`
	Name    string `json:"name"`
	P2PPort string `json:"p2pPort"`
	GUIPort string `json:"guiPort"`
}

func StartTrackerPolling(trackerURL string, p2pPort string, guiPort string) {
	// Clean ports
	p2pPortOnly := p2pPort
	if strings.Contains(p2pPort, ":") {
		parts := strings.Split(p2pPort, ":")
		p2pPortOnly = parts[len(parts)-1]
	}
	guiPortOnly := guiPort
	if strings.Contains(guiPort, ":") {
		parts := strings.Split(guiPort, ":")
		guiPortOnly = parts[len(parts)-1]
	}

	for {
		userID, userName := getLocalIdentity()
		if userID != "" {
			// 1. Register with tracker
			reqBody := TrackerRegisterReq{
				UserID:  userID,
				Name:    userName,
				P2PPort: p2pPortOnly,
				GUIPort: guiPortOnly,
			}
			jsonData, _ := json.Marshal(reqBody)
			_, err := http.Post(trackerURL+"/register", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Println("[Tracker] Registration failed:", err)
			} else {
				// 2. Fetch peers
				resp, err := http.Get(trackerURL + "/peers")
				if err == nil {
					var trackerPeers []struct {
						UserID   string    `json:"userId"`
						Name     string    `json:"name"`
						Address  string    `json:"address"`
						GUIPort  string    `json:"guiPort"`
						LastSeen time.Time `json:"lastSeen"`
					}
					if err := json.NewDecoder(resp.Body).Decode(&trackerPeers); err == nil {
						peerMutex.Lock()
						for _, tp := range trackerPeers {
							if tp.UserID != userID { // Don't add self
								ActivePeers[tp.UserID] = Peer{
									Name:     tp.Name,
									Address:  tp.Address,
									UserID:   tp.UserID,
									GUIPort:  tp.GUIPort,
									LastSeen: tp.LastSeen,
								}
							}
						}
						// Clean up stale peers
						now := time.Now()
						for k, p := range ActivePeers {
							if now.Sub(p.LastSeen) > 2*time.Minute {
								delete(ActivePeers, k)
							}
						}
						peerMutex.Unlock()
					}
					resp.Body.Close()
				}
			}
		}

		time.Sleep(10 * time.Second) // Poll every 10 seconds
	}
}
