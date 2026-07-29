package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
)

type Account struct {
	ID           string           `json:"id"`
	Email        string           `json:"email"`
	PasswordHash string           `json:"passwordHash"`
	DisplayName  string           `json:"displayName"`
	AvatarColor  string           `json:"avatarColor"`
	AvatarBase64 string           `json:"avatarBase64"`
	CreatedAt    time.Time        `json:"createdAt"`
	PairedPeers             []string          `json:"pairedPeers"`
	PeerAliases             map[string]string `json:"peerAliases"` // peerId -> alias
	History                 []TransferRecord  `json:"history"`
	PendingPairRequests     []PairRequest     `json:"pendingPairRequests"`
	PendingTransferRequests []TransferRequest `json:"pendingTransferRequests"`
}

type PairRequest struct {
	ID        string `json:"id"` // Benzersiz istek ID'si
	FromID    string `json:"fromId"`
	FromName  string `json:"fromName"`
	Timestamp string `json:"timestamp"`
	PeerIP    string `json:"peerIp"`
	PeerPort  string `json:"peerPort"`
}

type TransferRequest struct {
	ID        string `json:"id"`
	FromID    string `json:"fromId"`
	FromName  string `json:"fromName"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	Timestamp string `json:"timestamp"`
	PeerIP    string `json:"peerIp"`
	PeerPort  string `json:"peerPort"`
}

type TransferRecord struct {
	Filename  string `json:"filename"`
	Direction string `json:"direction"`
	PeerID    string `json:"peerId"`
	PeerName  string `json:"peerName"`
	Size      int64  `json:"size"`
	Timestamp string `json:"timestamp"`
}

type AccountStore struct {
	mu          sync.RWMutex
	filePath    string
	Accounts    map[string]*Account `json:"accounts"` // email -> Account
	CurrentUser string              `json:"-"`         // Şu an giriş yapmış kullanıcının email'i (kalıcı değil)
}

var avatarColors = []string{
	"#818cf8", "#a78bfa", "#c084fc", "#f472b6", "#fb7185",
	"#f97316", "#fbbf24", "#34d399", "#2dd4bf", "#38bdf8",
}

func NewAccountStore(filePath string) *AccountStore {
	store := &AccountStore{
		filePath: filePath,
		Accounts: make(map[string]*Account),
	}
	store.load()
	return store
}

func generateID() string {
	part1, _ := rand.Int(rand.Reader, big.NewInt(10000))
	part2, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("#%04d-%04d", part1.Int64(), part2.Int64())
}

func randomColor() string {
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(avatarColors))))
	return avatarColors[idx.Int64()]
}

func hashPassword(password string) string {
	salted := "SYNC_P2P_SALT_2026:" + password
	hash := sha256.Sum256([]byte(salted))
	return hex.EncodeToString(hash[:])
}

func (s *AccountStore) Register(email, password string) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// E-posta zaten kayıtlı mı?
	if _, exists := s.Accounts[email]; exists {
		return nil, fmt.Errorf("bu e-posta adresi zaten kayıtlı")
	}

	account := &Account{
		ID:           generateID(),
		Email:        email,
		PasswordHash: hashPassword(password),
		DisplayName:             "",
		AvatarColor:             randomColor(),
		AvatarBase64:            "",
		CreatedAt:               time.Now(),
		PairedPeers:             []string{},
		PeerAliases:             make(map[string]string),
		History:                 []TransferRecord{},
		PendingPairRequests:     []PairRequest{},
		PendingTransferRequests: []TransferRequest{},
	}

	s.Accounts[email] = account
	s.CurrentUser = email
	s.save()
	return account, nil
}

func (s *AccountStore) Login(email, password string) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	account, exists := s.Accounts[email]
	if !exists {
		return nil, fmt.Errorf("hesap bulunamadı")
	}

	if account.PasswordHash != hashPassword(password) {
		return nil, fmt.Errorf("şifre yanlış")
	}

	s.CurrentUser = email
	return account, nil
}

func (s *AccountStore) Logout() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CurrentUser = ""
}

func (s *AccountStore) GetCurrent() *Account {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.CurrentUser == "" {
		return nil
	}
	return s.Accounts[s.CurrentUser]
}

func (s *AccountStore) UpdateProfile(displayName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentUser == "" {
		return fmt.Errorf("giriş yapılmamış")
	}

	account := s.Accounts[s.CurrentUser]
	if account == nil {
		return fmt.Errorf("hesap bulunamadı")
	}

	account.DisplayName = displayName
	s.save()
	return nil
}

func (s *AccountStore) SetAvatar(base64Data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return fmt.Errorf("giriş yapılmamış") }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return fmt.Errorf("hesap bulunamadı") }
	acc.AvatarBase64 = base64Data
	s.save()
	return nil
}

func (s *AccountStore) SetPeerAlias(peerID, alias string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return fmt.Errorf("giriş yapılmamış") }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return fmt.Errorf("hesap bulunamadı") }
	
	if acc.PeerAliases == nil {
		acc.PeerAliases = make(map[string]string)
	}
	
	if alias == "" {
		delete(acc.PeerAliases, peerID)
	} else {
		acc.PeerAliases[peerID] = alias
	}
	s.save()
	return nil
}

func (s *AccountStore) GetPeerAlias(peerID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.CurrentUser == "" { return "" }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return "" }
	if acc.PeerAliases == nil { return "" }
	return acc.PeerAliases[peerID]
}

func (s *AccountStore) AddPairedPeer(peerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentUser == "" {
		return
	}
	account := s.Accounts[s.CurrentUser]
	if account == nil {
		return
	}

	for _, id := range account.PairedPeers {
		if id == peerID {
			return
		}
	}

	account.PairedPeers = append(account.PairedPeers, peerID)
	s.save()
}

func (s *AccountStore) IsPaired(peerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.CurrentUser == "" {
		return false
	}
	account := s.Accounts[s.CurrentUser]
	if account == nil {
		return false
	}

	for _, id := range account.PairedPeers {
		if id == peerID {
			return true
		}
	}
	return false
}

func (s *AccountStore) AddTransfer(record TransferRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.CurrentUser == "" {
		return
	}
	account := s.Accounts[s.CurrentUser]
	if account == nil {
		return
	}

	record.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	account.History = append([]TransferRecord{record}, account.History...)

	if len(account.History) > 50 {
		account.History = account.History[:50]
	}
	s.save()
}

func (s *AccountStore) GetHistory() []TransferRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.CurrentUser == "" {
		return []TransferRecord{}
	}
	account := s.Accounts[s.CurrentUser]
	if account == nil {
		return []TransferRecord{}
	}
	
	// Clean up any old buggy records that saved IP instead of ID
	var cleanHistory []TransferRecord
	for _, rec := range account.History {
		if !strings.Contains(rec.PeerID, "127.0.0.1") {
			cleanHistory = append(cleanHistory, rec)
		}
	}
	
	return cleanHistory
}

func (s *AccountStore) save() {
	data, err := json.MarshalIndent(s.Accounts, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(s.filePath, data, 0644)
}

func (s *AccountStore) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return
	}

	var accounts map[string]*Account
	if err := json.Unmarshal(data, &accounts); err == nil {
		s.Accounts = accounts
	}
}

// ---- Pending Request Functions ----

func (s *AccountStore) AddPairRequest(req PairRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return }
	
	// Check if already requested
	for _, r := range acc.PendingPairRequests {
		if r.FromID == req.FromID {
			return
		}
	}
	acc.PendingPairRequests = append(acc.PendingPairRequests, req)
	s.save()
}

func (s *AccountStore) GetPairRequests() []PairRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.CurrentUser == "" { return nil }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return nil }
	return acc.PendingPairRequests
}

func (s *AccountStore) RemovePairRequest(reqID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return }

	var filtered []PairRequest
	for _, r := range acc.PendingPairRequests {
		if r.ID != reqID {
			filtered = append(filtered, r)
		}
	}
	acc.PendingPairRequests = filtered
	s.save()
}

func (s *AccountStore) AddTransferRequest(req TransferRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return }

	acc.PendingTransferRequests = append(acc.PendingTransferRequests, req)
	s.save()
}

func (s *AccountStore) GetTransferRequests() []TransferRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.CurrentUser == "" { return nil }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return nil }
	return acc.PendingTransferRequests
}

func (s *AccountStore) RemoveTransferRequest(reqID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.CurrentUser == "" { return }
	acc := s.Accounts[s.CurrentUser]
	if acc == nil { return }

	var filtered []TransferRequest
	for _, r := range acc.PendingTransferRequests {
		if r.ID != reqID {
			filtered = append(filtered, r)
		}
	}
	acc.PendingTransferRequests = filtered
	s.save()
}
