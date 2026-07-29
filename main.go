package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/api"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/auth"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/crypto"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/discovery"
	"github.com/alarakokbudak/dagitikSistemTransferPlatformu/network"
)

func main() {
	dir := flag.String("dir", "./sunucu_dosyalari", "directory to sync")
	guiPort := flag.String("gui", ":9090", "port for Web GUI")
	p2pPort := flag.String("p2p", ":8080", "port for P2P Sync Server")
	trackerURL := flag.String("tracker", "", "URL of the central tracker server (e.g., http://localhost:8000)")
	secret := flag.String("secret", "my-super-secret-key", "secret key for E2EE")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Println("Error getting absolute path:", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		err = os.MkdirAll(absDir, 0755)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			os.Exit(1)
		}
	}

	// PBKDF2 ile Güçlü Anahtar Türetme (KDF)
	salt := []byte("MerkleSync-Secure-Salt-2026")
	key := crypto.PBKDF2([]byte(*secret), salt, 100000, 32)

	syncManager := network.NewSyncManager(absDir, key)

	// Hesap deposunu yükle
	accountsFile := filepath.Join(absDir, "..", "accounts_"+(*guiPort)[1:]+".json")
	accountStore := auth.NewAccountStore(accountsFile)

	// Eğer zaten hesap varsa identity'i ayarla
	if account := accountStore.GetCurrent(); account != nil {
		discovery.SetLocalIdentity(account.ID, account.DisplayName)
		fmt.Printf("Mevcut hesap yüklendi: %s (%s)\n", account.ID, account.DisplayName)
	}

	if *trackerURL != "" {
		fmt.Println("WAN Mode Enabled: Using Tracker Server at", *trackerURL)
		go discovery.StartTrackerPolling(*trackerURL, *p2pPort, *guiPort)
	} else {
		fmt.Println("LAN Mode Enabled: Using mDNS for local discovery")
		// Arka planda mDNS Yayını başlat
		go discovery.StartBroadcasting(*p2pPort, "Peer", *guiPort)

		// Arka planda mDNS Dinlemeyi başlat
		go discovery.StartListening(*p2pPort)
	}

	// Arka planda P2P Sunucusunu başlat
	go func() {
		fmt.Printf("Background P2P Server listening on %s\n", *p2pPort)
		syncManager.StartServer(*p2pPort)
	}()

	// Ön planda Web GUI Sunucusunu başlat
	frontendDir, _ := filepath.Abs("./frontend")
	apiServer := api.NewAPIServer(*guiPort, *p2pPort, syncManager, frontendDir, accountStore)

	err = apiServer.Start()
	if err != nil {
		fmt.Println("API Server Error:", err)
	}
}
