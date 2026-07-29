package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type PeerInfo struct {
	UserID   string    `json:"userId"`
	Name     string    `json:"name"`
	Address  string    `json:"address"` // PublicIP:P2PPort
	GUIPort  string    `json:"guiPort"`
	LastSeen time.Time `json:"lastSeen"`
}

var (
	peers      = make(map[string]PeerInfo) // map[UserID]PeerInfo
	peersMutex sync.RWMutex
)

// Extract pure IP from RemoteAddr
func getIP(remoteAddr string) string {
	if strings.Contains(remoteAddr, ":") {
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil {
			return host
		}
	}
	return remoteAddr
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID  string `json:"userId"`
		Name    string `json:"name"`
		P2PPort string `json:"p2pPort"`
		GUIPort string `json:"guiPort"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.P2PPort == "" {
		http.Error(w, "userId and p2pPort are required", http.StatusBadRequest)
		return
	}

	clientIP := getIP(r.RemoteAddr)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	peer := PeerInfo{
		UserID:   req.UserID,
		Name:     req.Name,
		Address:  fmt.Sprintf("%s:%s", clientIP, req.P2PPort),
		GUIPort:  req.GUIPort,
		LastSeen: time.Now(),
	}

	peersMutex.Lock()
	peers[req.UserID] = peer
	peersMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	log.Printf("Registered peer %s at %s", req.UserID, peer.Address)
}

func handlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	peersMutex.RLock()
	var activePeers []PeerInfo
	now := time.Now()
	for _, p := range peers {
		if now.Sub(p.LastSeen) < 2*time.Minute {
			activePeers = append(activePeers, p)
		}
	}
	peersMutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activePeers)
}

func cleanupLoop() {
	for {
		time.Sleep(1 * time.Minute)
		peersMutex.Lock()
		now := time.Now()
		for id, p := range peers {
			if now.Sub(p.LastSeen) > 5*time.Minute {
				delete(peers, id)
				log.Printf("Removed inactive peer %s", id)
			}
		}
		peersMutex.Unlock()
	}
}

func main() {
	port := "8000" // Default port
	
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/peers", handlePeers)

	go cleanupLoop()

	log.Printf("Tracker Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
