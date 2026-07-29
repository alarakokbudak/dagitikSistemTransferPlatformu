package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/auth"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/discovery"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/network"
)

type APIServer struct {
	Port         string
	MyP2PPort    string
	SyncManager  *network.SyncManager
	FrontendDir  string
	AccountStore *auth.AccountStore
}

func NewAPIServer(port string, myP2PPort string, sm *network.SyncManager, frontendDir string, accountStore *auth.AccountStore) *APIServer {
	return &APIServer{
		Port:         port,
		MyP2PPort:    myP2PPort,
		SyncManager:  sm,
		FrontendDir:  frontendDir,
		AccountStore: accountStore,
	}
}

func (s *APIServer) Start() error {
	// Serve static frontend files
	fs := http.FileServer(http.Dir(s.FrontendDir))
	http.Handle("/", fs)

	// API Endpoints - Auth
	http.HandleFunc("/api/register", s.handleRegister)
	http.HandleFunc("/api/login", s.handleLogin)
	http.HandleFunc("/api/logout", s.handleLogout)
	http.HandleFunc("/api/profile", s.handleProfile)
	http.HandleFunc("/api/profile/avatar", s.handleUpdateAvatar)

	// API Endpoints - Core
	http.HandleFunc("/api/status", s.handleStatus)
	http.HandleFunc("/api/sync", s.handleSync)
	http.HandleFunc("/api/upload", s.handleUpload)
	http.HandleFunc("/api/peers", s.handlePeers)
	http.HandleFunc("/api/peers/alias", s.handleSetPeerAlias)
	http.HandleFunc("/api/files", s.handleFiles)
	http.HandleFunc("/api/download", s.handleDownload)

	// API Endpoints - New
	http.HandleFunc("/api/pair", s.handlePair)
	http.HandleFunc("/api/history", s.handleHistory)

	// API Endpoints - Requests
	http.HandleFunc("/api/pair-request", s.handlePairRequest)
	http.HandleFunc("/api/pair-requests", s.handlePairRequestsGet)
	http.HandleFunc("/api/pair-accept", s.handlePairAccept)
	http.HandleFunc("/api/pair-reject", s.handlePairReject)
	http.HandleFunc("/api/pair-callback", s.handlePairCallback)
	http.HandleFunc("/api/transfer-request", s.handleTransferRequest)
	http.HandleFunc("/api/transfer-requests", s.handleTransferRequestsGet)
	http.HandleFunc("/api/transfer-accept", s.handleTransferAccept)
	http.HandleFunc("/api/transfer-reject", s.handleTransferReject)
	http.HandleFunc("/api/transfer-callback", s.handleTransferCallback)

	fmt.Printf("Web GUI is running! Open your browser and navigate to: http://localhost%s\n", s.Port)
	return http.ListenAndServe(s.Port, nil)
}

// ─── Auth Handlers ───

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "E-posta ve şifre gerekli", http.StatusBadRequest)
		return
	}

	account, err := s.AccountStore.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	discovery.SetLocalIdentity(account.ID, account.DisplayName)

	// Şifre hash'ini response'a dahil etme
	type safeAccount struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		AvatarColor string `json:"avatarColor"`
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeAccount{
		ID: account.ID, Email: account.Email,
		DisplayName: account.DisplayName, AvatarColor: account.AvatarColor,
	})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Email == "" || req.Password == "" {
		http.Error(w, "E-posta ve şifre gerekli", http.StatusBadRequest)
		return
	}

	account, err := s.AccountStore.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	discovery.SetLocalIdentity(account.ID, account.DisplayName)

	type safeAccount struct {
		ID          string `json:"id"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		AvatarColor string `json:"avatarColor"`
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeAccount{
		ID: account.ID, Email: account.Email,
		DisplayName: account.DisplayName, AvatarColor: account.AvatarColor,
	})
}

func (s *APIServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	s.AccountStore.Logout()
	discovery.SetLocalIdentity("", "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Çıkış yapıldı"})
}

func (s *APIServer) handleProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		account := s.AccountStore.GetCurrent()
		if account == nil {
			http.Error(w, "Giriş yapılmamış", http.StatusUnauthorized)
			return
		}
		type safeAccount struct {
			ID           string            `json:"id"`
			Email        string            `json:"email"`
			DisplayName  string            `json:"displayName"`
			AvatarColor  string            `json:"avatarColor"`
			AvatarBase64 string            `json:"avatarBase64"`
			PairedPeers  []string          `json:"pairedPeers"`
			PeerAliases  map[string]string `json:"peerAliases"`
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(safeAccount{
			ID: account.ID, Email: account.Email,
			DisplayName: account.DisplayName, AvatarColor: account.AvatarColor,
			AvatarBase64: account.AvatarBase64, PairedPeers: account.PairedPeers,
			PeerAliases: account.PeerAliases,
		})

	case http.MethodPut:
		var req struct {
			DisplayName string `json:"displayName"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		err := s.AccountStore.UpdateProfile(req.DisplayName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		account := s.AccountStore.GetCurrent()
		discovery.SetLocalIdentity(account.ID, account.DisplayName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(account)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *APIServer) handleUpdateAvatar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		AvatarBase64 string `json:"avatarBase64"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := s.AccountStore.SetAvatar(req.AvatarBase64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Avatar güncellendi"})
}

// ─── Pairing Handler ───

func (s *APIServer) handlePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PeerID string `json:"peerId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.PeerID == "" {
		http.Error(w, "peerId gerekli", http.StatusBadRequest)
		return
	}

	// Peer'in ağda olup olmadığını kontrol et
	peers := discovery.GetActivePeers()
	found := false
	var foundPeer discovery.Peer
	for _, p := range peers {
		if p.UserID == req.PeerID {
			found = true
			foundPeer = p
			break
		}
	}

	if !found {
		http.Error(w, "Bu ID ile ağda aktif cihaz bulunamadı. Cihazın çalışır durumda olduğundan emin olun.", http.StatusNotFound)
		return
	}

	// Send PairRequest to target peer
	acc := s.AccountStore.GetCurrent()
	if acc == nil {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	myPort := s.Port
	if strings.Contains(myPort, ":") {
		parts := strings.Split(myPort, ":")
		myPort = parts[len(parts)-1]
	}

	targetURL := fmt.Sprintf("http://%s:%s/api/pair-request", strings.Split(foundPeer.Address, ":")[0], foundPeer.GUIPort)
	
	payload := auth.PairRequest{
		FromID:   acc.ID,
		FromName: acc.DisplayName,
		PeerIP:   "127.0.0.1",
		PeerPort: myPort,
	}
	if payload.FromName == "" {
		payload.FromName = acc.ID
	}

	jsonData, _ := json.Marshal(payload)
	_, err := http.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Karşı cihaza ulaşılamadı", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "Eşleştirme isteği gönderildi. Karşı tarafın onayı bekleniyor.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ─── History Handler ───

func (s *APIServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	history := s.AccountStore.GetHistory()
	
	// Override names with aliases if available
	for i, record := range history {
		alias := s.AccountStore.GetPeerAlias(record.PeerID)
		if alias != "" {
			history[i].PeerName = alias
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// ─── Core Handlers ───

func (s *APIServer) handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "File parameter is missing", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(s.SyncManager.DirPath, filename)
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.SyncManager.DirPath)) {
		http.Error(w, "Invalid file path", http.StatusForbidden)
		return
	}

	// Provide proper quotes for filename to handle spaces
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	http.ServeFile(w, r, cleanPath)
}

type FileInfo struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	ModTime  string `json:"modTime"`
	PeerName string `json:"peerName,omitempty"`
}

func (s *APIServer) handleFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	dir := s.SyncManager.DirPath
	entries, err := os.ReadDir(dir)
	if err != nil {
		http.Error(w, "Error reading directory", http.StatusInternalServerError)
		return
	}

	history := s.AccountStore.GetHistory()
	fileToPeer := make(map[string]string)
	for _, h := range history {
		if _, exists := fileToPeer[h.Filename]; !exists {
			alias := s.AccountStore.GetPeerAlias(h.PeerID)
			pName := h.PeerName
			if alias != "" {
				pName = alias
			} else if pName == "" {
				pName = h.PeerID
			}
			fileToPeer[h.Filename] = pName
		}
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err == nil {
			if peerName, exists := fileToPeer[info.Name()]; exists {
				files = append(files, FileInfo{
					Name:     info.Name(),
					Size:     info.Size(),
					ModTime:  info.ModTime().Format("2006-01-02 15:04:05"),
					PeerName: peerName,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *APIServer) handlePeers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
		return
	}

	allPeers := discovery.GetActivePeers()

	// Sadece eşleşmiş peer'ları filtrele
	showAll := r.URL.Query().Get("all") == "true"
	var result []discovery.Peer

	if showAll {
		result = allPeers
	} else {
		for _, p := range allPeers {
			if s.AccountStore.IsPaired(p.UserID) {
				alias := s.AccountStore.GetPeerAlias(p.UserID)
				if alias != "" {
					p.Name = alias
				}
				result = append(result, p)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *APIServer) handleSetPeerAlias(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		PeerID string `json:"peerId"`
		Alias  string `json:"alias"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.PeerID == "" {
		http.Error(w, "peerId gerekli", http.StatusBadRequest)
		return
	}

	err := s.AccountStore.SetPeerAlias(req.PeerID, req.Alias)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Takma ad güncellendi"})
}

func (s *APIServer) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(50 << 20)

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	destPath := s.SyncManager.DirPath + "/" + handler.Filename
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Error saving the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)

	response := map[string]string{
		"message":  "File uploaded successfully",
		"filename": handler.Filename,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	peers := discovery.GetActivePeers()

	account := s.AccountStore.GetCurrent()
	name := "Unknown"
	userID := ""
	if account != nil {
		if account.DisplayName != "" {
			name = account.DisplayName
		} else {
			name = account.ID
		}
		userID = account.ID
	}

	status := map[string]interface{}{
		"status":      "online",
		"port":        s.Port,
		"p2pPort":     s.MyP2PPort,
		"name":        name,
		"userId":      userID,
		"activePeers": len(peers),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

type SyncRequest struct {
	Filename   string `json:"filename"`
	TargetPeer string `json:"targetPeer"`
}

func (s *APIServer) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SyncRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.Filename == "" || req.TargetPeer == "" {
		http.Error(w, "Invalid request. Need filename and targetPeer.", http.StatusBadRequest)
		return
	}

	targetFilename := req.Filename
	targetAddr := req.TargetPeer

	// Find target peer to get GUI port
	peers := discovery.GetActivePeers()
	var foundPeer discovery.Peer
	for _, p := range peers {
		if p.Address == targetAddr {
			foundPeer = p
			break
		}
	}

	if foundPeer.GUIPort == "" {
		http.Error(w, "Target peer GUI port unknown", http.StatusInternalServerError)
		return
	}

	acc := s.AccountStore.GetCurrent()
	if acc == nil {
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	myPort := s.Port
	if strings.Contains(myPort, ":") {
		parts := strings.Split(myPort, ":")
		myPort = parts[len(parts)-1]
	}

	// Send Transfer Request
	targetURL := fmt.Sprintf("http://%s:%s/api/transfer-request", strings.Split(foundPeer.Address, ":")[0], foundPeer.GUIPort)
	
	payload := auth.TransferRequest{
		FromID:   acc.ID,
		FromName: acc.DisplayName,
		Filename: targetFilename,
		Size:     0, // TODO: Get actual file size
		PeerIP:   "127.0.0.1",
		PeerPort: myPort,
	}
	if payload.FromName == "" {
		payload.FromName = acc.ID
	}

	jsonData, _ := json.Marshal(payload)
	_, err = http.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to send transfer request", http.StatusInternalServerError)
		return
	}

	// Transfer geçmişine ekle
	s.AccountStore.AddTransfer(auth.TransferRecord{
		Filename:  targetFilename,
		Direction: "sent",
		PeerID:    foundPeer.UserID,
		PeerName:  foundPeer.Name,
	})

	response := map[string]string{
		"message": "Transfer isteği gönderildi. Karşı tarafın onayı bekleniyor.",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
