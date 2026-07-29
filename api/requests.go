package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/auth"
)

// ---- Pair Requests ----

func (s *APIServer) handlePairRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}
	var req auth.PairRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	req.ID = fmt.Sprintf("pr_%d", time.Now().UnixNano())
	req.Timestamp = time.Now().Format("15:04")
	
	s.AccountStore.AddPairRequest(req)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handlePairRequestsGet(w http.ResponseWriter, r *http.Request) {
	reqs := s.AccountStore.GetPairRequests()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqs)
}

func (s *APIServer) handlePairAccept(w http.ResponseWriter, r *http.Request) {
	var req struct { ReqID string `json:"reqId"` }
	json.NewDecoder(r.Body).Decode(&req)

	reqs := s.AccountStore.GetPairRequests()
	var targetReq *auth.PairRequest
	for _, pr := range reqs {
		if pr.ID == req.ReqID {
			targetReq = &pr
			break
		}
	}

	if targetReq != nil {
		s.AccountStore.AddPairedPeer(targetReq.FromID)
		s.AccountStore.RemovePairRequest(targetReq.ID)

		acc := s.AccountStore.GetCurrent()
		if acc != nil {
			go func() {
				url := fmt.Sprintf("http://%s:%s/api/pair-callback", targetReq.PeerIP, targetReq.PeerPort)
				payload := map[string]string{"peerId": acc.ID}
				jsonData, _ := json.Marshal(payload)
				http.Post(url, "application/json", bytes.NewBuffer(jsonData))
			}()
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handlePairReject(w http.ResponseWriter, r *http.Request) {
	var req struct { ReqID string `json:"reqId"` }
	json.NewDecoder(r.Body).Decode(&req)
	s.AccountStore.RemovePairRequest(req.ReqID)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handlePairCallback(w http.ResponseWriter, r *http.Request) {
	var req struct { PeerID string `json:"peerId"` }
	json.NewDecoder(r.Body).Decode(&req)
	s.AccountStore.AddPairedPeer(req.PeerID)
	w.WriteHeader(http.StatusOK)
}

// ---- Transfer Requests ----

func (s *APIServer) handleTransferRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}
	var req auth.TransferRequest
	json.NewDecoder(r.Body).Decode(&req)
	
	req.ID = fmt.Sprintf("tr_%d", time.Now().UnixNano())
	req.Timestamp = time.Now().Format("15:04")
	
	s.AccountStore.AddTransferRequest(req)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handleTransferRequestsGet(w http.ResponseWriter, r *http.Request) {
	reqs := s.AccountStore.GetTransferRequests()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqs)
}

func (s *APIServer) handleTransferAccept(w http.ResponseWriter, r *http.Request) {
	var req struct { ReqID string `json:"reqId"` }
	json.NewDecoder(r.Body).Decode(&req)

	reqs := s.AccountStore.GetTransferRequests()
	var targetReq *auth.TransferRequest
	for _, tr := range reqs {
		if tr.ID == req.ReqID {
			targetReq = &tr
			break
		}
	}

	if targetReq != nil {
		s.AccountStore.RemoveTransferRequest(targetReq.ID)
		
		myP2PPort := s.MyP2PPort
		if strings.Contains(myP2PPort, ":") {
			parts := strings.Split(myP2PPort, ":")
			myP2PPort = parts[len(parts)-1]
		}
		
		// Find my IP (hacky but for localhost it works, usually we extract from request)
		myIP := "127.0.0.1"
		host := r.Host
		if strings.Contains(host, ":") {
			myIP = strings.Split(host, ":")[0]
		}

		go func() {
			url := fmt.Sprintf("http://%s:%s/api/transfer-callback", targetReq.PeerIP, targetReq.PeerPort)
			payload := map[string]string{
				"filename": targetReq.Filename,
				"peerAddr": fmt.Sprintf("%s:%s", myIP, myP2PPort),
			}
			jsonData, _ := json.Marshal(payload)
			http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		}()

		// Transfer geçmişine ekle (Alıcı)
		s.AccountStore.AddTransfer(auth.TransferRecord{
			Filename:  targetReq.Filename,
			Direction: "received",
			PeerID:    targetReq.FromID,
			PeerName:  targetReq.FromName,
		})
	}
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handleTransferReject(w http.ResponseWriter, r *http.Request) {
	var req struct { ReqID string `json:"reqId"` }
	json.NewDecoder(r.Body).Decode(&req)
	s.AccountStore.RemoveTransferRequest(req.ReqID)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) handleTransferCallback(w http.ResponseWriter, r *http.Request) {
	var req struct { 
		Filename string `json:"filename"`
		PeerAddr string `json:"peerAddr"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	go func() {
		fmt.Printf("[GUI Command] Request accepted. Syncing %s to %s.\n", req.Filename, req.PeerAddr)
		err := s.SyncManager.ConnectAndSync(req.PeerAddr, req.Filename)
		if err != nil {
			fmt.Println("[GUI Command] Sync failed:", err)
		} else {
			fmt.Println("[GUI Command] Sync completed successfully!")
		}
	}()
	w.WriteHeader(http.StatusOK)
}
