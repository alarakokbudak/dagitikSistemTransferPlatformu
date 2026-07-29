# Dağıtık Sistem Transfer Platformu

Bu proje, uçtan uca şifreleme (E2EE) ile desteklenen, eşler arası (P2P) dosya transferi ve senkronizasyonu sağlayan bir platformdur. Go dili ile geliştirilmiştir ve hem yerel ağlarda (LAN) hem de geniş ağlarda (WAN) çalışabilir.

## Özellikler

- **Eşler Arası (P2P) Senkronizasyon:** Merkezi bir sunucuya ihtiyaç duymadan cihazlar arasında doğrudan dosya paylaşımı.
- **Uçtan Uca Şifreleme (E2EE):** Dosyalarınız PBKDF2 tabanlı güçlü bir anahtar türetme fonksiyonu kullanılarak korunur.
- **Kullanıcı Dostu Web GUI:** Transfer işlemlerini yönetmek için entegre edilmiş web arayüzü.
- **LAN ve WAN Desteği:** Yerel ağda mDNS ile otomatik eş bulma veya Tracker (izleyici) sunucusu ile WAN üzerinden cihazları eşleştirme.
- **Çoklu Hesap Desteği:** Farklı portlar üzerinden çalışan uygulamalar için kendi hesap/kimlik deposunu (accounts) yönetir.

## Kurulum ve Çalıştırma

Projeyi derlemek ve çalıştırmak için sisteminizde Go'nun yüklü olması gerekmektedir.

```bash
# Projeyi bilgisayarınıza kopyalayın
git clone https://github.com/alarakokbudak/dagitikSistemTransferPlatformu.git
cd dagitikSistemTransferPlatformu

# Bağımlılıkları yükleyin
go mod tidy

# Projeyi varsayılan ayarlarla başlatın
go run main.go
```

## Komut Satırı Parametreleri

Projeyi çalıştırırken belirli yapılandırmaları değiştirmek için argümanlar kullanabilirsiniz:

- `-dir`: Senkronize edilecek klasör yolu (Varsayılan: `./sunucu_dosyalari`)
- `-gui`: Web GUI arayüzünün yayınlanacağı port (Varsayılan: `:9090`)
- `-p2p`: P2P Sunucusunun dinleyeceği port (Varsayılan: `:8080`)
- `-tracker`: WAN modu için merkezi Tracker sunucu adresi (Örn: `http://localhost:8000`)
- `-secret`: Şifreleme (E2EE) için özel gizli anahtar (Varsayılan: `my-super-secret-key`)

### Örnek Kullanım (LAN Modu)
Aynı ağdaki iki bilgisayar arasında (veya aynı bilgisayarda 2 farklı terminal üzerinden) çalıştırabilirsiniz:

**İstemci 1:**
```bash
go run main.go -dir="./sunucu_dosyalari" -gui=":9090" -p2p=":8080"
```

**İstemci 2:**
```bash
go run main.go -dir="./istemci_dosyalari" -gui=":9091" -p2p=":8081"
```
