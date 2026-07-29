# 🗺️ Yeni Harita Kullanım Kılavuzu

> **Proje:** Teknofest Robotaksi Otonom Araç Simülasyonu  
> **Aktif Harita Modeli:** `yeni_harita` (Blender GLB, köşeli versiyon)  
> **Son Güncelleme:** 19 Haziran 2026

---

## 🚀 Hızlı Başlatma (3 Adım)

```bash
# 1. Proje dizinine git
cd /home/emre/Documents/GitHub/Tekonofest_OtonomveKaraVerme

# 2. Build et ve source al
colcon build --packages-select robotaksi_sim robotaksi_autonomy
source install/setup.bash

# 3. Simülasyonu yeni harita ile başlat
ros2 launch robotaksi_autonomy robotaksi_full.launch.py \
    world:=$(pwd)/src/robotaksi_sim/worlds/yeni_pist.world
```

> ⚠️ `world:=` parametresinde **mutlaka tam yol (absolute path)** kullan! `$(pwd)` bunu otomatik yapar.

---

## 📁 Dosya Konumları

| Dosya | Yol |
|---|---|
| **Harita modeli (SDF)** | `src/robotaksi_sim/models/yeni_harita/model.sdf` |
| **Harita mesh (GLB)** | `src/robotaksi_sim/models/yeni_harita/meshes/map_tunelsiz.glb` |
| **World dosyası** | `src/robotaksi_sim/worlds/yeni_pist.world` |
| **Launch dosyası** | `src/robotaksi_autonomy/launch/robotaksi_full.launch.py` |
| **Model config** | `src/robotaksi_sim/models/yeni_harita/model.config` |

---

## 🔧 Harita Ayarları (model.sdf)

Harita boyutu veya yönü yanlışsa `model.sdf` dosyasını düzenle:

### Ölçek (Scale) Değiştirme
```xml
<scale>1 1 1</scale>    <!-- Gerçek boyut: 123m x 81m -->
<scale>0.5 0.5 0.5</scale>  <!-- Yarı boyut: ~62m x 40m -->
<scale>0.1 0.1 0.1</scale>  <!-- Küçük: ~12m x 8m -->
```

### Yön (Rotation) Düzeltme
```xml
<!-- pose = "x y z roll pitch yaw" (radyan cinsinden) -->
<pose>0 0 0.05 -1.5708 0 0</pose>  <!-- Düz yatırılmış (doğru) -->
<pose>0 0 0.05 0 0 0</pose>         <!-- Dik duruyor (yanlış) -->
```

### Değişiklik Sonrası
```bash
colcon build --packages-select robotaksi_sim
source install/setup.bash
# Simülasyonu yeniden başlat
```

---

## 🆕 Harita Güncelleme (Blender'dan Yeni Export)

1. Blender'da modeli düzenle
2. **File → Export → glTF 2.0 (.glb)** olarak export et
3. Dosyayı kopyala:
```bash
cp "/home/emre/Documents/Yeniharita/map(tunelsiz).glb" \
   src/robotaksi_sim/models/yeni_harita/meshes/map_tunelsiz.glb
```
4. Build et: `colcon build --packages-select robotaksi_sim`
5. Source et: `source install/setup.bash`
6. Simülasyonu başlat

> ⚠️ Dosya adlarında **parantez `()`, boşluk, Türkçe karakter** kullanma!

---

## ❌ Sorun Yaşıyorsan

Detaylı sorun giderme rehberi için: **[gazebo_harita_rehberi.md](gazebo_harita_rehberi.md)**

Bu rehberde şu sorunların çözümleri var:
1. Harita dik duruyor (90° dönmüş)
2. Harita çok büyük / çok küçük
3. PC / Gazebo çöküyor (crash)
4. Harita zemin altında kalıyor
5. Harita görünmüyor (boş sahne)
6. GPU hatası / siyah ekran
7. Araç haritanın içine düşüyor
8. Tekstürler/renkler görünmüyor

---

## 🛑 Simülasyonu Durdurma

```bash
# Tüm ROS 2 ve Gazebo süreçlerini kapat
pkill -f "gz sim"
pkill -f "ros2"
```

---

## 📊 Referans Bilgiler

| Parametre | Değer |
|---|---|
| Araç uzunluğu | ~2.0 m |
| Araç genişliği | ~1.1 m (tekerlekler arası) |
| Tekerlek yarıçapı | 0.207 m |
| Harita boyutu (scale=1) | ~123m x 81m |
| Harita format | GLB (Binary glTF 2.0) |
| Simülasyon motoru | Gazebo Harmonic |
| ROS 2 sürümü | Jazzy |

---

*Detaylı teknik rehber ve sorun giderme: [gazebo_harita_rehberi.md](gazebo_harita_rehberi.md)*
