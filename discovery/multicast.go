package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	MulticastAddr = "224.0.0.250:9999"
	MagicString   = "MERKLESYNC_PEER:" // Format: MERKLESYNC_PEER:<port>:<name>:<userID>
)

type Peer struct {
	Name     string    `json:"name"`
	Address  string    `json:"address"` // IP:Port (P2P)
	UserID   string    `json:"userId"`
	GUIPort  string    `json:"guiPort"`
	LastSeen time.Time `json:"lastSeen"`
}

var (
	ActivePeers   = make(map[string]Peer)
	peerMutex     sync.RWMutex
	localUserID   string
	localUserName string
	identityMutex sync.RWMutex
)

// SetLocalIdentity hesap oluşturulduğunda/giriş yapıldığında çağrılır
func SetLocalIdentity(userID string, displayName string) {
	identityMutex.Lock()
	defer identityMutex.Unlock()
	localUserID = userID
	localUserName = displayName
}

func getLocalIdentity() (string, string) {
	identityMutex.RLock()
	defer identityMutex.RUnlock()
	return localUserID, localUserName
}

// GetActivePeers returns a list of currently discovered peers
func GetActivePeers() []Peer {
	peerMutex.RLock()
	defer peerMutex.RUnlock()

	var peers []Peer
	for _, p := range ActivePeers {
		peers = append(peers, p)
	}
	return peers
}

// StartBroadcasting runs in the background on the Server.
func StartBroadcasting(tcpPort string, peerName string, guiPort string) {
	addr, err := net.ResolveUDPAddr("udp", MulticastAddr)
	if err != nil {
		fmt.Println("Multicast resolve error:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Multicast dial error:", err)
		return
	}
	defer conn.Close()

	portOnly := tcpPort
	if strings.Contains(tcpPort, ":") {
		parts := strings.Split(tcpPort, ":")
		portOnly = parts[len(parts)-1]
	}

	fmt.Printf("Started mDNS Broadcaster on port %s\n", portOnly)
	for {
		userID, userName := getLocalIdentity()
		name := peerName
		if userName != "" {
			name = userName
		}

		// Payload: MERKLESYNC_PEER:<p2pPort>:<name>:<userID>:<guiPort>
		payload := fmt.Sprintf("%s%s:%s:%s:%s", MagicString, portOnly, name, userID, guiPort)
		message := []byte(payload)

		_, err := conn.Write(message)
		if err != nil {
			fmt.Println("Broadcast error:", err)
		}
		time.Sleep(2 * time.Second)
	}
}

// StartListening runs continuously in the background to discover peers.
func StartListening(ignorePort string) {
	// TCP LAN Sweep (Bypass Windows UDP Multicast Firewall)
	go func() {
		for {
			for port := 9090; port <= 9095; port++ {
				url := fmt.Sprintf("http://127.0.0.1:%d/api/status", port)
				client := &http.Client{Timeout: 1 * time.Second}
				resp, err := client.Get(url)
				if err == nil && resp.StatusCode == 200 {
					var result map[string]interface{}
					if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
						p2pPort, _ := result["p2pPort"].(string)
						peerName, _ := result["name"].(string)
						peerUserID, _ := result["userId"].(string)
						peerGUIPort := fmt.Sprintf("%d", port)

						if p2pPort != "" && peerUserID != "" {
							portOnly := p2pPort
							if strings.Contains(p2pPort, ":") {
								parts := strings.Split(p2pPort, ":")
								portOnly = parts[len(parts)-1]
							}

							// Kendi kendimizi eklemeyelim
							myID, _ := getLocalIdentity()
							if peerUserID != myID && peerUserID != "" {
								if ":"+portOnly != ignorePort && portOnly != ignorePort {
									serverAddress := fmt.Sprintf("127.0.0.1:%s", portOnly)
									peerMutex.Lock()
									ActivePeers[serverAddress] = Peer{
										Name:     peerName,
										Address:  serverAddress,
										UserID:   peerUserID,
										GUIPort:  peerGUIPort,
										LastSeen: time.Now(),
									}
									peerMutex.Unlock()
								}
							}
						}
					}
					resp.Body.Close()
				}
			}
			time.Sleep(3 * time.Second)
		}
	}()

	addr, err := net.ResolveUDPAddr("udp", MulticastAddr)
	if err != nil {
		fmt.Println("Listener resolve error:", err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Listener error:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		msg := string(buffer[:n])
		if strings.HasPrefix(msg, MagicString) {
			data := strings.TrimPrefix(msg, MagicString)
			parts := strings.SplitN(data, ":", 4)
			if len(parts) < 2 {
				continue
			}

			serverPort := parts[0]
			peerName := parts[1]
			peerUserID := ""
			if len(parts) >= 3 {
				peerUserID = parts[2]
			}
			peerGUIPort := ""
			if len(parts) == 4 {
				peerGUIPort = parts[3]
			}

			// Ignore our own broadcast
			if ":"+serverPort == ignorePort || serverPort == ignorePort {
				continue
			}

			serverAddress := fmt.Sprintf("%s:%s", src.IP.String(), serverPort)

			peerMutex.Lock()
			ActivePeers[serverAddress] = Peer{
				Name:     peerName,
				Address:  serverAddress,
				UserID:   peerUserID,
				GUIPort:  peerGUIPort,
				LastSeen: time.Now(),
			}
			peerMutex.Unlock()
		}
	}
}
